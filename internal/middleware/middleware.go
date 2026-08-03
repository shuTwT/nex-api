// Package middleware holds the HTTP request middleware: request ID, panic
// recovery, access logging, body limits and development CORS. It depends only
// on service and infra packages; error responses are written inline so the
// package never depends on the handler layer.
package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/shuTwT/nex-api/internal/infra/logger"
)

// RequestIDHeader is the HTTP correlation header exposed to handlers without
// requiring them to depend on the logger infrastructure package.
const RequestIDHeader = logger.RequestIDHeader

// RequestID returns the request correlation value installed by
// RequestIDMiddleware.
func RequestID(ctx context.Context) string { return logger.RequestID(ctx) }

// RequestIDMiddleware assigns (or sanitizes) an X-Request-ID header and
// stores it in the request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := sanitizeRequestID(r.Header.Get(logger.RequestIDHeader))
		ctx := logger.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		w.Header().Set(logger.RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

// RequestIDHandler is an alias for RequestIDMiddleware.
func RequestIDHandler(next http.Handler) http.Handler { return RequestIDMiddleware(next) }

// Recovery converts panics into a 500 error envelope and logs the stack.
func Recovery(loggerInstance *slog.Logger) func(http.Handler) http.Handler {
	loggerInstance = loggerOrDiscard(loggerInstance)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					loggerInstance.ErrorContext(r.Context(), "http panic recovered",
						slog.Any("panic", recovered),
						slog.String("stack", string(debug.Stack())),
					)
					writeErrorResponse(w, r, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestLogger logs one line per request with status, bytes and duration.
func RequestLogger(loggerInstance *slog.Logger) func(http.Handler) http.Handler {
	loggerInstance = loggerOrDiscard(loggerInstance)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)
			status := wrapped.Status()
			if status == 0 {
				status = http.StatusOK
			}
			loggerInstance.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int("bytes", wrapped.BytesWritten()),
				slog.Duration("duration", time.Since(started)),
			)
		})
	}
}

// MaxBodySize rejects requests whose body exceeds the configured limit.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes <= 0 {
				next.ServeHTTP(w, r)
				return
			}
			if maxBytes > 0 && r.ContentLength > maxBytes {
				writeErrorResponse(w, r, http.StatusRequestEntityTooLarge, "request body is too large")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// DevelopmentCORS applies permissive CORS headers for development origins.
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

// writeErrorResponse emits the same envelope shape as the handler layer's
// WriteError without depending on it: {"success":false,"error":"<message>"}.
func writeErrorResponse(w http.ResponseWriter, r *http.Request, status int, message string) {
	if requestID := logger.RequestID(r.Context()); requestID != "" {
		w.Header().Set(logger.RequestIDHeader, requestID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "error": message})
}

func loggerOrDiscard(loggerInstance *slog.Logger) *slog.Logger {
	if loggerInstance != nil {
		return loggerInstance
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
