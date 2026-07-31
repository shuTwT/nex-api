package runtime

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/backend/internal/config"
)

type Dependencies struct {
	Handler   http.Handler
	Readiness []DependencyCheck
	Logger    *slog.Logger
}

type Server struct {
	config config.Config
	http   *http.Server
	logger *slog.Logger
}

func NewServer(cfg config.Config, dependencies Dependencies) (*Server, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate runtime config: %w", err)
	}
	logger := dependencies.Logger
	if logger == nil {
		var err error
		logger, err = NewLogger(cfg.Log)
		if err != nil {
			return nil, fmt.Errorf("create runtime logger: %w", err)
		}
	}

	router := chi.NewRouter()
	router.Use(RequestIDMiddleware)
	router.Use(Recovery(logger))
	router.Use(RequestLogger(logger))
	router.Use(MaxBodySize(cfg.Server.MaxBodyBytes))
	if strings.EqualFold(cfg.Environment, "development") {
		router.Use(DevelopmentCORS(cfg.Server.CORSOrigins))
	}

	health := NewHealthWithLogger(logger, dependencies.Readiness...)
	router.Get("/healthz", health.Liveness)
	router.Get("/readyz", health.Readiness)
	if dependencies.Handler == nil {
		dependencies.Handler = http.NotFoundHandler()
	}
	router.Mount("/", dependencies.Handler)

	return &Server{
		config: cfg,
		logger: logger,
		http: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:           router,
			ReadTimeout:       cfg.Server.ReadTimeout,
			ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
			MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
		},
	}, nil
}

func NewRouter(cfg config.Config, dependencies Dependencies) (http.Handler, error) {
	server, err := NewServer(cfg, dependencies)
	if err != nil {
		return nil, err
	}
	return server.Handler(), nil
}

func NewLogger(cfg config.Log) (*slog.Logger, error) {
	return NewLoggerWithWriter(cfg, os.Stdout)
}

func NewLoggerWithWriter(cfg config.Log, writer io.Writer) (*slog.Logger, error) {
	if writer == nil {
		return nil, fmt.Errorf("logger writer is nil")
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

func (s *Server) Handler() http.Handler { return s.http.Handler }

func (s *Server) Addr() string { return s.http.Addr }

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	return nil
}

func (s *Server) Run(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("server context is nil")
	}
	errCh := make(chan error, 1)
	go func() {
		err := s.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("serve HTTP: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), s.config.Server.ShutdownTimeout)
		defer cancel()
		if err := s.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}

func (s *Server) RunWithSignals(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("server context is nil")
	}
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return s.Run(signalCtx)
}
