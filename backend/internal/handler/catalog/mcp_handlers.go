package catalog

import (
	"net/http"

	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
)

func (h *Handler) listMCP(w http.ResponseWriter, r *http.Request) {
	options, err := parseMCPListOptions(r)
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	result, err := h.mcp.List(r.Context(), options)
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	data := make([]mcpData, len(result.Items))
	for index, item := range result.Items {
		data[index] = toMCPData(item)
	}
	writePaginated(w, data, options.Page, options.Limit, result.Total)
}

func (h *Handler) createMCP(w http.ResponseWriter, r *http.Request) {
	var request mcpRequest
	if err := decodeJSON(r, &request); err != nil {
		writeCatalogError(w, r, err)
		return
	}
	item, err := h.mcp.Create(r.Context(), servicecatalog.MCPInput{Name: request.Name, Identifier: request.Identifier, Type: request.Type, Command: request.Command, Endpoint: request.Endpoint, EnvVars: request.EnvVars, Pricing: intValue(request.Pricing), IsActive: boolValue(request.IsActive, true)})
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusCreated, toMCPData(item))
}

func (h *Handler) getMCP(w http.ResponseWriter, r *http.Request) {
	item, err := h.mcp.GetByIdentifier(r.Context(), r.PathValue("id"))
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toMCPData(item))
}

func (h *Handler) updateMCP(w http.ResponseWriter, r *http.Request) {
	var request mcpUpdateRequest
	if err := decodeJSON(r, &request); err != nil {
		writeCatalogError(w, r, err)
		return
	}
	item, err := h.mcp.Update(r.Context(), r.PathValue("id"), servicecatalog.MCPUpdateInput{Name: request.Name, Identifier: request.Identifier, Type: request.Type, Command: request.Command, Endpoint: request.Endpoint, EnvVars: request.EnvVars, Pricing: request.Pricing, IsActive: request.IsActive})
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toMCPData(item))
}

func (h *Handler) deleteMCP(w http.ResponseWriter, r *http.Request) {
	if err := h.mcp.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"message": "MCP 服务已删除"})
}

func (h *Handler) toggleMCP(w http.ResponseWriter, r *http.Request) {
	item, err := h.mcp.Toggle(r.Context(), r.PathValue("id"))
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, toMCPData(item))
}

func (h *Handler) mcpStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.mcp.Stats(r.Context())
	if err != nil {
		writeCatalogError(w, r, err)
		return
	}
	writeData(w, http.StatusOK, stats)
}
