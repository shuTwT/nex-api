// Package marketplace provides the public marketplace listing service.
package marketplace

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"

	"github.com/shuTwT/nex-api/ent"
	"github.com/shuTwT/nex-api/ent/api"
	"github.com/shuTwT/nex-api/ent/mcpservice"
	servicemcpgateway "github.com/shuTwT/nex-api/internal/service/mcpgateway"
	"github.com/shuTwT/nex-api/internal/service/stats"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type APIView = model.MarketplaceAPIResp
type MCPView = model.MarketplaceMCPResp

// ListOptions filters a marketplace listing.
type ListOptions struct {
	Page     int
	Limit    int
	Search   string
	Category string
	Type     string
}

// Service owns marketplace queries; the handler only adapts HTTP.
type Service struct {
	db       *ent.Client
	stats    *stats.Store
	executor servicemcpgateway.Executor
}

func NewService(db *ent.Client, statStore *stats.Store) (*Service, error) {
	if db == nil || statStore == nil {
		return nil, fmt.Errorf("marketplace: database and stats store are required")
	}
	executor, err := servicemcpgateway.NewProxyExecutor(servicemcpgateway.StdioOptions{})
	if err != nil {
		return nil, fmt.Errorf("marketplace: create MCP tool discovery executor: %w", err)
	}
	return &Service{db: db, stats: statStore, executor: executor}, nil
}

// snapshot reads the global stats snapshot, degrading to an empty snapshot
// when the stats service (Redis) is unavailable so the public marketplace
// stays available.
func (s *Service) snapshot(ctx context.Context) stats.Snapshot {
	snapshot, err := s.stats.Snapshot(ctx)
	if err != nil {
		return stats.Snapshot{}
	}
	return snapshot
}

// ListAPIs returns a page of active marketplace APIs.
func (s *Service) ListAPIs(ctx context.Context, options ListOptions) ([]APIView, int, error) {
	query := s.db.Api.Query().Where(api.IsActive(true))
	if options.Search != "" {
		query = query.Where(api.Or(api.NameContainsFold(options.Search), api.DescriptionContainsFold(options.Search), api.AliasContainsFold(options.Search)))
	}
	if options.Category != "" && options.Category != "all" {
		query = query.Where(api.CategoryId(options.Category))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count marketplace APIs: %w", err)
	}
	items, err := query.WithCategory().Order(api.ByCallCount(sql.OrderDesc())).Offset((options.Page - 1) * options.Limit).Limit(options.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list marketplace APIs: %w", err)
	}
	snapshot := s.snapshot(ctx)
	views := make([]APIView, 0, len(items))
	for _, item := range items {
		views = append(views, s.toAPIView(ctx, item, snapshot, false))
	}
	return views, total, nil
}

// GetAPI returns one active API by ID.
func (s *Service) GetAPI(ctx context.Context, id string) (APIView, error) {
	item, err := s.db.Api.Query().Where(api.ID(id), api.IsActive(true)).WithCategory().Only(ctx)
	if err != nil {
		return APIView{}, err
	}
	return s.toAPIView(ctx, item, s.snapshot(ctx), true), nil
}

// APIStats returns aggregate API counts.
func (s *Service) APIStats(ctx context.Context) (map[string]int64, error) {
	base := s.db.Api.Query().Where(api.IsActive(true))
	total, err := base.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count marketplace APIs: %w", err)
	}
	free, err := s.db.Api.Query().Where(api.IsActive(true), api.Pricing(0)).Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count free marketplace APIs: %w", err)
	}
	items, err := base.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("sum marketplace API calls: %w", err)
	}
	snapshot := s.snapshot(ctx)
	var calls int64
	for _, item := range items {
		calls += canonicalOrDatabase(snapshot.APIs, item.Alias, int64(item.CallCount))
	}
	return map[string]int64{"totalApis": int64(total), "freeApis": int64(free), "paidApis": int64(total - free), "totalCallCount": calls}, nil
}

