// Package httpserver owns the HTTP server lifecycle (construction, graceful
// shutdown) and health-check handlers. Request middleware assembly lives in
// the caller (cmd/server), which builds the full router before handing it to
// NewServer.
package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
	"github.com/shuTwT/nex-api/backend/internal/infra/logger"
)

// Dependencies carries the fully assembled router (including request
// middleware and health routes) plus an optional logger.
type Dependencies struct {
	Handler http.Handler
	Logger  *slog.Logger
}

// Server wraps an http.Server with lifecycle helpers.
type Server struct {
	config config.Config
	http   *http.Server
	logger *slog.Logger
}

// NewServer validates the configuration and wraps the provided handler.
func NewServer(cfg config.Config, dependencies Dependencies) (*Server, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate runtime config: %w", err)
	}
	loggerInstance := dependencies.Logger
	if loggerInstance == nil {
		var err error
		loggerInstance, err = logger.NewLogger(cfg.Log)
		if err != nil {
			return nil, fmt.Errorf("create runtime logger: %w", err)
		}
	}
	handler := dependencies.Handler
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	return &Server{
		config: cfg,
		logger: loggerInstance,
		http: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:           handler,
			ReadTimeout:       cfg.Server.ReadTimeout,
			ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
			WriteTimeout:      cfg.Server.WriteTimeout,
			IdleTimeout:       cfg.Server.IdleTimeout,
			MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
		},
	}, nil
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
