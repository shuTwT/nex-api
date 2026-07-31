package runtime

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := sanitizeRequestID(r.Header.Get(RequestIDHeader))
		ctx := WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

func RequestIDHandler(next http.Handler) http.Handler { return RequestIDMiddleware(next) }

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	logger = loggerOrDiscard(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "http panic recovered",
						slog.Any("panic", recovered),
						slog.String("stack", string(debug.Stack())),
					)
					if writeErr := WriteError(w, r, NewAPIError(http.StatusInternalServerError, "internal_error", "internal server error", nil)); writeErr != nil {
						logger.ErrorContext(r.Context(), "write panic response failed", slog.Any("err", writeErr))
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	logger = loggerOrDiscard(logger)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)
			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			logger.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("bytes", wrapped.BytesWritten()),
				slog.Duration("duration", time.Since(started)),
			)
		})
	}
}

func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			if maxBytes > 0 && r.ContentLength > maxBytes {
				if err := WriteError(w, r, NewAPIError(http.StatusRequestEntityTooLarge, "request_too_large", "request body is too large", nil)); err != nil {
					return
				}
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func DevelopmentCORS(origins []string) func(http.Handler) http.Handler {
	allowed := append([]string(nil), origins...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && originAllowed(origin, allowed) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			}
			if r.Method == http.MethodOptions && originAllowed(origin, allowed) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type requestContextHandler struct{ slog.Handler }

func (h requestContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if requestID := RequestID(ctx); requestID != "" {
		record.AddAttrs(slog.String("request_id", requestID))
	}
	return h.Handler.Handle(ctx, record)
}

func (h requestContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return requestContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h requestContextHandler) WithGroup(name string) slog.Handler {
	return requestContextHandler{Handler: h.Handler.WithGroup(name)}
}

func loggerOrDiscard(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func sanitizeRequestID(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && len(value) <= 128 {
		valid := true
		for _, char := range value {
			if char < 0x20 || char == 0x7f {
				valid = false
				break
			}
		}
		if valid {
			return value
		}
	}
	return uuid.NewString()
}

func originAllowed(origin string, allowed []string) bool {
	for _, candidate := range allowed {
		if candidate == "*" || candidate == origin {
			return true
		}
	}
	return false
}
