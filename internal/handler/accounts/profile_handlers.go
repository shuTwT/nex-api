package accounts

import (
	"fmt"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
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
	handlerutils.WriteData(w, http.StatusOK, view)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.ProfileUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode profile: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Profiles.Update(r.Context(), owner.UserID, request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.PasswordUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode password: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	if err := h.services.Profiles.UpdatePassword(r.Context(), owner.UserID, request, requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "密码已更新"})
}
