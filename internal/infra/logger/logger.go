// Package logger creates slog loggers and exposes the request-ID context
// helpers shared by the middleware and handler layers. It never depends on
// other layers.
package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/shuTwT/nex-api/internal/infra/config"
)

// RequestIDHeader is the HTTP header carrying the per-request correlation ID.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// WithRequestID stores a request ID in the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestID reads the request ID from the context, returning "" when absent.
func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

// requestContextHandler attaches the current request ID to every log record
// emitted while handling a request.
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

// NewLogger builds a slog logger from the configured level and format.
func NewLogger(cfg config.Log) (*slog.Logger, error) {
	return NewLoggerWithWriter(cfg, os.Stdout)
}

// NewLoggerWithWriter builds a slog logger writing to writer.
func NewLoggerWithWriter(cfg config.Log, writer io.Writer) (*slog.Logger, error) {
	if writer == nil {
		return nil, errors.New("logger writer is nil")
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(cfg.Level)))); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}
	opts := &slog.HandlerOptions{Level: level, AddSource: cfg.AddSource}
	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}
	return slog.New(requestContextHandler{Handler: handler}), nil
}

// DiscardLogger returns a logger that writes nowhere; used by middleware when
// no logger is configured.
func DiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
