package settings

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	servicesettings "github.com/shuTwT/nex-api/backend/internal/service/settings"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
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

func (h *Handler) defaults(w http.ResponseWriter, _ *http.Request) {
	handlerutils.WriteData(w, http.StatusOK, h.service.Defaults())
}

func (h *Handler) announcement(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.Announcement(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, values)
}
