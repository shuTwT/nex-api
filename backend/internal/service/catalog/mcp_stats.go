package catalog

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/backend/ent/mcpservice"
)

func (s *MCPService) Stats(ctx context.Context) (MCPStats, error) {
	total, err := s.db.McpService.Query().Count(ctx)
	if err != nil {
		return MCPStats{}, fmt.Errorf("count MCP services: %w", err)
	}
	active, err := s.db.McpService.Query().Where(mcpservice.IsActive(true)).Count(ctx)
	if err != nil {
		return MCPStats{}, fmt.Errorf("count active MCP services: %w", err)
	}
	inactive, err := s.db.McpService.Query().Where(mcpservice.IsActive(false)).Count(ctx)
	if err != nil {
		return MCPStats{}, fmt.Errorf("count inactive MCP services: %w", err)
	}
	calls, err := s.db.McpService.Query().Aggregate(func(selector *sql.Selector) string {
		return "COALESCE(" + sql.Sum(selector.C(mcpservice.FieldCallCount)) + ", 0)"
	}).Int(ctx)
	if err != nil {
		return MCPStats{}, fmt.Errorf("sum MCP calls: %w", err)
	}
	return MCPStats{TotalServices: total, ActiveServices: active, InactiveServices: inactive, TotalCalls: calls}, nil
}
