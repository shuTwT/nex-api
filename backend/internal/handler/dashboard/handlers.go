package dashboard

import (
	"errors"
	"net/http"

	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
	servicedashboard "github.com/shuTwT/nex-api/backend/internal/service/dashboard"
)

type Handler struct{ service *servicedashboard.Service }

func NewHandler(service *servicedashboard.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("dashboard: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("dashboard: mux and handler are required")
	}
	mux.HandleFunc("GET /api/dashboard/stats", handler.dashboardStats)
	mux.HandleFunc("GET /api/dashboard/activity", handler.activity)
	mux.HandleFunc("GET /api/dashboard/top-apis", handler.topAPIs)
	mux.HandleFunc("GET /api/dashboard/usage-trend", handler.usageTrend)
	mux.HandleFunc("GET /api/usage", handler.usage)
	mux.HandleFunc("GET /api/stats", handler.globalStats)
	mux.HandleFunc("GET /api/stats/{alias}", handler.apiStats)
	return nil
}

func (h *Handler) dashboardStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	stats, err := h.service.DashboardStats(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}

func (h *Handler) activity(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	activities, err := h.service.Activity(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, activities)
}

func (h *Handler) topAPIs(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	result, err := h.service.TopAPIs(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

func (h *Handler) usageTrend(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	trend, err := h.service.UsageTrend(r.Context(), principal.UserID, principal.Role)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, trend)
}

func (h *Handler) usage(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	usage, err := h.service.Usage(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, usage)
}

func (h *Handler) globalStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireUser(w, r); !ok {
		return
	}
	stats, err := h.service.GlobalStats(r.Context(), r.URL.Query().Get("type"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}

func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	stats, err := h.service.APIStats(r.Context(), principal.UserID, r.PathValue("alias"), r.URL.Query().Get("user") == "true")
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}

func (h *Handler) requireUser(w http.ResponseWriter, r *http.Request) (serviceauthz.Principal, bool) {
	principal, err := serviceauthz.UserPolicy(r.Context())
	if err != nil {
		status := http.StatusUnauthorized
		code := "unauthorized"
		message := "authentication required"
		if errors.Is(err, serviceauthz.ErrForbidden) {
			status = http.StatusForbidden
			code = "forbidden"
			message = "access denied"
		}
		writeError(w, r, appRuntime.NewAPIError(status, code, message, err))
		return serviceauthz.Principal{}, false
	}
	return principal, true
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = appRuntime.WriteData(w, status, data)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}
