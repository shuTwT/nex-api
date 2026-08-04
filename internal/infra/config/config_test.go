package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_env_overrides_yaml_and_defaults(t *testing.T) {
	// Given
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	configYAML := []byte("server:\n  port: 9000\ndatabase:\n  url: file:./from-yaml.db\n")
	if err := os.WriteFile(configPath, configYAML, 0o600); err != nil {
		t.Fatalf("write config fixture: %v", err)
	}
	t.Setenv("SERVER_PORT", "8123")

	// When
	cfg, err := Load(WithConfigFile(configPath))

	// Then
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Server.Port != 8123 {
		t.Fatalf("expected environment port 8123, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "file:./from-yaml.db" {
		t.Fatalf("expected YAML database URL, got %q", cfg.Database.URL)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Fatalf("expected development default host, got %q", cfg.Server.Host)
	}
}

func TestLoad_production_requires_external_secrets_and_urls(t *testing.T) {
	// Given
	t.Setenv("APP_ENV", "production")

	// When
	_, err := Load()

	// Then
	if err == nil {
		t.Fatal("expected production validation error")
	}
	var validationErr *ValidationErrors
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}
	if !validationErr.HasField("database.url") {
		t.Fatal("expected database.url validation error")
	}
	if !validationErr.HasField("redis.url") {
		t.Fatal("expected redis.url validation error")
	}
	if !validationErr.HasField("auth.session_secret") {
		t.Fatal("expected auth.session_secret validation error")
	}
	if !validationErr.HasField("app_url") {
		t.Fatal("expected app_url validation error")
	}
}

func TestLoad_readsDotEnvWithoutOverridingProcessEnvironment(t *testing.T) {
	workspace := t.TempDir()
	if err := os.WriteFile(filepath.Join(workspace, ".env"), []byte("REDIS_URL=redis://localhost:6379/7\nREDIS_PASSWORD=from-file\n"), 0o600); err != nil {
		t.Fatalf("write .env fixture: %v", err)
	}
	previousURL, hadURL := os.LookupEnv("REDIS_URL")
	if err := os.Unsetenv("REDIS_URL"); err != nil {
		t.Fatalf("unset Redis URL: %v", err)
	}
	t.Cleanup(func() {
		if hadURL {
			_ = os.Setenv("REDIS_URL", previousURL)
			return
		}
		_ = os.Unsetenv("REDIS_URL")
	})
	t.Setenv("REDIS_PASSWORD", "from-process")

	cfg, err := Load(WithDotEnvFile(filepath.Join(workspace, ".env")))
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	if cfg.Redis.URL != "redis://localhost:6379/7" {
		t.Fatalf("Redis URL = %q, want value from .env", cfg.Redis.URL)
	}
	if cfg.Redis.Password != "from-process" {
		t.Fatalf("Redis password = %q, want process value", cfg.Redis.Password)
	}
}

func TestLoad_cronEndpointDoesNotReadInterval(t *testing.T) {
	// CRON_INTERVAL used to bootstrap the stats_sync scheduled job. Scheduling
	// is now configured by ScheduledJob, so this legacy environment variable
	// must not affect configuration loading or endpoint validation.
	t.Setenv("CRON_ENABLED", "true")
	t.Setenv("CRON_SECRET", "cron-secret")
	t.Setenv("CRON_INTERVAL", "not-a-duration")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Cron.Enabled || cfg.Cron.Secret != "cron-secret" {
		t.Fatalf("cron endpoint config = %#v, want enabled endpoint with configured secret", cfg.Cron)
	}
}

func TestConfig_Redacted_hides_credentials(t *testing.T) {
	// Given
	cfg := Config{
		Database: Database{URL: "postgres://user:password@example.com/app"},
		Redis: Redis{
			URL:      "rediss://:redis-password@example.com:6380/0",
			Password: "redis-password",
		},
		Auth: Auth{SessionSecret: "session-secret"},
	}

	// When
	redacted := cfg.Redacted()

	// Then
	if redacted["database.url"] == cfg.Database.URL {
		t.Fatal("database credentials were not redacted")
	}
	if redacted["redis.password"] != RedactedValue {
		t.Fatalf("expected redacted Redis password, got %q", redacted["redis.password"])
	}
	if redacted["auth.session_secret"] != RedactedValue {
		t.Fatalf("expected redacted session secret, got %q", redacted["auth.session_secret"])
	}
}
