package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// DependencyCheck verifies a single readiness dependency.
type DependencyCheck struct {
	Name  string
	Check func(context.Context) error
}

// HealthResponse is the JSON body of the liveness/readiness endpoints.
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Health serves the /healthz (liveness) and /readyz (readiness) endpoints.
type Health struct {
	checks []DependencyCheck
	logger *slog.Logger
}

func NewHealth(checks ...DependencyCheck) *Health {
	return &Health{checks: append([]DependencyCheck(nil), checks...)}
}

func NewHealthWithLogger(logger *slog.Logger, checks ...DependencyCheck) *Health {
	return &Health{checks: append([]DependencyCheck(nil), checks...), logger: logger}
}

func (h *Health) Liveness(w http.ResponseWriter, _ *http.Request) {
	h.write(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (h *Health) Readiness(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string, len(h.checks))
	status := "ok"
	for _, dependency := range h.checks {
		if dependency.Check == nil {
			checks[dependency.Name] = "failed"
			status = "failed"
			continue
		}
		if err := dependency.Check(r.Context()); err != nil {
			checks[dependency.Name] = "failed"
			status = "failed"
			if h.logger != nil {
				h.logger.WarnContext(r.Context(), "readiness check failed", slog.String("dependency", dependency.Name), slog.Any("err", err))
			}
			continue
		}
		checks[dependency.Name] = "ok"
	}
	statusCode := http.StatusOK
	if status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}
	h.write(w, statusCode, HealthResponse{Status: status, Checks: checks})
}

func (h *Health) write(w http.ResponseWriter, status int, response HealthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil && h.logger != nil {
		h.logger.Error("write health response failed", slog.Any("err", err))
	}
}
