package stats

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

func (s *Store) migratePattern(ctx context.Context, pattern string) error {
	return s.scan(ctx, pattern, func(keys []string) error {
		values, err := s.client.MGet(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("read legacy MCP counters: %w", err)
		}
		_, err = s.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			for i, key := range keys {
				value, ok := redisValueString(values[i])
				if !ok {
					continue
				}
				canonical, ok := s.legacyMCPCanonicalKey(key)
				if ok {
					pipe.SetNX(ctx, canonical, value, 0)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("migrate legacy MCP counters: %w", err)
		}
		return nil
	})
}

func (s *Store) legacyMCPCanonicalKey(key string) (string, bool) {
	if strings.HasPrefix(key, legacyAPIRequests+MCPAliasPrefix) {
		return s.matrix.MCPRequests(strings.TrimPrefix(key, legacyAPIRequests+MCPAliasPrefix)), true
	}
	if !strings.HasPrefix(key, legacyUserRequests) {
		return "", false
	}
	rest := strings.TrimPrefix(key, legacyUserRequests)
	separator := strings.Index(rest, ":"+MCPAliasPrefix)
	if separator < 1 {
		return "", false
	}
	return s.matrix.UserMCPRequests(rest[:separator], rest[separator+len(":"+MCPAliasPrefix):]), true
}

func (s *Store) scan(ctx context.Context, pattern string, visit func([]string) error) error {
	var cursor uint64
	for {
		keys, next, err := s.client.Scan(ctx, cursor, pattern, scanPageSize).Result()
		if err != nil {
			return fmt.Errorf("scan %q: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := visit(keys); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

func (s *Store) assignCanonical(snapshot *Snapshot, key string, value int64) {
	rest := strings.TrimPrefix(key, statsPrefix)
	switch {
	case strings.HasPrefix(rest, "api:"):
		snapshot.APIs[strings.TrimSuffix(strings.TrimPrefix(rest, "api:"), ":requests")] = value
	case strings.HasPrefix(rest, "mcp:"):
		snapshot.MCPs[strings.TrimSuffix(strings.TrimPrefix(rest, "mcp:"), ":requests")] = value
	case strings.HasPrefix(rest, "user:"):
		s.assignUserCanonical(snapshot, strings.TrimPrefix(rest, "user:"), value)
	}
}

func (s *Store) assignUserCanonical(snapshot *Snapshot, rest string, value int64) {
	if separator := strings.Index(rest, ":api:"); separator > 0 {
		alias := strings.TrimSuffix(rest[separator+len(":api:"):], ":requests")
		snapshot.UserAPIs[UserAPIKey{UserID: rest[:separator], Alias: alias}] = value
		return
	}
	if separator := strings.Index(rest, ":mcp:"); separator > 0 {
		identifier := strings.TrimSuffix(rest[separator+len(":mcp:"):], ":requests")
		snapshot.UserMCPs[UserMCPKey{UserID: rest[:separator], Identifier: identifier}] = value
	}
}

func (s *Store) assignLegacy(snapshot *Snapshot, key string, value int64) {
	if strings.HasPrefix(key, legacyAPIRequests) {
		alias := strings.TrimPrefix(key, legacyAPIRequests)
		if isMCPAlias(alias) {
			if _, exists := snapshot.MCPs[mcpIdentifier(alias)]; !exists {
				snapshot.MCPs[mcpIdentifier(alias)] = value
			}
			return
		}
		if _, exists := snapshot.APIs[alias]; !exists {
			snapshot.APIs[alias] = value
		}
		return
	}
	if !strings.HasPrefix(key, legacyUserRequests) {
		return
	}
	rest := strings.TrimPrefix(key, legacyUserRequests)
	separator := strings.Index(rest, ":")
	if separator < 1 {
		return
	}
	userID, alias := rest[:separator], rest[separator+1:]
	if isMCPAlias(alias) {
		mcpKey := UserMCPKey{UserID: userID, Identifier: mcpIdentifier(alias)}
		if _, exists := snapshot.UserMCPs[mcpKey]; !exists {
			snapshot.UserMCPs[mcpKey] = value
		}
		return
	}
	userKey := UserAPIKey{UserID: userID, Alias: alias}
	if _, exists := snapshot.UserAPIs[userKey]; !exists {
		snapshot.UserAPIs[userKey] = value
	}
}

func validateEvent(ctx context.Context, event RequestEvent) error {
	if ctx == nil {
		return fmt.Errorf("stats: context is nil")
	}
	if strings.TrimSpace(event.Alias) == "" {
		return fmt.Errorf("stats: alias is empty")
	}
	if event.Credits < 0 {
		return fmt.Errorf("stats: credits cannot be negative")
	}
	return nil
}

func redisValueString(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case nil:
		return "", false
	default:
		return fmt.Sprint(typed), true
	}
}

func redisValueInt64(value any) (int64, bool) {
	raw, ok := redisValueString(value)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	return parsed, err == nil
}

func redisValueFloat(value any) (float64, bool) {
	raw, ok := redisValueString(value)
	if !ok {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	return parsed, err == nil
}
