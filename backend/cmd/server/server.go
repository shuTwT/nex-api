package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
	"github.com/shuTwT/nex-api/backend/internal/infra/httpserver"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
)

// buildHTTPServer assembles the request middleware chain, health routes and
// the business router, then wraps them in the HTTP server.
func buildHTTPServer(ctx context.Context, cfg config.Config, deps serverDependencies, logger *slog.Logger, businessHandler http.Handler) (*httpserver.Server, error) {
	httpRouter := chi.NewRouter()
	httpRouter.Use(middleware.RequestIDMiddleware)
	httpRouter.Use(middleware.Recovery(logger))
	httpRouter.Use(middleware.RequestLogger(logger))
	httpRouter.Use(middleware.MaxBodySize(cfg.Server.MaxBodyBytes))
	if strings.EqualFold(cfg.Environment, "development") {
		httpRouter.Use(middleware.DevelopmentCORS(cfg.Server.CORSOrigins))
	}

	health := httpserver.NewHealthWithLogger(logger,
		httpserver.DependencyCheck{
			Name: "database",
			Check: func(ctx context.Context) error {
				_, err := deps.Client.User.Query().Count(ctx)
				return err
			},
		},
		httpserver.DependencyCheck{
			Name:  "redis",
			Check: func(ctx context.Context) error { return deps.Redis.Ping(ctx).Err() },
		},
	)
	httpRouter.Get("/healthz", health.Liveness)
	httpRouter.Get("/readyz", health.Readiness)
	httpRouter.Mount("/", middleware.SessionAuth(deps.AuthService, businessHandler))

	return httpserver.NewServer(cfg, httpserver.Dependencies{
		Handler: httpRouter,
		Logger:  logger,
	})
}
