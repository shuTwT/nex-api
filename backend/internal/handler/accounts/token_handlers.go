package accounts

import (
	"fmt"
	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	"net/http"
)

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
	writePage(w, http.StatusOK, views, info)
}

func (h *Handler) createToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.TokenCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode token: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Tokens.Create(r.Context(), owner.UserID, request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, view)
}

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
	writeData(w, http.StatusOK, view)
}

func (h *Handler) updateToken(w http.ResponseWriter, r *http.Request) {
	owner, err := principal(r.Context())
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.TokenUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode token: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Tokens.Update(r.Context(), owner.UserID, r.PathValue("id"), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

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
	writeData(w, http.StatusOK, view)
}

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
	writeData(w, http.StatusOK, map[string]string{"message": "令牌已删除"})
}

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
	writeData(w, http.StatusOK, view)
}
