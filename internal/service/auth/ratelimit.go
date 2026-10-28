package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrRateLimiterUnavailable is returned when the login rate limiter cannot
// reach a decision (Redis down, timeout, malformed reply, or no limiter
// configured). Callers must reject the login attempt: the limiter fails closed.
var ErrRateLimiterUnavailable = errors.New("auth: login rate limiter unavailable")

const (
	loginRateLimitKeyPrefix = "auth:login-rl:"
	loginRateLimitTimeout   = 2 * time.Second
)

// loginRateLimitScript atomically counts one attempt and returns {count, ttl_ms}.
// The TTL is only set when the counter is created, so the window is fixed and
// not extended by further attempts. The PTTL guard heals a counter that lost
// its expiry (for example after a crash between INCR and PEXPIRE).
var loginRateLimitScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
	redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
if ttl < 0 then
	redis.call('PEXPIRE', KEYS[1], ARGV[1])
	ttl = tonumber(ARGV[1])
end
return {count, ttl}
`)

// loginRateLimiter counts login attempts per key (email + client IP) in Redis.
// Counters expire on their own, so nothing needs sweeping, and every instance
// sharing the same Redis shares the same counters.
type loginRateLimiter struct {
	client redis.Cmdable
	limit  int
	window time.Duration
}

func newLoginRateLimiter(client redis.Cmdable, limit int, window time.Duration) *loginRateLimiter {
	if window < time.Millisecond {
		window = time.Millisecond
	}
	return &loginRateLimiter{client: client, limit: limit, window: window}
}

// AllowLogin records an attempt for key and reports whether it is allowed. When
// it is not, retryAfter is the time left in the current window. Any error means
// no decision could be made and the attempt must be rejected.
func (s *Service) AllowLogin(ctx context.Context, key string) (bool, time.Duration, error) {
	if s.rateLimiter == nil {
		return false, 0, fmt.Errorf("%w: not configured", ErrRateLimiterUnavailable)
	}
	return s.rateLimiter.allow(ctx, key)
}

// ResetLogin clears the attempt counter for key after a successful login.
func (s *Service) ResetLogin(ctx context.Context, key string) error {
	if s.rateLimiter == nil {
		return nil
	}
	return s.rateLimiter.reset(ctx, key)
}

func (l *loginRateLimiter) allow(ctx context.Context, key string) (bool, time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, loginRateLimitTimeout)
	defer cancel()
	reply, err := loginRateLimitScript.Run(ctx, l.client, []string{redisKey(key)}, l.window.Milliseconds()).Slice()
	if err != nil {
		return false, 0, fmt.Errorf("%w: %w", ErrRateLimiterUnavailable, err)
	}
	if len(reply) != 2 {
		return false, 0, fmt.Errorf("%w: unexpected reply length %d", ErrRateLimiterUnavailable, len(reply))
	}
	count, countOK := reply[0].(int64)
	ttl, ttlOK := reply[1].(int64)
	if !countOK || !ttlOK {
		return false, 0, fmt.Errorf("%w: unexpected reply types", ErrRateLimiterUnavailable)
	}
	if count > int64(l.limit) {
		return false, time.Duration(ttl) * time.Millisecond, nil
	}
	return true, 0, nil
}

func (l *loginRateLimiter) reset(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, loginRateLimitTimeout)
	defer cancel()
	if err := l.client.Del(ctx, redisKey(key)).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRateLimiterUnavailable, err)
	}
	return nil
}

// redisKey hashes the caller's key so the Redis key has a fixed size no matter
// how long the submitted email is, and so addresses are not stored in Redis.
func redisKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return loginRateLimitKeyPrefix + hex.EncodeToString(sum[:])
}
