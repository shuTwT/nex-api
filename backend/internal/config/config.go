package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Option func(*loadOptions) error

type loadOptions struct {
	configFile string
}

type envBinding struct {
	key   string
	names []string
}

var envBindings = []envBinding{
	{key: "environment", names: []string{"APP_ENV", "NODE_ENV"}},
	{key: "app_url", names: []string{"APP_URL", "NEXT_PUBLIC_APP_URL"}},
	{key: "server.host", names: []string{"SERVER_HOST", "HOST"}},
	{key: "server.port", names: []string{"SERVER_PORT", "PORT"}},
	{key: "server.read_timeout", names: []string{"SERVER_READ_TIMEOUT", "READ_TIMEOUT"}},
	{key: "server.read_header_timeout", names: []string{"SERVER_READ_HEADER_TIMEOUT", "READ_HEADER_TIMEOUT"}},
	{key: "server.write_timeout", names: []string{"SERVER_WRITE_TIMEOUT", "WRITE_TIMEOUT"}},
	{key: "server.idle_timeout", names: []string{"SERVER_IDLE_TIMEOUT", "IDLE_TIMEOUT"}},
	{key: "server.shutdown_timeout", names: []string{"SERVER_SHUTDOWN_TIMEOUT", "SHUTDOWN_TIMEOUT"}},
	{key: "server.max_body_bytes", names: []string{"SERVER_MAX_BODY_BYTES", "MAX_BODY_BYTES"}},
	{key: "server.max_header_bytes", names: []string{"SERVER_MAX_HEADER_BYTES", "MAX_HEADER_BYTES"}},
	{key: "server.cors_origins", names: []string{"SERVER_CORS_ORIGINS", "CORS_ORIGINS"}},
	{key: "database.url", names: []string{"DATABASE_URL"}},
	{key: "database.max_open_conns", names: []string{"DATABASE_MAX_OPEN_CONNS"}},
	{key: "database.max_idle_conns", names: []string{"DATABASE_MAX_IDLE_CONNS"}},
	{key: "database.conn_max_lifetime", names: []string{"DATABASE_CONN_MAX_LIFETIME"}},
	{key: "redis.url", names: []string{"REDIS_URL"}},
	{key: "redis.host", names: []string{"REDIS_HOST"}},
	{key: "redis.port", names: []string{"REDIS_PORT"}},
	{key: "redis.password", names: []string{"REDIS_PASSWORD"}},
	{key: "redis.db", names: []string{"REDIS_DB"}},
	{key: "redis.tls", names: []string{"REDIS_TLS"}},
	{key: "auth.session_secret", names: []string{"SESSION_SECRET"}},
	{key: "auth.jwt_secret", names: []string{"JWT_SECRET"}},
	{key: "auth.nextauth_secret", names: []string{"NEXTAUTH_SECRET"}},
	{key: "auth.session_cookie_name", names: []string{"SESSION_COOKIE_NAME"}},
	{key: "auth.github_client_id", names: []string{"GITHUB_OAUTH_CLIENT_ID"}},
	{key: "auth.github_client_secret", names: []string{"GITHUB_OAUTH_CLIENT_SECRET"}},
	{key: "auth.sso_client_id", names: []string{"SSO_CLIENT_ID"}},
	{key: "auth.sso_client_secret", names: []string{"SSO_CLIENT_SECRET"}},
	{key: "auth.sso_authorization_url", names: []string{"SSO_AUTHORIZATION_URL"}},
	{key: "auth.sso_token_url", names: []string{"SSO_TOKEN_URL"}},
	{key: "auth.sso_user_info_url", names: []string{"SSO_USER_INFO_URL"}},
	{key: "auth.sso_scope", names: []string{"SSO_SCOPE"}},
	{key: "upload.directory", names: []string{"UPLOAD_DIRECTORY", "UPLOAD_DIR"}},
	{key: "upload.max_bytes", names: []string{"UPLOAD_MAX_BYTES"}},
	{key: "upload.allowed_types", names: []string{"UPLOAD_ALLOWED_TYPES"}},
	{key: "upload.create_on_start", names: []string{"UPLOAD_CREATE_ON_START"}},
	{key: "cron.enabled", names: []string{"CRON_ENABLED"}},
	{key: "cron.secret", names: []string{"CRON_SECRET"}},
	{key: "cron.interval", names: []string{"CRON_INTERVAL"}},
	{key: "log.level", names: []string{"LOG_LEVEL"}},
	{key: "log.format", names: []string{"LOG_FORMAT"}},
	{key: "log.add_source", names: []string{"LOG_ADD_SOURCE"}},
}

