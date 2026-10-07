package catalog

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
)

// listAPIs handles GET /api/apis.
//
// @Summary GET /api/apis
// @ID apis_route_get
// @Tags apis
// @Produce json
// @Security SessionCookie
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.CatalogAPIDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis [get]
func (h *Handler) listAPIs(w http.ResponseWriter, r *http.Request) {
	options, err := parseAPIListOptions(r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.apis.List(r.Context(), options)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	data := make([]apiData, len(result.Items))
	for index, item := range result.Items {
		data[index] = toAPIData(item, false)
	}
	handlerutils.WritePaginated(w, data, options.Page, options.Limit, result.Total)
}

// createAPI handles POST /api/apis.
//
// @Summary POST /api/apis
// @ID apis_route_post
// @Tags apis
// @Accept json
// @Produce json
// @Param body body model.CatalogAPICreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogAPIDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis [post]
func (h *Handler) createAPI(w http.ResponseWriter, r *http.Request) {
	var request apiRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	input := servicecatalog.APIInput{Name: request.Name, Alias: request.Alias, Description: request.Description, Endpoint: request.Endpoint, Method: request.Method, CategoryID: request.CategoryID, Documentation: request.Documentation, PreScript: request.PreScript, PostScript: request.PostScript, Pricing: intValue(request.Pricing), IsActive: boolValue(request.IsActive, true)}
	item, err := h.apis.Create(r.Context(), input)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, toAPIData(item, true))
}

// getAPI handles GET /api/apis/{id}.
//
// @Summary GET /api/apis/{id}
// @ID apis_id_route_get
// @Tags apis
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogAPIDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis/{id} [get]
func (h *Handler) getAPI(w http.ResponseWriter, r *http.Request) {
	item, err := h.apis.GetByIdentifier(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, true))
}

// updateAPI handles PUT /api/apis/{id}.
//
// @Summary PUT /api/apis/{id}
// @ID apis_id_route_put
// @Tags apis
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body model.CatalogAPIUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogAPIDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis/{id} [put]
func (h *Handler) updateAPI(w http.ResponseWriter, r *http.Request) {
	var request apiUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.apis.Update(r.Context(), r.PathValue("id"), servicecatalog.APIUpdateInput{Name: request.Name, Alias: request.Alias, Description: request.Description, Endpoint: request.Endpoint, Method: request.Method, CategoryID: request.CategoryID, Pricing: request.Pricing, Documentation: request.Documentation, PreScript: request.PreScript, PostScript: request.PostScript, IsActive: request.IsActive})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, false))
}

// deleteAPI handles DELETE /api/apis/{id}.
//
// @Summary DELETE /api/apis/{id}
// @ID apis_id_route_delete
// @Tags apis
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis/{id} [delete]
func (h *Handler) deleteAPI(w http.ResponseWriter, r *http.Request) {
	if err := h.apis.Delete(r.Context(), r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "API 已删除"})
}

// toggleAPI handles PUT /api/apis/{id}/toggle.
//
// @Summary PUT /api/apis/{id}/toggle
// @ID apis_id_toggle_route_put
// @Tags apis
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogAPIDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis/{id}/toggle [put]
func (h *Handler) toggleAPI(w http.ResponseWriter, r *http.Request) {
	item, err := h.apis.Toggle(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toAPIData(item, false))
}

// apiStats handles GET /api/apis/stats.
//
// @Summary GET /api/apis/stats
// @ID apis_stats_route_get
// @Tags apis
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=catalog.APIStats}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/apis/stats [get]
func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.apis.Stats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}
