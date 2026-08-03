package schedule

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/backend/internal/middleware"
	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	serviceschedule "github.com/shuTwT/nex-api/backend/internal/service/schedule"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
)

type Handler struct{ service *serviceschedule.Service }

func RegisterRoutes(mux chi.Router, service *serviceschedule.Service) error {
	if mux == nil || service == nil {
		return errors.New("schedule: mux and service are required")
	}
	handler := &Handler{service: service}
	admin := func(next http.Handler) http.Handler { return middleware.RequireAdmin(next) }
	mux.Method(http.MethodGet, "/api/scheduled-jobs", admin(http.HandlerFunc(handler.list)))
	mux.Method(http.MethodPost, "/api/scheduled-jobs", admin(http.HandlerFunc(handler.create)))
	mux.Method(http.MethodGet, "/api/scheduled-jobs/tasks", admin(http.HandlerFunc(handler.tasks)))
	mux.Method(http.MethodGet, "/api/scheduled-jobs/{id}", admin(http.HandlerFunc(handler.get)))
	mux.Method(http.MethodPut, "/api/scheduled-jobs/{id}", admin(http.HandlerFunc(handler.update)))
	mux.Method(http.MethodDelete, "/api/scheduled-jobs/{id}", admin(http.HandlerFunc(handler.delete)))
	mux.Method(http.MethodPost, "/api/scheduled-jobs/{id}/run", admin(http.HandlerFunc(handler.runNow)))
	return nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, items)
}

func (h *Handler) tasks(w http.ResponseWriter, _ *http.Request) {
	handlerutils.WriteData(w, http.StatusOK, h.service.Tasks())
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input model.ScheduleJobUpsertReq
	if err := handlerutils.DecodeJSON(r, &input); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.service.Create(r.Context(), input)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, item)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var input model.ScheduleJobUpsertReq
	if err := handlerutils.DecodeJSON(r, &input); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "scheduled job deleted"})
}

func (h *Handler) runNow(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RunNow(r.Context(), chi.URLParam(r, "id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusAccepted, map[string]string{"message": "scheduled job triggered"})
}
