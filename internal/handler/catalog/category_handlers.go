package catalog

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
)

// listCategories handles GET /api/categories.
//
// @Summary GET /api/categories
// @ID categories_route_get
// @Tags categories
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/categories [get]
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

// createCategory handles POST /api/categories.
//
// @Summary POST /api/categories
// @ID categories_route_post
// @Tags categories
// @Accept json
// @Produce json
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/categories [post]
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

// getCategory handles GET /api/categories/{id}.
//
// @Summary GET /api/categories/{id}
// @ID categories_id_route_get
// @Tags catalog
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/categories/{id} [get]
func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	item, err := h.categories.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toCategoryData(item))
}

// updateCategory handles PUT /api/categories/{id}.
//
// @Summary PUT /api/categories/{id}
// @ID categories_id_route_put
// @Tags categories
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body main.SwaggerRequest false "JSON request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/categories/{id} [put]
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

// deleteCategory handles DELETE /api/categories/{id}.
//
// @Summary DELETE /api/categories/{id}
// @ID categories_id_route_delete
// @Tags categories
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/categories/{id} [delete]
func (h *Handler) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if err := h.categories.Delete(r.Context(), r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "分类已删除"})
}
