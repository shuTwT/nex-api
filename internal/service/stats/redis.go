package stats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	scanPageSize   int64 = 100
	usageRetention       = 24 * time.Hour * 31
	trendRetention       = 24 * time.Hour * 31
	maxTrendHours        = 168
)

var ErrNilRedis = errors.New("stats: redis client is nil")

type Store struct {
	client redis.UniversalClient
	matrix KeyMatrix
	now    func() time.Time
}

type Snapshot struct {
	GlobalRequests int64
	APIs           map[string]int64
	MCPs           map[string]int64
	UserAPIs       map[UserAPIKey]int64
	UserMCPs       map[UserMCPKey]int64
}

func NewStore(client redis.UniversalClient) (*Store, error) {
	if client == nil {
		return nil, ErrNilRedis
	}
	return &Store{client: client, matrix: NewKeyMatrix(), now: time.Now}, nil
}

func (s *Store) Increment(ctx context.Context, event RequestEvent) error {
	if err := validateEvent(ctx, event); err != nil {
		return err
	}
	if isMCPAlias(event.Alias) {
		if err := s.MigrateLegacyMCPAliases(ctx); err != nil {
			return fmt.Errorf("migrate legacy MCP aliases: %w", err)
		}
	}

	hour := event.At
	if hour.IsZero() {
		hour = s.now()
	}
	hour = hour.UTC().Truncate(time.Hour)
	keys := s.requestKeys(event)
	usageKeys := s.usageKeys(event, hour)
	trendKeys := s.trendKeys(event, hour)
	_, err := s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			pipe.Incr(ctx, key)
		}
		for _, key := range trendKeys {
			pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, trendRetention)
		}
		if event.Credits > 0 {
			for _, key := range usageKeys {
				pipe.IncrByFloat(ctx, key, event.Credits)
				pipe.Expire(ctx, key, usageRetention)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("increment stats: %w", err)
	}
	return nil
}

func (s *Store) Snapshot(ctx context.Context) (Snapshot, error) {
	if ctx == nil {
		return Snapshot{}, errors.New("stats: context is nil")
	}
	if err := s.MigrateLegacyMCPAliases(ctx); err != nil {
		return Snapshot{}, fmt.Errorf("migrate legacy MCP aliases: %w", err)
	}
	snapshot := Snapshot{
		APIs:     make(map[string]int64),
		MCPs:     make(map[string]int64),
		UserAPIs: make(map[UserAPIKey]int64),
		UserMCPs: make(map[UserMCPKey]int64),
	}
	global, err := s.readGlobal(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read global requests: %w", err)
	}
	snapshot.GlobalRequests = global
	if err := s.readCanonical(ctx, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("read canonical counters: %w", err)
	}
	if err := s.readLegacy(ctx, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("read legacy counters: %w", err)
	}
	return snapshot, nil
}

func (s *Store) GlobalRequestCount(ctx context.Context) (int64, error) {
	return s.readGlobal(ctx)
}

func (s *Store) APIRequestCount(ctx context.Context, alias string) (int64, error) {
	return s.readCounter(ctx, s.matrix.APIRequests(alias), legacyAPIKey(alias))
}

func (s *Store) UserAPIRequestCount(ctx context.Context, userID, alias string) (int64, error) {
	return s.readCounter(ctx, s.matrix.UserAPIRequests(userID, alias), legacyUserAPIKey(userID, alias))
}

func (s *Store) HourlyUsageTrend(ctx context.Context, userID string, hours int) ([]float64, error) {
	if ctx == nil {
		return nil, errors.New("stats: context is nil")
	}
	if hours < 1 {
		return []float64{}, nil
	}
	if hours > maxTrendHours {
		hours = maxTrendHours
	}
	now := s.now().UTC().Truncate(time.Hour)
	keys := make([]string, 0, hours)
	legacyKeys := make([]string, 0, hours)
	for offset := hours - 1; offset >= 0; offset-- {
		hour := now.Add(-time.Duration(offset) * time.Hour)
		if userID == "" {
			keys = append(keys, s.matrix.GlobalCredits(hour))
			legacyKeys = append(legacyKeys, "global:usage:hourly:"+hourKey(hour))
			continue
		}
		keys = append(keys, s.matrix.UserCredits(userID, hour))
		legacyKeys = append(legacyKeys, "user:usage:hourly:"+userID+":"+hourKey(hour))
	}
	values, err := s.client.MGet(ctx, append(keys, legacyKeys...)...).Result()
	if err != nil {
		return nil, fmt.Errorf("read hourly usage: %w", err)
	}
	trend := make([]float64, hours)
	for i := range trend {
		if value, ok := redisValueFloat(values[i]); ok {
			trend[i] = value
			continue
		}
		if value, ok := redisValueFloat(values[i+hours]); ok {
			trend[i] = value
		}
	}
	return trend, nil
}

func (s *Store) MigrateLegacyMCPAliases(ctx context.Context) error {
	if ctx == nil {
		return errors.New("stats: context is nil")
	}
	marker := statsPrefix + "migration:mcp-aliases"
	marked, err := s.client.Exists(ctx, marker).Result()
	if err != nil {
		return fmt.Errorf("check MCP migration marker: %w", err)
	}
	if marked > 0 {
		return nil
	}
	if err := s.migratePattern(ctx, legacyAPIRequests+MCPAliasPrefix+"*"); err != nil {
		return err
	}
	if err := s.migratePattern(ctx, legacyUserRequests+"*:"+MCPAliasPrefix+"*"); err != nil {
		return err
	}
	if err := s.client.Set(ctx, marker, "1", 0).Err(); err != nil {
		return fmt.Errorf("write MCP migration marker: %w", err)
	}
	return nil
}

func (s *Store) readGlobal(ctx context.Context) (int64, error) {
	return s.readCounter(ctx, s.matrix.GlobalRequests(), legacyGlobalRequests)
}

func (s *Store) readCounter(ctx context.Context, canonical, legacy string) (int64, error) {
	values, err := s.client.MGet(ctx, canonical, legacy).Result()
	if err != nil {
		return 0, err
	}
	for _, value := range values {
		if parsed, ok := redisValueInt64(value); ok && parsed > 0 {
			return parsed, nil
		}
	}
	return 0, nil
}

func (s *Store) readCanonical(ctx context.Context, snapshot *Snapshot) error {
	return s.scan(ctx, statsPrefix+"*:requests", func(keys []string) error {
		values, err := s.client.MGet(ctx, keys...).Result()
		if err != nil {
			return err
		}
		for i, key := range keys {
			value, ok := redisValueInt64(values[i])
			if !ok {
				continue
			}
			s.assignCanonical(snapshot, key, value)
		}
		return nil
	})
}

func (s *Store) readLegacy(ctx context.Context, snapshot *Snapshot) error {
	if err := s.readLegacyPattern(ctx, legacyAPIRequests+"*", snapshot); err != nil {
		return err
	}
	return s.readLegacyPattern(ctx, legacyUserRequests+"*", snapshot)
}

func (s *Store) readLegacyPattern(ctx context.Context, pattern string, snapshot *Snapshot) error {
	return s.scan(ctx, pattern, func(keys []string) error {
		values, err := s.client.MGet(ctx, keys...).Result()
		if err != nil {
			return err
		}
		for i, key := range keys {
			value, ok := redisValueInt64(values[i])
			if !ok {
				continue
			}
			s.assignLegacy(snapshot, key, value)
		}
		return nil
	})
}
