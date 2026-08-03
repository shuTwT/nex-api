package httpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

func TestNewServer_injects_handler_and_health_routes(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	health := NewHealth()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Liveness)
	mux.HandleFunc("GET /injected", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted) })
	server, err := NewServer(cfg, Dependencies{Handler: mux})
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("health response = %d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/injected", nil))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("injected status = %d", rec.Code)
	}
}

func TestServer_readiness_reports_failed_dependency(t *testing.T) {
	health := NewHealth(DependencyCheck{Name: "database", Check: func(context.Context) error { return errors.New("database unavailable") }})
	rec := httptest.NewRecorder()
	health.Readiness(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"database":"failed"`) {
		t.Fatalf("readiness response = %s", rec.Body.String())
	}
}
