package dashboard

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceauthz "github.com/shuTwT/nex-api/internal/service/authz"
	servicedashboard "github.com/shuTwT/nex-api/internal/service/dashboard"
)

type Handler struct{ service *servicedashboard.Service }

func NewHandler(service *servicedashboard.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("dashboard: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux chi.Router, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("dashboard: mux and handler are required")
	}
	mux.Get("/api/dashboard/stats", handler.dashboardStats)
	mux.Get("/api/dashboard/activity", handler.activity)
	mux.Get("/api/dashboard/top-apis", handler.topAPIs)
	mux.Get("/api/dashboard/usage-trend", handler.usageTrend)
	mux.Get("/api/usage", handler.usage)
	mux.Get("/api/stats", handler.globalStats)
	mux.Get("/api/stats/{alias}", handler.apiStats)
	return nil
}

// dashboardStats handles GET /api/dashboard/stats.
//
// @Summary GET /api/dashboard/stats
// @ID dashboard_stats_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/dashboard/stats [get]
func (h *Handler) dashboardStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	stats, err := h.service.DashboardStats(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}

// activity handles GET /api/dashboard/activity.
//
// @Summary GET /api/dashboard/activity
// @ID dashboard_activity_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/dashboard/activity [get]
func (h *Handler) activity(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	activities, err := h.service.Activity(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, activities)
}

// topAPIs handles GET /api/dashboard/top-apis.
//
// @Summary GET /api/dashboard/top-apis
// @ID dashboard_top_apis_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/dashboard/top-apis [get]
func (h *Handler) topAPIs(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	result, err := h.service.TopAPIs(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, result)
}

// usageTrend handles GET /api/dashboard/usage-trend.
//
// @Summary GET /api/dashboard/usage-trend
// @ID dashboard_usage_trend_route_get
// @Tags dashboard
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/dashboard/usage-trend [get]
func (h *Handler) usageTrend(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	trend, err := h.service.UsageTrend(r.Context(), principal.UserID, principal.Role)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, trend)
}

// usage handles GET /api/usage.
//
// @Summary GET /api/usage
// @ID usage_route_get
// @Tags usage
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/usage [get]
func (h *Handler) usage(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	usage, err := h.service.Usage(r.Context(), principal.UserID)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, usage)
}

// globalStats handles GET /api/stats.
//
// @Summary GET /api/stats
// @ID stats_route_get
// @Tags stats
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/stats [get]
func (h *Handler) globalStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireUser(w, r); !ok {
		return
	}
	stats, err := h.service.GlobalStats(r.Context(), r.URL.Query().Get("type"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}

// apiStats handles GET /api/stats/{alias}.
//
// @Summary GET /api/stats/{alias}
// @ID stats_alias_route_get
// @Tags stats
// @Produce json
// @Param alias path string true "alias"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/stats/{alias} [get]
func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireUser(w, r)
	if !ok {
		return
	}
	stats, err := h.service.APIStats(r.Context(), principal.UserID, chi.URLParam(r, "alias"), r.URL.Query().Get("user") == "true")
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
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
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(status, code, message, err))
		return serviceauthz.Principal{}, false
	}
	return principal, true
}
