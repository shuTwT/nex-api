package settings

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicesettings "github.com/shuTwT/nex-api/internal/service/settings"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type Handler struct{ service *servicesettings.Service }

func NewHandler(service *servicesettings.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("settings: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux chi.Router, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("settings: mux and handler are required")
	}
	mux.Get("/api/system-settings", handler.list)
	mux.Put("/api/system-settings", handler.update)
	mux.Get("/api/system-settings/defaults", handler.defaults)
	mux.Get("/api/system-settings/announcement", handler.announcement)
	return nil
}

// list handles GET /api/system-settings.
//
// @Summary GET /api/system-settings
// @ID system_settings_route_get
// @Tags system-settings
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system-settings [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	items, err := h.service.List(r.Context(), r.URL.Query().Get("category"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, items)
}

// update handles PUT /api/system-settings.
//
// @Summary PUT /api/system-settings
// @ID system_settings_route_put
// @Tags system-settings
// @Accept json
// @Produce json
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system-settings [put]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	var request model.SystemSettingsUpdateReq
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	if err := h.service.Update(r.Context(), request.Settings); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "设置已更新"})
}

// defaults handles GET /api/system-settings/defaults.
//
// @Summary GET /api/system-settings/defaults
// @ID system_settings_defaults_route_get
// @Tags system-settings
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system-settings/defaults [get]
func (h *Handler) defaults(w http.ResponseWriter, _ *http.Request) {
	handlerutils.WriteData(w, http.StatusOK, h.service.Defaults())
}

// announcement handles GET /api/system-settings/announcement.
//
// @Summary GET /api/system-settings/announcement
// @ID system_settings_announcement_route_get
// @Tags system-settings
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/system-settings/announcement [get]
func (h *Handler) announcement(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.Announcement(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, values)
}