type ConfigFileError struct {
	Path  string
	Cause error
}

func (e *ConfigFileError) Error() string {
	return fmt.Sprintf("read config file %q: %v", e.Path, e.Cause)
}

func (e *ConfigFileError) Unwrap() error { return e.Cause }

func WithConfigFile(path string) Option {
	return func(options *loadOptions) error {
		path = strings.TrimSpace(path)
		if path == "" {
			return fmt.Errorf("config file path is empty")
		}
		options.configFile = filepath.Clean(path)
		return nil
	}
}

func LoadFromFile(path string) (Config, error) {
	return Load(WithConfigFile(path))
}

func Load(options ...Option) (Config, error) {
	loadOptions := loadOptions{}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&loadOptions); err != nil {
			return Config{}, fmt.Errorf("configure config loader: %w", err)
		}
	}

	v, err := newViper()
	if err != nil {
		return Config{}, fmt.Errorf("initialize config loader: %w", err)
	}
	if loadOptions.configFile != "" {
		v.SetConfigFile(loadOptions.configFile)
		if filepath.Ext(loadOptions.configFile) == "" {
			v.SetConfigType("yaml")
		}
		if err := v.ReadInConfig(); err != nil {
			return Config{}, &ConfigFileError{Path: loadOptions.configFile, Cause: err}
		}
	}

	var cfg Config
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
	if err := v.Unmarshal(&cfg, viper.DecodeHook(decodeHook)); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	cfg = normalize(cfg)
	if err := validateLoaded(cfg, v); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func newViper() (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetDefault("environment", "development")
	v.SetDefault("app_url", "http://localhost:3000")
	v.SetDefault("server.host", "127.0.0.1")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.read_header_timeout", 5*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 20*time.Second)
	v.SetDefault("server.max_body_bytes", int64(10<<20))
	v.SetDefault("server.max_header_bytes", 1<<20)
	v.SetDefault("server.cors_origins", []string{"http://localhost:3000"})
	v.SetDefault("database.url", "file:./data/dev.db")
	v.SetDefault("database.max_open_conns", 10)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("upload.directory", "data/upload/")
	v.SetDefault("upload.max_bytes", int64(10<<20))
	v.SetDefault("upload.allowed_types", []string{"image/jpeg", "image/png", "image/gif", "image/webp"})
	v.SetDefault("upload.create_on_start", true)
	v.SetDefault("cron.enabled", false)
	v.SetDefault("cron.interval", 5*time.Minute)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "text")
	v.SetDefault("log.add_source", false)

	for _, binding := range envBindings {
		args := append([]string{binding.key}, binding.names...)
		if err := v.BindEnv(args...); err != nil {
			return nil, fmt.Errorf("bind environment %s: %w", binding.key, err)
		}
	}
	return v, nil
}

func normalize(cfg Config) Config {
	cfg.Environment = strings.ToLower(strings.TrimSpace(cfg.Environment))
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Redis.URL == "" {
		host := cfg.Redis.Host
		if host == "" {
			host = "localhost"
		}
		cfg.Redis.URL = "redis://" + net.JoinHostPort(host, strconv.Itoa(cfg.Redis.Port)) + "/" + strconv.Itoa(cfg.Redis.DB)
	}
	if cfg.Upload.Directory == "" {
		cfg.Upload.Directory = "data/upload/"
	}
	return cfg
}
