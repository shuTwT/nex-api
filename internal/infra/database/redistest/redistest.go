// Package redistest connects tests to the development Redis instance.
//
// Connection settings come from the REDIS_URL, REDIS_PASSWORD and REDIS_TLS
// environment variables, falling back to the same names in the repository's
// .env file, then to redis://localhost:6379/0. Tests share that instance with
// the developer, so they must use unique keys and clean up after themselves.
package redistest

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/subosito/gotenv"

	"github.com/shuTwT/nex-api/internal/infra/config"
	"github.com/shuTwT/nex-api/internal/infra/database"
)

// New returns a client for the development Redis. The test is skipped when the
// server cannot be reached, so `go test ./...` still passes without Redis.
func New(t testing.TB) *redis.Client {
	t.Helper()
	client, err := database.NewRedis(loadConfig())
	if err != nil {
		t.Fatalf("create development Redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("development Redis is not reachable (%v); start it or set REDIS_URL", err)
	}
	return client
}

// Unreachable returns a client pointing at a closed local port, for tests of
// behavior when Redis is down. It needs no running Redis.
func Unreachable(t testing.TB) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func loadConfig() config.Redis {
	fileValues := map[string]string{}
	if path := findDotEnv(); path != "" {
		if values, err := gotenv.Read(path); err == nil {
			fileValues = values
		}
	}
	get := func(name string) string {
		if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
			return value
		}
		return fileValues[name]
	}
	cfg := config.Redis{
		URL:      strings.TrimSpace(get("REDIS_URL")),
		Password: get("REDIS_PASSWORD"),
		TLS:      strings.EqualFold(strings.TrimSpace(get("REDIS_TLS")), "true"),
	}
	if cfg.URL == "" {
		cfg.URL = "redis://localhost:6379/0"
	}
	return cfg
}

// findDotEnv walks up from the test's working directory to the module root
// and returns the path of its .env file, or "" when there is none.
func findDotEnv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			candidate := filepath.Join(dir, ".env")
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
