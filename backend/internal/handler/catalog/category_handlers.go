package catalog

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
)

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.categories.List(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	data := make([]categoryListData, len(items))
	for index, item := range items {
		data[index] = toCategoryListData(item)
	}
	handlerutils.WriteData(w, http.StatusOK, data)
}

func (h *Handler) createCategory(w http.ResponseWriter, r *http.Request) {
	var request categoryRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.categories.Create(r.Context(), servicecatalog.CategoryInput{Name: request.Name, Description: request.Description, Icon: request.Icon})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, toCategoryData(item))
}

func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	item, err := h.categories.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toCategoryData(item))
}

func (h *Handler) updateCategory(w http.ResponseWriter, r *http.Request) {
	var request categoryUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.categories.Update(r.Context(), r.PathValue("id"), servicecatalog.CategoryUpdateInput{Name: request.Name, Description: request.Description, Icon: request.Icon})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toCategoryData(item))
}

func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if err := h.categories.Delete(r.Context(), r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "分类已删除"})
}
