package stats

import (
	"context"
	"fmt"
	"time"
)

func (s *Store) APIRequestTrend(ctx context.Context, alias string, hours int) ([]int64, error) {
	return s.requestTrend(ctx, hours, func(hour time.Time) string {
		return s.matrix.APIRequestTrend(alias, hour)
	})
}

func (s *Store) MCPRequestTrend(ctx context.Context, identifier string, hours int) ([]int64, error) {
	return s.requestTrend(ctx, hours, func(hour time.Time) string {
		return s.matrix.MCPRequestTrend(identifier, hour)
	})
}

func (s *Store) UserAPIRequestTrend(ctx context.Context, userID, alias string, hours int) ([]int64, error) {
	return s.requestTrend(ctx, hours, func(hour time.Time) string {
		return s.matrix.UserAPIRequestTrend(userID, alias, hour)
	})
}

func (s *Store) UserMCPRequestTrend(ctx context.Context, userID, identifier string, hours int) ([]int64, error) {
	return s.requestTrend(ctx, hours, func(hour time.Time) string {
		return s.matrix.UserMCPRequestTrend(userID, identifier, hour)
	})
}

func (s *Store) GlobalRequestTrend(ctx context.Context, hours int) ([]int64, error) {
	return s.requestTrend(ctx, hours, s.matrix.GlobalRequestTrend)
}

func (s *Store) requestTrend(ctx context.Context, hours int, key func(time.Time) string) ([]int64, error) {
	if ctx == nil {
		return nil, fmt.Errorf("stats: context is nil")
	}
	if hours < 1 {
		return []int64{}, nil
	}
	if hours > maxTrendHours {
		hours = maxTrendHours
	}
	now := s.now().UTC().Truncate(time.Hour)
	keys := make([]string, 0, hours)
	for offset := hours - 1; offset >= 0; offset-- {
		keys = append(keys, key(now.Add(-time.Duration(offset)*time.Hour)))
	}
	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("read request trend: %w", err)
	}
	trend := make([]int64, len(values))
	for index, value := range values {
		if parsed, ok := redisValueInt64(value); ok {
			trend[index] = parsed
		}
	}
	return trend, nil
}
