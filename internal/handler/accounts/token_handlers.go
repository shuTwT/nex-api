package accounts

import (
	"fmt"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	serviceaccounts "github.com/shuTwT/nex-api/internal/service/accounts"
	"net/http"
)

// listTokens handles GET /api/tokens.
//
// @Summary GET /api/tokens
// @ID tokens_route_get
// @Tags tokens
// @Produce json
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param status query string false "status"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.TokenResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens [get]
func (h *Handler) listTokens(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	page, err := parsePage(r)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	views, info, err := h.services.Tokens.List(r.Context(), owner.UserID, serviceaccounts.TokenFilter{Search: r.URL.Query().Get("search"), Status: r.URL.Query().Get("status")}, page)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, info.Page, info.PageSize, info.Total)
}

// createToken handles POST /api/tokens.
//
// @Summary POST /api/tokens
// @ID tokens_route_post
// @Tags tokens
// @Accept json
// @Produce json
// @Param body body model.TokenCreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.TokenCreateResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens [post]
func (h *Handler) createToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.TokenCreateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode token: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Tokens.Create(r.Context(), owner.UserID, request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, view)
}

// getToken handles GET /api/tokens/{id}.
//
// @Summary GET /api/tokens/{id}
// @ID tokens_id_route_get
// @Tags tokens
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.TokenResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens/{id} [get]
func (h *Handler) getToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Tokens.Get(r.Context(), owner.UserID, r.PathValue("id"))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// updateToken handles PUT /api/tokens/{id}.
//
// @Summary PUT /api/tokens/{id}
// @ID tokens_id_route_put
// @Tags tokens
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body model.TokenUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.TokenResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens/{id} [put]
func (h *Handler) updateToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.TokenUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode token: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Tokens.Update(r.Context(), owner.UserID, r.PathValue("id"), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// toggleToken handles PUT /api/tokens/{id}/toggle.
//
// @Summary PUT /api/tokens/{id}/toggle
// @ID tokens_id_toggle_route_put
// @Tags tokens
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.TokenResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens/{id}/toggle [put]
func (h *Handler) toggleToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Tokens.Toggle(r.Context(), owner.UserID, r.PathValue("id"), requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// deleteToken handles DELETE /api/tokens/{id}.
//
// @Summary DELETE /api/tokens/{id}
// @ID tokens_id_route_delete
// @Tags tokens
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens/{id} [delete]
func (h *Handler) deleteToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	if err := h.services.Tokens.Delete(r.Context(), owner.UserID, r.PathValue("id"), requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "令牌已删除"})
}

// tokenStats handles GET /api/tokens/stats.
//
// @Summary GET /api/tokens/stats
// @ID tokens_stats_route_get
// @Tags tokens
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.TokenStatsResp}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/tokens/stats [get]
func (h *Handler) tokenStats(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	view, err := h.services.Tokens.Stats(r.Context(), owner.UserID)
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}
