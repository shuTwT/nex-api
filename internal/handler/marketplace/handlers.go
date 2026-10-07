package marketplace

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/internal/pkg/utils"
	servicemarketplace "github.com/shuTwT/nex-api/internal/service/marketplace"
)

type Handler struct{ service *servicemarketplace.Service }

func NewHandler(service *servicemarketplace.Service) (*Handler, error) {
	if service == nil {
		return nil, errors.New("marketplace: service is required")
	}
	return &Handler{service: service}, nil
}

func RegisterRoutes(mux chi.Router, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("marketplace: mux and handler are required")
	}
	mux.Get("/api/marketplace/apis", handler.listAPIs)
	mux.Get("/api/marketplace/apis/{id}", handler.getAPI)
	mux.Get("/api/marketplace/stats", handler.apiStats)
	mux.Get("/api/marketplace/mcp-services", handler.listMCP)
	mux.Get("/api/marketplace/mcp-services/{id}/tools", handler.listMCPTools)
	mux.Get("/api/marketplace/mcp-services/{id}", handler.getMCP)
	mux.Get("/api/marketplace/mcp-stats", handler.mcpStats)
	return nil
}

// listAPIs handles GET /api/marketplace/apis.
//
// @Summary GET /api/marketplace/apis
// @ID marketplace_apis_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/apis [get]
func (h *Handler) listAPIs(w http.ResponseWriter, r *http.Request) {
	options, err := parsePage(r, 20, 20)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	views, total, err := h.service.ListAPIs(r.Context(), servicemarketplace.ListOptions{
		Page: options.page, Limit: options.limit,
		Search: options.search, Category: options.category,
	})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, options.page, options.limit, total)
}

// getAPI handles GET /api/marketplace/apis/{id}.
//
// @Summary GET /api/marketplace/apis/{id}
// @ID marketplace_apis_id_route_get
// @Tags marketplace
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/apis/{id} [get]
func (h *Handler) getAPI(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.GetAPI(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(http.StatusNotFound, "not_found", "API 不存在", appRuntime.ErrNotFound))
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// apiStats handles GET /api/marketplace/stats.
//
// @Summary GET /api/marketplace/stats
// @ID marketplace_stats_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/stats [get]
func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.APIStats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}

// listMCP handles GET /api/marketplace/mcp-services.
//
// @Summary GET /api/marketplace/mcp-services
// @ID marketplace_mcp_services_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/mcp-services [get]
func (h *Handler) listMCP(w http.ResponseWriter, r *http.Request) {
	options, err := parsePage(r, 20, 20)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	views, total, err := h.service.ListMCP(r.Context(), servicemarketplace.ListOptions{
		Page: options.page, Limit: options.limit,
		Search: options.search, Category: strings.TrimSpace(r.URL.Query().Get("category")), Type: strings.TrimSpace(r.URL.Query().Get("type")),
	})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, options.page, options.limit, total)
}

// getMCP handles GET /api/marketplace/mcp-services/{id}.
//
// @Summary GET /api/marketplace/mcp-services/{id}
// @ID marketplace_mcp_services_id_route_get
// @Tags marketplace
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 404 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/mcp-services/{id} [get]
func (h *Handler) getMCP(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.GetMCP(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(http.StatusNotFound, "not_found", "MCP 服务不存在", appRuntime.ErrNotFound))
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

// listMCPTools handles GET /api/marketplace/mcp-services/{id}/tools.
//
// @Summary GET /api/marketplace/mcp-services/{id}/tools
// @ID marketplace_mcp_services_id_tools_route_get
// @Tags marketplace
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 404 {object} main.SwaggerEnvelope
// @Failure 502 {object} main.SwaggerEnvelope
// @Router /api/marketplace/mcp-services/{id}/tools [get]
func (h *Handler) listMCPTools(w http.ResponseWriter, r *http.Request) {
	tools, err := h.service.ListMCPTools(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(http.StatusBadGateway, "mcp_tools_unavailable", "暂时无法获取 MCP 工具", err))
		return
	}
	handlerutils.WriteData(w, http.StatusOK, tools)
}

// mcpStats handles GET /api/marketplace/mcp-stats.
//
// @Summary GET /api/marketplace/mcp-stats
// @ID marketplace_mcp_stats_route_get
// @Tags marketplace
// @Produce json
// @Success 200 {object} main.SwaggerEnvelope
// @Failure 400 {object} main.SwaggerEnvelope
// @Failure 401 {object} main.SwaggerEnvelope
// @Failure 500 {object} main.SwaggerEnvelope
// @Router /api/marketplace/mcp-stats [get]
func (h *Handler) mcpStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.MCPStats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}
