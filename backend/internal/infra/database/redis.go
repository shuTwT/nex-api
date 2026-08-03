package database

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shuTwT/nex-api/backend/internal/infra/config"
)

// NewRedis creates the Redis client used by the application. Credentials in
// REDIS_URL take precedence over the separately configured password.
func NewRedis(cfg config.Redis) (*redis.Client, error) {
	options, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	if options.Password == "" {
		options.Password = cfg.Password
	}
	if cfg.TLS {
		options.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	options.DialTimeout = 2 * time.Second
	options.MaxRetries = 1
	return redis.NewClient(options), nil
}
