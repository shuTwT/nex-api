package accounts

import (
	"fmt"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	"net/http"
)

// listUsers handles GET /api/users.
//
// @Summary GET /api/users
// @ID users_route_get
// @Tags users
// @Produce json
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param role query string false "role"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.UserResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users [get]
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	page, err := parsePage(r)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	views, info, err := h.services.Users.List(r.Context(), serviceaccounts.UserListFilter{Role: r.URL.Query().Get("role"), Search: r.URL.Query().Get("search")}, page)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, info.Page, info.PageSize, info.Total)
}

// createUser handles POST /api/users.
//
// @Summary POST /api/users
// @ID users_route_post
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.UserCreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.UserResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users [post]
func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.UserCreateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode user: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Users.Create(r.Context(), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, view)
}

// userStats handles GET /api/users/stats.
//
// @Summary GET /api/users/stats
// @ID users_stats_route_get
// @Tags users
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.UserStatsResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users/stats [get]
func (h *Handler) userStats(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Users.Stats(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// getUser handles GET /api/users/{id}.
//
// @Summary GET /api/users/{id}
// @ID users_id_route_get
// @Tags users
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.UserResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users/{id} [get]
func (h *Handler) getUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Users.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// updateUser handles PUT /api/users/{id}.
//
// @Summary PUT /api/users/{id}
// @ID users_id_route_put
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body model.UserUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.UserResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users/{id} [put]
func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.UserUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode user: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Users.Update(r.Context(), r.PathValue("id"), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// deleteUser handles DELETE /api/users/{id}.
//
// @Summary DELETE /api/users/{id}
// @ID users_id_route_delete
// @Tags users
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/users/{id} [delete]
func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	if err := h.services.Users.Delete(r.Context(), r.PathValue("id"), requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "用户已删除"})
}