// ListMCP returns a page of active marketplace MCP services.
func (s *Service) ListMCP(ctx context.Context, options ListOptions) ([]MCPView, int, error) {
	query := s.db.McpService.Query().Where(mcpservice.IsActive(true))
	if options.Search != "" {
		query = query.Where(mcpservice.Or(mcpservice.NameContainsFold(options.Search), mcpservice.IdentifierContainsFold(options.Search), mcpservice.DescriptionContainsFold(options.Search)))
	}
	if options.Category != "" && options.Category != "all" {
		query = query.Where(mcpservice.CategoryId(options.Category))
	}
	if options.Type != "" && options.Type != "all" {
		query = query.Where(mcpservice.Type(options.Type))
	}
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count marketplace MCP services: %w", err)
	}
	items, err := query.WithCategory().Order(mcpservice.ByCallCount(sql.OrderDesc())).Offset((options.Page - 1) * options.Limit).Limit(options.Limit).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list marketplace MCP services: %w", err)
	}
	snapshot := s.snapshot(ctx)
	views := make([]MCPView, 0, len(items))
	for _, item := range items {
		views = append(views, s.toMCPView(ctx, item, snapshot, false))
	}
	return views, total, nil
}

// GetMCP returns one active MCP service by ID.
func (s *Service) GetMCP(ctx context.Context, id string) (MCPView, error) {
	item, err := s.db.McpService.Query().Where(mcpservice.ID(id), mcpservice.IsActive(true)).WithCategory().Only(ctx)
	if err != nil {
		return MCPView{}, err
	}
	return s.toMCPView(ctx, item, s.snapshot(ctx), true), nil
}

// MCPStats returns aggregate MCP service counts.
func (s *Service) MCPStats(ctx context.Context) (map[string]int64, error) {
	items, err := s.db.McpService.Query().Where(mcpservice.IsActive(true)).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list marketplace MCP services: %w", err)
	}
	free := 0
	for _, item := range items {
		if item.Pricing == 0 {
			free++
		}
	}
	snapshot := s.snapshot(ctx)
	var calls int64
	for _, item := range items {
		calls += canonicalOrDatabase(snapshot.MCPs, item.Identifier, int64(item.CallCount))
	}
	total := len(items)
	return map[string]int64{"totalServices": int64(total), "freeServices": int64(free), "paidServices": int64(total - free), "totalCallCount": calls}, nil
}

func (s *Service) toAPIView(ctx context.Context, item *ent.Api, snapshot stats.Snapshot, detail bool) APIView {
	trend, _ := s.stats.APIRequestTrend(ctx, item.Alias, 24)
	return APIView{ID: item.ID, Name: item.Name, Description: item.Description, Alias: item.Alias, Endpoint: item.Endpoint, Method: item.Method, Pricing: item.Pricing, Category: item.Edges.Category.Name, IsFree: item.Pricing == 0, IsActive: detail && item.IsActive, TodayCallCount: sumInt64(trend), UserCount: countAPIUsers(snapshot.UserAPIs, item.Alias), TotalCallCount: canonicalOrDatabase(snapshot.APIs, item.Alias, int64(item.CallCount)), CreatedAt: item.CreatedAt.UTC().Format(timeFormat), UpdatedAt: item.UpdatedAt.UTC().Format(timeFormat)}
}

func (s *Service) toMCPView(ctx context.Context, item *ent.McpService, snapshot stats.Snapshot, detail bool) MCPView {
	trend, _ := s.stats.MCPRequestTrend(ctx, item.Identifier, 24)
	category := "未分类"
	if item.Edges.Category != nil {
		category = item.Edges.Category.Name
	}
	return MCPView{ID: item.ID, Name: item.Name, Identifier: item.Identifier, Category: category, Description: item.Description, Documentation: item.Documentation, Type: item.Type, Pricing: item.Pricing, IsFree: item.Pricing == 0, IsActive: detail && item.IsActive, TodayCallCount: sumInt64(trend), UserCount: countMCPUsers(snapshot.UserMCPs, item.Identifier), TotalCallCount: canonicalOrDatabase(snapshot.MCPs, item.Identifier, int64(item.CallCount)), CreatedAt: item.CreatedAt.UTC().Format(timeFormat), UpdatedAt: item.UpdatedAt.UTC().Format(timeFormat)}
}

const timeFormat = "2006-01-02T15:04:05.000Z"

func canonicalOrDatabase(values map[string]int64, key string, fallback int64) int64 {
	value, ok := values[key]
	if ok {
		return value
	}
	return fallback
}

func countAPIUsers(values map[stats.UserAPIKey]int64, alias string) int {
	count := 0
	for key, value := range values {
		if key.Alias == alias && value > 0 {
			count++
		}
	}
	return count
}

func countMCPUsers(values map[stats.UserMCPKey]int64, identifier string) int {
	count := 0
	for key, value := range values {
		if key.Identifier == identifier && value > 0 {
			count++
		}
	}
	return count
}

func sumInt64(values []int64) int64 {
	var total int64
	for _, value := range values {
		total += value
	}
	return total
}
