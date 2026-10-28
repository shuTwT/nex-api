package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/shuTwT/nex-api/internal/infra/database/redistest"
)

// newTestLimiter connects to the development Redis (the test is skipped when it
// is unreachable).
func newTestLimiter(t *testing.T, limit int, window time.Duration) (*loginRateLimiter, *redis.Client) {
	t.Helper()
	client := redistest.New(t)
	return newLoginRateLimiter(client, limit, window), client
}

// testKey returns a limiter key unique to this test run and deletes its counter
// on cleanup, so tests never collide with each other or with the developer's data.
func testKey(t *testing.T, client redis.Cmdable, name string) string {
	t.Helper()
	key := fmt.Sprintf("%s|%s|%d", t.Name(), name, time.Now().UnixNano())
	t.Cleanup(func() { _ = client.Del(context.Background(), redisKey(key)).Err() })
	return key
}

func TestLoginRateLimiter_blocksAfterLimitAndReportsRetryAfter(t *testing.T) {
	limiter, client := newTestLimiter(t, 3, time.Minute)
	ctx := context.Background()
	key := testKey(t, client, "k")
	for attempt := 1; attempt <= 3; attempt++ {
		allowed, _, err := limiter.allow(ctx, key)
		if err != nil || !allowed {
			t.Fatalf("attempt %d: allowed=%v err=%v", attempt, allowed, err)
		}
	}
	allowed, retryAfter, err := limiter.allow(ctx, key)
	if err != nil || allowed {
		t.Fatalf("attempt over limit: allowed=%v err=%v", allowed, err)
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Fatalf("retryAfter = %v, want within (0, 1m]", retryAfter)
	}
}

func TestLoginRateLimiter_keysAreIndependent(t *testing.T) {
	limiter, client := newTestLimiter(t, 1, time.Minute)
	ctx := context.Background()
	first, second := testKey(t, client, "first"), testKey(t, client, "second")
	if allowed, _, _ := limiter.allow(ctx, first); !allowed {
		t.Fatal("first key rejected")
	}
	if allowed, _, _ := limiter.allow(ctx, first); allowed {
		t.Fatal("first key allowed past its limit")
	}
	if allowed, _, _ := limiter.allow(ctx, second); !allowed {
		t.Fatal("second key was affected by the first")
	}
}

// The window is real time: this test takes about 1.2s.
func TestLoginRateLimiter_windowExpiresAndIsNotExtendedByAttempts(t *testing.T) {
	limiter, client := newTestLimiter(t, 1, time.Second)
	ctx := context.Background()
	key := testKey(t, client, "k")
	if allowed, _, err := limiter.allow(ctx, key); err != nil || !allowed {
		t.Fatalf("first attempt: allowed=%v err=%v", allowed, err)
	}
	time.Sleep(600 * time.Millisecond)
	if allowed, _, _ := limiter.allow(ctx, key); allowed {
		t.Fatal("attempt over limit was allowed")
	}
	// 600ms + 600ms is past the original 1s window. Had the rejected attempt at
	// 600ms pushed the expiry out, the counter would still be live here.
	time.Sleep(600 * time.Millisecond)
	if allowed, _, err := limiter.allow(ctx, key); err != nil || !allowed {
		t.Fatalf("after window: allowed=%v err=%v", allowed, err)
	}
}

func TestLoginRateLimiter_resetClearsCounter(t *testing.T) {
	limiter, client := newTestLimiter(t, 1, time.Minute)
	ctx := context.Background()
	key := testKey(t, client, "k")
	limiter.allow(ctx, key)
	if err := limiter.reset(ctx, key); err != nil {
		t.Fatal(err)
	}
	if allowed, _, err := limiter.allow(ctx, key); err != nil || !allowed {
		t.Fatalf("after reset: allowed=%v err=%v", allowed, err)
	}
}

func TestLoginRateLimiter_healsCounterWithoutExpiry(t *testing.T) {
	limiter, client := newTestLimiter(t, 5, time.Minute)
	ctx := context.Background()
	key := testKey(t, client, "k")
	// A counter that lost its TTL (for example a crash between INCR and PEXPIRE).
	if err := client.Set(ctx, redisKey(key), 2, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if allowed, _, err := limiter.allow(ctx, key); err != nil || !allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
	if ttl := client.PTTL(ctx, redisKey(key)).Val(); ttl <= 0 {
		t.Fatalf("counter still has no expiry: pttl=%v", ttl)
	}
}

func TestRedisKey_isFixedSizeAndHidesTheEmail(t *testing.T) {
	short := redisKey("a@example.com|1.1.1.1")
	long := redisKey(strings.Repeat("x", 1<<20) + "@example.com|1.1.1.1")
	if len(short) != len(long) {
		t.Fatalf("key length depends on input: %d vs %d", len(short), len(long))
	}
	if strings.Contains(short, "example.com") {
		t.Fatalf("redis key leaks the email: %s", short)
	}
}

// Needs no running Redis: the client points at a closed port.
func TestLoginRateLimiter_failsClosedOnRedisError(t *testing.T) {
	limiter := newLoginRateLimiter(redistest.Unreachable(t), 5, time.Minute)
	allowed, _, err := limiter.allow(context.Background(), "k")
	if allowed || !errors.Is(err, ErrRateLimiterUnavailable) {
		t.Fatalf("allowed=%v err=%v, want rejected with ErrRateLimiterUnavailable", allowed, err)
	}
}

func TestService_AllowLoginWithoutRedisFailsClosed(t *testing.T) {
	service, _ := newTestService(t, "no-redis", "user")
	allowed, _, err := service.AllowLogin(context.Background(), "k")
	if allowed || !errors.Is(err, ErrRateLimiterUnavailable) {
		t.Fatalf("allowed=%v err=%v, want rejected with ErrRateLimiterUnavailable", allowed, err)
	}
}
