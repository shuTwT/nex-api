package schedule

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shuTwT/nex-api/internal/middleware"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceschedule "github.com/shuTwT/nex-api/internal/service/schedule"
	"github.com/shuTwT/nex-api/pkg/domain/model"
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

// list handles GET /api/scheduled-jobs.
//
// @Summary GET /api/scheduled-jobs
// @ID scheduled_jobs_route_get
// @Tags schedule
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, items)
}

// tasks handles GET /api/scheduled-jobs/tasks.
//
// @Summary GET /api/scheduled-jobs/tasks
// @ID scheduled_jobs_tasks_route_get
// @Tags schedule
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs/tasks [get]
func (h *Handler) tasks(w http.ResponseWriter, _ *http.Request) {
	handlerutils.WriteData(w, http.StatusOK, h.service.Tasks())
}

// get handles GET /api/scheduled-jobs/{id}.
//
// @Summary GET /api/scheduled-jobs/{id}
// @ID scheduled_jobs_id_route_get
// @Tags schedule
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, item)
}

// create handles POST /api/scheduled-jobs.
//
// @Summary POST /api/scheduled-jobs
// @ID scheduled_jobs_route_post
// @Tags schedule
// @Accept json
// @Produce json
// @Param body body main.SwaggerRequest true "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs [post]
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

// update handles PUT /api/scheduled-jobs/{id}.
//
// @Summary PUT /api/scheduled-jobs/{id}
// @ID scheduled_jobs_id_route_put
// @Tags schedule
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body main.SwaggerRequest true "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs/{id} [put]
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

// delete handles DELETE /api/scheduled-jobs/{id}.
//
// @Summary DELETE /api/scheduled-jobs/{id}
// @ID scheduled_jobs_id_route_delete
// @Tags schedule
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs/{id} [delete]
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "scheduled job deleted"})
}

// runNow handles POST /api/scheduled-jobs/{id}/run.
//
// @Summary POST /api/scheduled-jobs/{id}/run
// @ID scheduled_jobs_id_run_route_post
// @Tags schedule
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/scheduled-jobs/{id}/run [post]
func (h *Handler) runNow(w http.ResponseWriter, r *http.Request) {
	if err := h.service.RunNow(r.Context(), chi.URLParam(r, "id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusAccepted, map[string]string{"message": "scheduled job triggered"})
}
