package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/shuTwT/nex-api/internal/handler/cron"
	"github.com/shuTwT/nex-api/internal/handler/oauth"
	"github.com/shuTwT/nex-api/internal/handler/router"
	"github.com/shuTwT/nex-api/internal/infra/config"
	servicegateway "github.com/shuTwT/nex-api/internal/service/gateway"
)

func TestBuildRouterUsesChiForHandwrittenAndOpenAPIRoutes(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		Environment: "test",
		AppURL:      "http://localhost:3000",
		Database: config.Database{
			URL:             "file:" + filepath.Join(t.TempDir(), "router.db"),
			ConnMaxLifetime: time.Minute,
		},
		Redis: config.Redis{URL: "redis://localhost:6379/0"},
		Auth: config.Auth{
			SessionSecret:     strings.Repeat("s", 32),
			SessionCookieName: "session",
		},
		Upload: config.Upload{
			Directory:     filepath.Join(t.TempDir(), "uploads"),
			AllowedTypes:  []string{"image/png"},
			CreateOnStart: true,
		},
		Cron:   config.Cron{Interval: time.Minute},
		Server: config.Server{ShutdownTimeout: time.Second},
	}
	deps, scheduler, err := buildServices(ctx, cfg, slog.Default())
	if err != nil {
		t.Fatalf("build services: %v", err)
	}
	t.Cleanup(func() {
		shutdownServerResources(ctx, time.Second, deps, scheduler, slog.Default())
	})
	deps.Accountant = servicegateway.NewEntAccountant(deps.Client)

	handler, err := router.BuildRouter(ctx, router.Config{
		Cron: cron.Config{Enabled: true, Secret: "cron-secret"},
		OAuth: oauth.Config{
			AppURL:        cfg.AppURL,
			SessionSecret: []byte(cfg.Auth.SessionSecret),
		},
	}, deps.Dependencies)
	if err != nil {
		t.Fatalf("build router: %v", err)
	}
	if _, ok := handler.(*chi.Mux); !ok {
		t.Fatalf("router type = %T, want *chi.Mux", handler)
	}

	checks := []struct {
		name   string
		method string
		path   string
		status int
	}{
		{name: "handwritten", method: http.MethodGet, path: "/api/system/initialized", status: http.StatusOK},
		{name: "oauth alongside auth", method: http.MethodGet, path: "/api/auth/providers", status: http.StatusOK},
		{name: "generated gateway", method: http.MethodGet, path: "/api/v1/missing", status: http.StatusUnauthorized},
		{name: "generated cron", method: http.MethodPost, path: "/api/cron/sync-stats", status: http.StatusUnauthorized},
		{name: "method not allowed", method: http.MethodPost, path: "/api/system/initialized", status: http.StatusMethodNotAllowed},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(check.method, check.path, nil))
			if recorder.Code != check.status {
				t.Fatalf("%s %s status = %d, want %d; body=%s", check.method, check.path, recorder.Code, check.status, recorder.Body.String())
			}
		})
	}
}
