package catalog

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"github.com/shuTwT/nex-api/backend/internal/database/ent/api"
)

func (s *APIService) Stats(ctx context.Context) (APIStats, error) {
	total, err := s.db.Api.Query().Count(ctx)
	if err != nil {
		return APIStats{}, fmt.Errorf("count APIs: %w", err)
	}
	active, err := s.db.Api.Query().Where(api.IsActive(true)).Count(ctx)
	if err != nil {
		return APIStats{}, fmt.Errorf("count active APIs: %w", err)
	}
	inactive, err := s.db.Api.Query().Where(api.IsActive(false)).Count(ctx)
	if err != nil {
		return APIStats{}, fmt.Errorf("count inactive APIs: %w", err)
	}
	totalCalls, err := s.db.Api.Query().Aggregate(func(selector *sql.Selector) string {
		return "COALESCE(" + sql.Sum(selector.C(api.FieldCallCount)) + ", 0)"
	}).Int(ctx)
	if err != nil {
		return APIStats{}, fmt.Errorf("sum API calls: %w", err)
	}
	categories, err := s.db.ApiCategory.Query().Count(ctx)
	if err != nil {
		return APIStats{}, fmt.Errorf("count API categories: %w", err)
	}
	return APIStats{TotalAPIs: total, ActiveAPIs: active, InactiveAPIs: inactive, TotalCalls: totalCalls, CategoriesCount: categories}, nil
}
