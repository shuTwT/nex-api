package accounts

import (
	"fmt"
	"net/http"
)

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Profiles.Get(r.Context(), owner.UserID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request ProfileUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode profile: %w", ErrInvalidRequest))
		return
	}
	view, err := h.services.Profiles.Update(r.Context(), owner.UserID, request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request PasswordUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode password: %w", ErrInvalidRequest))
		return
	}
	if err := h.services.Profiles.UpdatePassword(r.Context(), owner.UserID, request, requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "密码已更新"})
}
