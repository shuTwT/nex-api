package config

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const RedactedValue = "[REDACTED]"

type Config struct {
	Environment string   `mapstructure:"environment"`
	AppURL      string   `mapstructure:"app_url"`
	Server      Server   `mapstructure:"server"`
	Database    Database `mapstructure:"database"`
	Redis       Redis    `mapstructure:"redis"`
	Auth        Auth     `mapstructure:"auth"`
	Upload      Upload   `mapstructure:"upload"`
	Cron        Cron     `mapstructure:"cron"`
	Log         Log      `mapstructure:"log"`
}

type Server struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	MaxBodyBytes      int64         `mapstructure:"max_body_bytes"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
	CORSOrigins       []string      `mapstructure:"cors_origins"`
}

type Database struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type Redis struct {
	URL      string `mapstructure:"url"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	TLS      bool   `mapstructure:"tls"`
}

type Auth struct {
	SessionSecret     string `mapstructure:"session_secret"`
	JWTSecret         string `mapstructure:"jwt_secret"`
	SessionCookieName string `mapstructure:"session_cookie_name"`
}

type Upload struct {
	Directory     string   `mapstructure:"directory"`
	MaxBytes      int64    `mapstructure:"max_bytes"`
	AllowedTypes  []string `mapstructure:"allowed_types"`
	CreateOnStart bool     `mapstructure:"create_on_start"`
}

type Cron struct {
	Enabled bool   `mapstructure:"enabled"`
	Secret  string `mapstructure:"secret"`
}

type Log struct {
	Level     string `mapstructure:"level"`
	Format    string `mapstructure:"format"`
	AddSource bool   `mapstructure:"add_source"`
}

func (c Config) Redacted() map[string]string {
	return map[string]string{
		"environment":         c.Environment,
		"app_url":             c.AppURL,
		"server.host":         c.Server.Host,
		"server.port":         formatInt(c.Server.Port),
		"database.url":        RedactURL(c.Database.URL),
		"redis.url":           RedactURL(c.Redis.URL),
		"redis.password":      RedactSecret(c.Redis.Password),
		"auth.session_secret": RedactSecret(c.Auth.SessionSecret),
		"auth.jwt_secret":     RedactSecret(c.Auth.JWTSecret),
		"cron.secret":         RedactSecret(c.Cron.Secret),
	}
}

func RedactSecret(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}
	return RedactedValue
}

func RedactURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return RedactedValue
	}
	if parsed.User != nil {
		parsed.User = url.User(parsed.User.Username())
	}
	for _, key := range []string{"password", "secret", "token", "key"} {
		values := parsed.Query()
		if values.Has(key) {
			values.Set(key, RedactedValue)
		}
		parsed.RawQuery = values.Encode()
	}
	return parsed.String()
}

func formatInt(value int) string {
	return strconv.Itoa(value)
}
