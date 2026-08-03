package marketplace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/backend/internal/database/ent"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/api"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/mcpservice"
	"github.com/shuTwT/nex-api/backend/internal/runtime"
	"github.com/shuTwT/nex-api/backend/internal/stats"
)

type Handler struct {
	db    *ent.Client
	stats *stats.Store
}

func NewHandler(db *ent.Client, statStore *stats.Store) (*Handler, error) {
	if db == nil || statStore == nil {
		return nil, errors.New("marketplace: database and stats store are required")
	}
	return &Handler{db: db, stats: statStore}, nil
}

// snapshot 读取全局统计快照。统计服务(Redis)不可用时降级为空快照,
// 展示层会回退到数据库中的调用计数,保证公开市场始终可用。
func (h *Handler) snapshot(ctx context.Context) stats.Snapshot {
	snapshot, err := h.stats.Snapshot(ctx)
	if err != nil {
		return stats.Snapshot{}
	}
	return snapshot
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler) error {
	if mux == nil || handler == nil {
		return errors.New("marketplace: mux and handler are required")
	}
	mux.HandleFunc("GET /api/marketplace/apis", handler.listAPIs)
	mux.HandleFunc("GET /api/marketplace/apis/{id}", handler.getAPI)
	mux.HandleFunc("GET /api/marketplace/stats", handler.apiStats)
	mux.HandleFunc("GET /api/marketplace/mcp-services", handler.listMCP)
	mux.HandleFunc("GET /api/marketplace/mcp-services/{id}", handler.getMCP)
	mux.HandleFunc("GET /api/marketplace/mcp-stats", handler.mcpStats)
	return nil
}

func (h *Handler) listAPIs(w http.ResponseWriter, r *http.Request) {
	options, err := parsePage(r, 20, 20)
	if err != nil {
		writeError(w, r, err)
		return
	}
	query := h.db.Api.Query().Where(api.IsActive(true))
	if options.search != "" {
		query = query.Where(api.Or(api.NameContainsFold(options.search), api.DescriptionContainsFold(options.search), api.AliasContainsFold(options.search)))
	}
	if options.category != "" && options.category != "all" {
		query = query.Where(api.CategoryId(options.category))
	}
	total, err := query.Count(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("count marketplace APIs: %w", err))
		return
	}
	items, err := query.WithCategory().Order(api.ByCallCount(sql.OrderDesc())).Offset((options.page - 1) * options.limit).Limit(options.limit).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("list marketplace APIs: %w", err))
		return
	}
	snapshot := h.snapshot(r.Context())
	views := make([]apiView, 0, len(items))
	for _, item := range items {
		views = append(views, h.toAPIView(r.Context(), item, snapshot, false))
	}
	writePaginated(w, views, options.page, options.limit, total)
}

func (h *Handler) getAPI(w http.ResponseWriter, r *http.Request) {
	item, err := h.db.Api.Query().Where(api.ID(r.PathValue("id")), api.IsActive(true)).WithCategory().Only(r.Context())
	if err != nil {
		writeError(w, r, runtime.NewAPIError(http.StatusNotFound, "not_found", "API 不存在", runtime.ErrNotFound))
		return
	}
	snapshot := h.snapshot(r.Context())
	view := h.toAPIView(r.Context(), item, snapshot, true)
	writeData(w, http.StatusOK, view)
}

func (h *Handler) apiStats(w http.ResponseWriter, r *http.Request) {
	base := h.db.Api.Query().Where(api.IsActive(true))
	total, err := base.Count(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("count marketplace APIs: %w", err))
		return
	}
	free, err := h.db.Api.Query().Where(api.IsActive(true), api.Pricing(0)).Count(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("count free marketplace APIs: %w", err))
		return
	}
	items, err := base.All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("sum marketplace API calls: %w", err))
		return
	}
	snapshot := h.snapshot(r.Context())
	var calls int64
	for _, item := range items {
		calls += canonicalOrDatabase(snapshot.APIs, item.Alias, int64(item.CallCount))
	}
	writeData(w, http.StatusOK, map[string]int64{"totalApis": int64(total), "freeApis": int64(free), "paidApis": int64(total - free), "totalCallCount": calls})
}

func (h *Handler) listMCP(w http.ResponseWriter, r *http.Request) {
	options, err := parsePage(r, 20, 20)
	if err != nil {
		writeError(w, r, err)
		return
	}
	query := h.db.McpService.Query().Where(mcpservice.IsActive(true))
	if options.search != "" {
		query = query.Where(mcpservice.Or(mcpservice.NameContainsFold(options.search), mcpservice.IdentifierContainsFold(options.search)))
	}
	if typeFilter := strings.TrimSpace(r.URL.Query().Get("type")); typeFilter != "" && typeFilter != "all" {
		query = query.Where(mcpservice.Type(typeFilter))
	}
	total, err := query.Count(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("count marketplace MCP services: %w", err))
		return
	}
	items, err := query.Order(mcpservice.ByCallCount(sql.OrderDesc())).Offset((options.page - 1) * options.limit).Limit(options.limit).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("list marketplace MCP services: %w", err))
		return
	}
	snapshot := h.snapshot(r.Context())
	views := make([]mcpView, 0, len(items))
	for _, item := range items {
		views = append(views, h.toMCPView(r.Context(), item, snapshot, false))
	}
	writePaginated(w, views, options.page, options.limit, total)
}

func (h *Handler) getMCP(w http.ResponseWriter, r *http.Request) {
	item, err := h.db.McpService.Query().Where(mcpservice.ID(r.PathValue("id")), mcpservice.IsActive(true)).Only(r.Context())
	if err != nil {
		writeError(w, r, runtime.NewAPIError(http.StatusNotFound, "not_found", "MCP 服务不存在", runtime.ErrNotFound))
		return
	}
	snapshot := h.snapshot(r.Context())
	writeData(w, http.StatusOK, h.toMCPView(r.Context(), item, snapshot, true))
}

func (h *Handler) mcpStats(w http.ResponseWriter, r *http.Request) {
	items, err := h.db.McpService.Query().Where(mcpservice.IsActive(true)).All(r.Context())
	if err != nil {
		writeError(w, r, fmt.Errorf("list marketplace MCP services: %w", err))
		return
	}
	free := 0
	for _, item := range items {
		if item.Pricing == 0 {
			free++
		}
	}
	snapshot := h.snapshot(r.Context())
	var calls int64
	for _, item := range items {
		calls += canonicalOrDatabase(snapshot.MCPs, item.Identifier, int64(item.CallCount))
	}
	total := len(items)
	writeData(w, http.StatusOK, map[string]int64{"totalServices": int64(total), "freeServices": int64(free), "paidServices": int64(total - free), "totalCallCount": calls})
}
