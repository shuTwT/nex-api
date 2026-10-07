package accounts

import (
	"fmt"
	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	"net/http"
)

// listAudits handles GET /api/audit-logs.
//
// @Summary GET /api/audit-logs
// @ID audit_logs_route_get
// @Tags audit-logs
// @Produce json
// @Security SessionCookie
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Param search query string false "keyword search"
// @Param level query string false "audit level filter"
// @Param status query string false "audit status filter"
// @Param startDate query string false "RFC3339 start time"
// @Param endDate query string false "RFC3339 end time"
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.AuditResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs [get]
func (h *Handler) listAudits(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	filter, err := auditFilter(r)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if owner.Role != "admin" {
		filter.UserID = owner.UserID
	}
	page, err := parsePage(r)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	views, info, err := h.services.Audits.List(r.Context(), filter, page)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, info.Page, info.PageSize, info.Total)
}

// createAudit handles POST /api/audit-logs.
//
// @Summary POST /api/audit-logs
// @ID audit_logs_route_post
// @Tags audit-logs
// @Accept json
// @Produce json
// @Param body body accounts.AuditEntry true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AuditResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs [post]
func (h *Handler) createAudit(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var entry serviceaccounts.AuditEntry
	if err := handlerutils.DecodeJSON(r, &entry); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode audit: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	entry.UserID = owner.UserID
	metadata := requestMetadata(r)
	entry.IPAddress, entry.UserAgent, entry.Metadata = metadata.IP, metadata.UserAgent, metadata.Metadata
	view, err := h.services.Audits.Create(r.Context(), entry)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, view)
}

// getAudit handles GET /api/audit-logs/{id}.
//
// @Summary GET /api/audit-logs/{id}
// @ID audit_logs_id_route_get
// @Tags audit
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AuditResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs/{id} [get]
func (h *Handler) getAudit(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Audits.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if owner.Role != "admin" && view.UserID != owner.UserID {
		writeServiceError(w, r, appRuntime.ErrForbidden)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// updateAudit handles PUT /api/audit-logs/{id}.
//
// @Summary PUT /api/audit-logs/{id}
// @ID audit_logs_id_route_put
// @Tags audit
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body accounts.AuditEntry true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AuditResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs/{id} [put]
func (h *Handler) updateAudit(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	current, err := h.services.Audits.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if owner.Role != "admin" && current.UserID != owner.UserID {
		writeServiceError(w, r, appRuntime.ErrForbidden)
		return
	}
	var entry serviceaccounts.AuditEntry
	if err := handlerutils.DecodeJSON(r, &entry); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode audit: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Audits.Update(r.Context(), r.PathValue("id"), entry, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// deleteAudit handles DELETE /api/audit-logs/{id}.
//
// @Summary DELETE /api/audit-logs/{id}
// @ID audit_logs_id_route_delete
// @Tags audit
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs/{id} [delete]
func (h *Handler) deleteAudit(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	current, err := h.services.Audits.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if owner.Role != "admin" && current.UserID != owner.UserID {
		writeServiceError(w, r, appRuntime.ErrForbidden)
		return
	}
	if err := h.services.Audits.Delete(r.Context(), r.PathValue("id"), requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "审计日志已删除"})
}

// exportAudits handles GET /api/audit-logs/export.
//
// @Summary GET /api/audit-logs/export
// @ID audit_logs_export_route_get
// @Tags audit-logs
// @Produce text/csv
// @Security SessionCookie
// @Param search query string false "keyword search"
// @Param level query string false "audit level filter"
// @Param status query string false "audit status filter"
// @Param startDate query string false "RFC3339 start time"
// @Param endDate query string false "RFC3339 end time"
// @Success 200 {string} string "CSV document"
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs/export [get]
func (h *Handler) exportAudits(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	filter, err := auditFilter(r)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if owner.Role != "admin" {
		filter.UserID = owner.UserID
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="audit-logs.csv"`)
	document, err := h.services.Audits.Export(r.Context(), filter)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	_, _ = w.Write(document)
}

// auditStats handles GET /api/audit-logs/stats.
//
// @Summary GET /api/audit-logs/stats
// @ID audit_logs_stats_route_get
// @Tags audit-logs
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.AuditStatsResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/audit-logs/stats [get]
func (h *Handler) auditStats(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	userID := ""
	if owner.Role != "admin" {
		userID = owner.UserID
	}
	view, err := h.services.Audits.Stats(r.Context(), userID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}
