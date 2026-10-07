package ads

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceads "github.com/shuTwT/nex-api/internal/service/ads"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type Handler struct{ service *serviceads.Service }

func NewHandler(service *serviceads.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("ads: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux chi.Router, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("ads: mux and handler are required")
	}
	mux.Get("/api/advertisements", handler.list)
	mux.Post("/api/advertisements", handler.create)
	mux.Get("/api/advertisements/stats", handler.stats)
	mux.Get("/api/advertisements/by-position/{position}", handler.byPosition)
	mux.Get("/api/advertisements/{id}", handler.get)
	mux.Put("/api/advertisements/{id}", handler.update)
	mux.Delete("/api/advertisements/{id}", handler.delete)
	mux.Put("/api/advertisements/{id}/toggle", handler.toggle)
	return nil
}

// list handles GET /api/advertisements.
//
// @Summary GET /api/advertisements
// @ID advertisements_route_get
// @Tags advertisements
// @Produce json
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param position query string false "position"
// @Param isActive query string false "isActive"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	page, err := handlerutils.PositiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "page", Reason: "must be a positive integer"}))
		return
	}
	limit, err := handlerutils.PositiveInt(r.URL.Query().Get("limit"), 10)
	if err != nil || limit > 100 {
		handlerutils.WriteError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "limit", Reason: "must be between 1 and 100"}))
		return
	}
	var active *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("isActive")); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			handlerutils.WriteError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "isActive", Reason: "must be boolean"}))
			return
		}
		active = &value
	}
	result, err := h.service.List(r.Context(), serviceads.ListOptions{Search: r.URL.Query().Get("search"), Position: r.URL.Query().Get("position"), IsActive: active, Page: page, Limit: limit})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, result.Items, page, limit, result.Total)
}

// create handles POST /api/advertisements.
//
// @Summary POST /api/advertisements
// @ID advertisements_route_post
// @Tags advertisements
// @Accept json
// @Produce json
// @Param body body model.AdvertisementCreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	var request model.AdvertisementCreateReq
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	active := false
	if request.IsActive != nil {
		active = *request.IsActive
	}
	item, err := h.service.Create(r.Context(), serviceads.CreateInput{Image: request.Image, ImageWidth: request.ImageWidth, ImageHeight: request.ImageHeight, Link: request.Link, Title: request.Title, Position: request.Position, IsActive: active})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, item)
}

// get handles GET /api/advertisements/{id}.
//
// @Summary GET /api/advertisements/{id}
// @ID advertisements_id_route_get
// @Tags advertisements
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

// update handles PUT /api/advertisements/{id}.
//
// @Summary PUT /api/advertisements/{id}
// @ID advertisements_id_route_put
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body model.AdvertisementUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/{id} [put]
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	var request model.AdvertisementUpdateReq
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), serviceads.UpdateInput{Image: request.Image, ImageWidth: request.ImageWidth, ImageHeight: request.ImageHeight, Link: request.Link, Title: request.Title, Position: request.Position, IsActive: request.IsActive})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

// delete handles DELETE /api/advertisements/{id}.
//
// @Summary DELETE /api/advertisements/{id}
// @ID advertisements_id_route_delete
// @Tags advertisements
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/{id} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "广告已删除"})
}

// toggle handles PUT /api/advertisements/{id}/toggle.
//
// @Summary PUT /api/advertisements/{id}/toggle
// @ID advertisements_id_toggle_route_put
// @Tags advertisements
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/{id}/toggle [put]
func (h *Handler) toggle(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	item, err := h.service.Toggle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

// byPosition handles GET /api/advertisements/by-position/{position}.
//
// @Summary GET /api/advertisements/by-position/{position}
// @ID advertisements_by_position_position_route_get
// @Tags advertisements
// @Produce json
// @Param position path string true "position"
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/by-position/{position} [get]
func (h *Handler) byPosition(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.ByPosition(r.Context(), chi.URLParam(r, "position"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

// stats handles GET /api/advertisements/stats.
//
// @Summary GET /api/advertisements/stats
// @ID advertisements_stats_route_get
// @Tags advertisements
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AdvertisementStatsResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/advertisements/stats [get]
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	if !handlerutils.RequireAdmin(w, r) {
		return
	}
	result, err := h.service.Stats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, result)
}
