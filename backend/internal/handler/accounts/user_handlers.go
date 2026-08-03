package accounts

import (
	"fmt"
	serviceaccounts "github.com/shuTwT/nex-api/backend/internal/service/accounts"
	"net/http"
)

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
	writePage(w, http.StatusOK, views, info)
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.UserCreateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode user: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Users.Create(r.Context(), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, view)
}

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
	writeData(w, http.StatusOK, view)
}

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
	writeData(w, http.StatusOK, view)
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	var request serviceaccounts.UserUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeServiceError(w, r, fmt.Errorf("decode user: %w", serviceaccounts.ErrInvalidRequest))
		return
	}
	view, err := h.services.Users.Update(r.Context(), r.PathValue("id"), request, requestMetadata(r))
	if err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, view)
}

func (h *Handler) deleteUser(w http.ResponseWriter, r *http.Request) {
	if _, err := admin(r.Context()); err != nil {
		writeServiceError(w, r, err)
		return
	}
	if err := h.services.Users.Delete(r.Context(), r.PathValue("id"), requestMetadata(r)); err != nil {
		writeServiceError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "用户已删除"})
}
