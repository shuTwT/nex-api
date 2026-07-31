package accounts

import (
	"fmt"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"net/http"
)

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
	writePage(w, http.StatusOK, views, info)
}

func (h *Handler) createAudit(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var entry AuditEntry
	if err := decodeJSON(r, &entry); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode audit: %w", ErrInvalidRequest))
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
	writeData(w, http.StatusCreated, view)
}

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
		writeServiceError(w, r, runtime.ErrForbidden)
		return
	}
	writeData(w, http.StatusOK, view)
}

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
		writeServiceError(w, r, runtime.ErrForbidden)
		return
	}
	var entry AuditEntry
	if err := decodeJSON(r, &entry); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode audit: %w", ErrInvalidRequest))
		return
	}
	view, err := h.services.Audits.Update(r.Context(), r.PathValue("id"), entry, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

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
		writeServiceError(w, r, runtime.ErrForbidden)
		return
	}
	if err := h.services.Audits.Delete(r.Context(), r.PathValue("id"), requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "审计日志已删除"})
}

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
	if err := h.services.Audits.Export(r.Context(), filter, w); err != nil {
		return
	}
}

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
	writeData(w, http.StatusOK, view)
}
