package ads

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	serviceads "github.com/shuTwT/nex-api/backend/internal/service/ads"
	serviceauthz "github.com/shuTwT/nex-api/backend/internal/service/authz"
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

type createRequest struct {
	Image       string `json:"image"`
	ImageWidth  int    `json:"imageWidth"`
	ImageHeight int    `json:"imageHeight"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Position    string `json:"position"`
	IsActive    *bool  `json:"isActive"`
}

type updateRequest struct {
	Image       *string `json:"image"`
	ImageWidth  *int    `json:"imageWidth"`
	ImageHeight *int    `json:"imageHeight"`
	Link        *string `json:"link"`
	Title       *string `json:"title"`
	Position    *string `json:"position"`
	IsActive    *bool   `json:"isActive"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	page, err := positiveInt(r.URL.Query().Get("page"), 1)
	if err != nil {
		writeError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "page", Reason: "must be a positive integer"}))
		return
	}
	limit, err := positiveInt(r.URL.Query().Get("limit"), 10)
	if err != nil || limit > 100 {
		writeError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "limit", Reason: "must be between 1 and 100"}))
		return
	}
	var active *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("isActive")); raw != "" {
		value, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			writeError(w, r, appRuntime.NewValidationError(appRuntime.FieldError{Field: "isActive", Reason: "must be boolean"}))
			return
		}
		active = &value
	}
	result, err := h.service.List(r.Context(), serviceads.ListOptions{Search: r.URL.Query().Get("search"), Position: r.URL.Query().Get("position"), IsActive: active, Page: page, Limit: limit})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writePaginated(w, result.Items, page, limit, result.Total)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var request createRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	active := false
	if request.IsActive != nil {
		active = *request.IsActive
	}
	item, err := h.service.Create(r.Context(), serviceads.CreateInput{Image: request.Image, ImageWidth: request.ImageWidth, ImageHeight: request.ImageHeight, Link: request.Link, Title: request.Title, Position: request.Position, IsActive: active})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, item)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	var request updateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	item, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), serviceads.UpdateInput{Image: request.Image, ImageWidth: request.ImageWidth, ImageHeight: request.ImageHeight, Link: request.Link, Title: request.Title, Position: request.Position, IsActive: request.IsActive})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "广告已删除"})
}

func (h *Handler) toggle(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	item, err := h.service.Toggle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, item)
}

func (h *Handler) byPosition(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.ByPosition(r.Context(), chi.URLParam(r, "position"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, item)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.service.Stats(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, result)
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	principal, err := serviceauthz.AdminPolicy(r.Context())
	if err == nil && principal.UserID != "" {
		return true
	}
	status := http.StatusUnauthorized
	code, message := "unauthorized", "authentication required"
	if errors.Is(err, serviceauthz.ErrForbidden) {
		status, code, message = http.StatusForbidden, "forbidden", "access denied"
	}
	writeError(w, r, appRuntime.NewAPIError(status, code, message, err))
	return false
}

func positiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, errors.New("invalid positive integer")
	}
	return value, nil
}

func decodeJSON[T any](r *http.Request, destination *T) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "body", Reason: "invalid JSON"})
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return appRuntime.NewValidationError(appRuntime.FieldError{Field: "body", Reason: "must contain exactly one JSON value"})
	}
	return nil
}

func writeData[T any](w http.ResponseWriter, status int, data T) {
	_ = appRuntime.WriteData(w, status, data)
}

func writePaginated[T any](w http.ResponseWriter, data T, page, limit, total int) {
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	envelope := appRuntime.Envelope[T]{Success: true, Data: &data, Pagination: &appRuntime.Pagination{Page: page, PageSize: limit, Total: total, TotalPages: pages}}
	_ = appRuntime.WriteEnvelope(w, http.StatusOK, envelope)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	_ = appRuntime.WriteError(w, r, err)
}
