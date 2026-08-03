package config

import (
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

type FieldViolation struct {
	Field  string
	Reason string
}

type ValidationErrors struct {
	violations []FieldViolation
}

func (e *ValidationErrors) Error() string {
	if len(e.violations) == 0 {
		return "configuration validation failed"
	}
	parts := make([]string, 0, len(e.violations))
	for _, violation := range e.violations {
		parts = append(parts, violation.Field+": "+violation.Reason)
	}
	return "configuration validation failed: " + strings.Join(parts, "; ")
}

func (e *ValidationErrors) Fields() []FieldViolation {
	return append([]FieldViolation(nil), e.violations...)
}

func (e *ValidationErrors) HasField(field string) bool {
	for _, violation := range e.violations {
		if violation.Field == field {
			return true
		}
	}
	return false
}

func Validate(cfg Config) error {
	return validate(cfg, false, nil)
}

func validateLoaded(cfg Config, v *viper.Viper) error {
	return validate(cfg, true, func(key string) bool {
		return hasExternalValue(v, key)
	})
}

func validate(cfg Config, requireExternal bool, externalValue func(string) bool) error {
	violations := make([]FieldViolation, 0, 8)
	add := func(field string, reason string) {
		violations = append(violations, FieldViolation{Field: field, Reason: reason})
	}
	require := func(field string, value string) {
		if strings.TrimSpace(value) == "" {
			add(field, "is required")
		}
		if requireExternal && (externalValue == nil || !externalValue(field)) {
			add(field, "must be provided by environment or config file")
		}
	}

	if cfg.Environment == "production" || cfg.Environment == "prod" {
		require("database.url", cfg.Database.URL)
		require("redis.url", cfg.Redis.URL)
		require("auth.session_secret", cfg.Auth.SessionSecret)
		require("app_url", cfg.AppURL)
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		add("server.port", "must be between 1 and 65535")
	}
	if cfg.Server.ReadTimeout <= 0 {
		add("server.read_timeout", "must be positive")
	}
	if cfg.Server.ReadHeaderTimeout <= 0 {
		add("server.read_header_timeout", "must be positive")
	}
	if cfg.Server.WriteTimeout <= 0 {
		add("server.write_timeout", "must be positive")
	}
	if cfg.Server.IdleTimeout <= 0 {
		add("server.idle_timeout", "must be positive")
	}
	if cfg.Server.ShutdownTimeout <= 0 {
		add("server.shutdown_timeout", "must be positive")
	}
	if cfg.Server.MaxBodyBytes <= 0 {
		add("server.max_body_bytes", "must be positive")
	}
	if cfg.Server.MaxHeaderBytes <= 0 {
		add("server.max_header_bytes", "must be positive")
	}
	if cfg.Database.MaxOpenConns < 1 {
		add("database.max_open_conns", "must be positive")
	}
	if cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		add("database.max_idle_conns", "must be between zero and max_open_conns")
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		add("database.conn_max_lifetime", "must be positive")
	}
	if !validRedisURL(cfg.Redis.URL) {
		add("redis.url", "must be a valid redis:// or rediss:// URL")
	}
	if cfg.Redis.Port < 1 || cfg.Redis.Port > 65535 {
		add("redis.port", "must be between 1 and 65535")
	}
	if cfg.Redis.DB < 0 {
		add("redis.db", "must not be negative")
	}
	if strings.TrimSpace(cfg.AppURL) != "" && !validHTTPURL(cfg.AppURL) {
		add("app_url", "must be a valid http:// or https:// URL")
	}
	if strings.TrimSpace(cfg.Upload.Directory) == "" {
		add("upload.directory", "must not be empty")
	}
	if cfg.Upload.MaxBytes <= 0 {
		add("upload.max_bytes", "must be positive")
	}
	if cfg.Cron.Enabled {
		if strings.TrimSpace(cfg.Cron.Secret) == "" {
			add("cron.secret", "is required when cron is enabled")
		}
		if cfg.Cron.Interval <= 0 {
			add("cron.interval", "must be positive when cron is enabled")
		}
	}
	if !validLogLevel(cfg.Log.Level) {
		add("log.level", "must be debug, info, warn, or error")
	}
	if !validLogFormat(cfg.Log.Format) {
		add("log.format", "must be json or text")
	}
	if len(violations) == 0 {
		return nil
	}
	sort.Slice(violations, func(i, j int) bool { return violations[i].Field < violations[j].Field })
	return &ValidationErrors{violations: violations}
}

func hasExternalValue(v *viper.Viper, key string) bool {
	if v.InConfig(key) && strings.TrimSpace(v.GetString(key)) != "" {
		return true
	}
	for _, binding := range envBindings {
		if binding.key != key {
			continue
		}
		for _, name := range binding.names {
			if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
				return true
			}
		}
	}
	return false
}

func validHTTPURL(raw string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	return err == nil && parsed.IsAbs() && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func validRedisURL(raw string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	return err == nil && parsed.IsAbs() && parsed.Host != "" && (parsed.Scheme == "redis" || parsed.Scheme == "rediss")
}

func validLogLevel(level string) bool {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func validLogFormat(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json", "text":
		return true
	default:
		return false
	}
}
