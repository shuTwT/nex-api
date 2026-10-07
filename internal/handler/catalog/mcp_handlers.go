package catalog

import (
	"net/http"

	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
)

// listMCP handles GET /api/mcp-services.
//
// @Summary GET /api/mcp-services
// @ID mcp_services_route_get
// @Tags mcp-services
// @Produce json
// @Security SessionCookie
// @Param page query integer false "page"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param type query string false "type"
// @Param category query string false "category"
// @Param status query string false "status"
// @Success 200 {object} main.SwaggerEnvelope{data=[]model.CatalogMCPDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services [get]
func (h *Handler) listMCP(w http.ResponseWriter, r *http.Request) {
	options, err := parseMCPListOptions(r)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	result, err := h.mcp.List(r.Context(), options)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	data := make([]mcpData, len(result.Items))
	for index, item := range result.Items {
		data[index] = toMCPData(item)
	}
	handlerutils.WritePaginated(w, data, options.Page, options.Limit, result.Total)
}

// createMCP handles POST /api/mcp-services.
//
// @Summary POST /api/mcp-services
// @ID mcp_services_route_post
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param body body model.CatalogMCPCreateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogMCPDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services [post]
func (h *Handler) createMCP(w http.ResponseWriter, r *http.Request) {
	var request mcpRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.mcp.Create(r.Context(), servicecatalog.MCPInput{Name: request.Name, Identifier: request.Identifier, CategoryID: request.CategoryID, Description: request.Description, Documentation: request.Documentation, Type: request.Type, Command: request.Command, Endpoint: request.Endpoint, EnvVars: request.EnvVars, Pricing: intValue(request.Pricing), IsActive: boolValue(request.IsActive, true)})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusCreated, toMCPData(item))
}

// getMCP handles GET /api/mcp-services/{id}.
//
// @Summary GET /api/mcp-services/{id}
// @ID mcp_services_id_route_get
// @Tags catalog
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogMCPDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 403 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services/{id} [get]
func (h *Handler) getMCP(w http.ResponseWriter, r *http.Request) {
	item, err := h.mcp.GetByIdentifier(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toMCPData(item))
}

// updateMCP handles PUT /api/mcp-services/{id}.
//
// @Summary PUT /api/mcp-services/{id}
// @ID mcp_services_id_route_put
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param body body model.CatalogMCPUpdateReq true "request payload"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogMCPDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services/{id} [put]
func (h *Handler) updateMCP(w http.ResponseWriter, r *http.Request) {
	var request mcpUpdateRequest
	if err := handlerutils.DecodeJSON(r, &request); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	item, err := h.mcp.Update(r.Context(), r.PathValue("id"), servicecatalog.MCPUpdateInput{Name: request.Name, Identifier: request.Identifier, CategoryID: request.CategoryID, Description: request.Description, Documentation: request.Documentation, Type: request.Type, Command: request.Command, Endpoint: request.Endpoint, EnvVars: request.EnvVars, Pricing: request.Pricing, IsActive: request.IsActive})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toMCPData(item))
}

// deleteMCP handles DELETE /api/mcp-services/{id}.
//
// @Summary DELETE /api/mcp-services/{id}
// @ID mcp_services_id_route_delete
// @Tags mcp-services
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=map[string]string}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services/{id} [delete]
func (h *Handler) deleteMCP(w http.ResponseWriter, r *http.Request) {
	if err := h.mcp.Delete(r.Context(), r.PathValue("id")); err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, map[string]string{"message": "MCP 服务已删除"})
}

// toggleMCP handles PUT /api/mcp-services/{id}/toggle.
//
// @Summary PUT /api/mcp-services/{id}/toggle
// @ID mcp_services_id_toggle_route_put
// @Tags mcp-services
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=model.CatalogMCPDTO}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services/{id}/toggle [put]
func (h *Handler) toggleMCP(w http.ResponseWriter, r *http.Request) {
	item, err := h.mcp.Toggle(r.Context(), r.PathValue("id"))
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, toMCPData(item))
}

// mcpStats handles GET /api/mcp-services/stats.
//
// @Summary GET /api/mcp-services/stats
// @ID mcp_services_stats_route_get
// @Tags mcp-services
// @Produce json
// @Security SessionCookie
// @Success 200 {object} main.SwaggerEnvelope{data=catalog.MCPStats}
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/mcp-services/stats [get]
func (h *Handler) mcpStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.mcp.Stats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}
