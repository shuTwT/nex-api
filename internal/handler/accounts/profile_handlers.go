package accounts

import (
	"fmt"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	"net/http"
)

// getProfile handles GET /api/personal/profile.
//
// @Summary GET /api/personal/profile
// @ID personal_profile_route_get
// @Tags personal
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.ProfileResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/personal/profile [get]
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

// updateProfile handles PUT /api/personal/profile.
//
// @Summary PUT /api/personal/profile
// @ID personal_profile_route_put
// @Tags personal
// @Accept json
// @Produce json
// @Param body body model.ProfileUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.ProfileResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/personal/profile [put]
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

// updatePassword handles PUT /api/personal/profile/password.
//
// @Summary PUT /api/personal/profile/password
// @ID personal_profile_password_route_put
// @Tags personal
// @Accept json
// @Produce json
// @Param body body model.PasswordUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/personal/profile/password [put]
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
