package runtime

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/config"
)

func TestNewServer_injects_handler_and_health_routes(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}),
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}

	// When
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected healthy response, got %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/injected", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected injected handler status 202, got %d", rec.Code)
	}
}

func TestServer_recovery_returns_typed_error_envelope_and_request_id(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("test panic")
		}),
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	rec := httptest.NewRecorder()

	// When
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if rec.Header().Get(RequestIDHeader) != "test-request-id" {
		t.Fatalf("expected request ID header, got %q", rec.Header().Get(RequestIDHeader))
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Fatalf("expected internal error envelope, got %s", rec.Body.String())
	}
}

func TestServer_readiness_reports_failed_dependency(t *testing.T) {
	// Given
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load development config: %v", err)
	}
	server, err := NewServer(cfg, Dependencies{
		Readiness: []DependencyCheck{{
			Name:  "database",
			Check: func(context.Context) error { return errors.New("database unavailable") },
		}},
	})
	if err != nil {
		t.Fatalf("construct server: %v", err)
	}

	// When
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	// Then
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected readiness status 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"database":"failed"`) {
		t.Fatalf("expected failed database check, got %s", rec.Body.String())
	}
}
