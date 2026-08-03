package auth

import (
	"sync"
	"time"
)

// loginRateLimiter tracks failed login attempts per key (email + client IP).
type loginRateLimiter struct {
	mu      sync.Mutex
	entries map[string]loginAttempt
	limit   int
	window  time.Duration
	clock   Clock
}

type loginAttempt struct {
	startedAt time.Time
	count     int
}

func newLoginRateLimiter(limit int, window time.Duration, clock Clock) *loginRateLimiter {
	return &loginRateLimiter{entries: make(map[string]loginAttempt), limit: limit, window: window, clock: clock}
}

// AllowLogin checks and records an attempt; returns whether it is allowed and
// the retry delay when rate limited.
func (s *Service) AllowLogin(key string) (bool, time.Duration) {
	if s.rateLimiter == nil {
		return true, 0
	}
	return s.rateLimiter.allow(key)
}

// ResetLogin clears the failure counter for key after a successful login.
func (s *Service) ResetLogin(key string) {
	if s.rateLimiter == nil {
		return
	}
	s.rateLimiter.reset(key)
}

func (l *loginRateLimiter) allow(key string) (bool, time.Duration) {
	now := l.clock.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, exists := l.entries[key]
	if !exists || !now.Before(entry.startedAt.Add(l.window)) {
		l.entries[key] = loginAttempt{startedAt: now, count: 1}
		return true, 0
	}
	if entry.count >= l.limit {
		return false, entry.startedAt.Add(l.window).Sub(now)
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}

func (l *loginRateLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, key)
}
