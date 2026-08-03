package database

import (
	"testing"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

func TestNewRedisUsesConfiguredPasswordWhenURLHasNone(t *testing.T) {
	client, err := NewRedis(config.Redis{
		URL:      "redis://localhost:6379/7",
		Password: "configured-password",
	})
	if err != nil {
		t.Fatalf("create Redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	options := client.Options()
	if options.Password != "configured-password" {
		t.Fatalf("Redis password = %q, want configured password", options.Password)
	}
	if options.DB != 7 {
		t.Fatalf("Redis database = %d, want 7", options.DB)
	}
}

func TestNewRedisPrefersPasswordInURL(t *testing.T) {
	client, err := NewRedis(config.Redis{
		URL:      "redis://:url-password@localhost:6379/0",
		Password: "configured-password",
	})
	if err != nil {
		t.Fatalf("create Redis client: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	if options := client.Options(); options.Password != "url-password" {
		t.Fatalf("Redis password = %q, want URL password", options.Password)
	}
}
