package marketplace

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	appRuntime "github.com/shuTwT/nex-api/backend/internal/handler/httpkit"
	handlerutils "github.com/shuTwT/nex-api/backend/internal/pkg/utils"
	servicemarketplace "github.com/shuTwT/nex-api/backend/internal/service/marketplace"
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
	mux.Get("/api/marketplace/mcp-services/{id}", handler.getMCP)
	mux.Get("/api/marketplace/mcp-stats", handler.mcpStats)
	return nil
}

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

func (h *Handler) getAPI(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.GetAPI(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(http.StatusNotFound, "not_found", "API 不存在", appRuntime.ErrNotFound))
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.APIStats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}

func (h *Handler) listMCP(w http.ResponseWriter, r *http.Request) {
	options, err := parsePage(r, 20, 20)
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	views, total, err := h.service.ListMCP(r.Context(), servicemarketplace.ListOptions{
		Page: options.page, Limit: options.limit,
		Search: options.search, Type: strings.TrimSpace(r.URL.Query().Get("type")),
	})
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WritePaginated(w, views, options.page, options.limit, total)
}

func (h *Handler) getMCP(w http.ResponseWriter, r *http.Request) {
	view, err := h.service.GetMCP(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		handlerutils.WriteError(w, r, appRuntime.NewAPIError(http.StatusNotFound, "not_found", "MCP 服务不存在", appRuntime.ErrNotFound))
		return
	}
	handlerutils.WriteData(w, http.StatusOK, view)
}

func (h *Handler) mcpStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.MCPStats(r.Context())
	if err != nil {
		handlerutils.WriteError(w, r, err)
		return
	}
	handlerutils.WriteData(w, http.StatusOK, stats)
}
