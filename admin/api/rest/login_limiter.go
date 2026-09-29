package rest

import (
	"context"
	"sync"
	"time"
)

// LoginAttemptStore counts failed admin logins per key (the client IP and
// the username) and locks a key out after too many. The default keeps the
// counts in process memory; stores.LoginAttempts keeps them in the database
// so the lockout holds across restarts and instances.
type LoginAttemptStore interface {
	// Blocked reports whether key is locked out and for how long.
	Blocked(ctx context.Context, key string) (bool, time.Duration, error)
	// Failed records a failed login for key.
	Failed(ctx context.Context, key string) error
	// Succeeded clears the failures of key.
	Succeeded(ctx context.Context, key string) error
}

type loginLimiter struct {
	mu          sync.Mutex
	failures    map[string]*failureWindow
	max         int
	window      time.Duration
	now         func() time.Time
	lastCleanup time.Time
}

type failureWindow struct {
	count   int
	resetAt time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{failures: make(map[string]*failureWindow), max: 5,
		window: 15 * time.Minute, now: time.Now}
}

func (l *loginLimiter) getNow() time.Time {
	if l.now != nil {
		return l.now()
	}
	return time.Now()
}

// blocked reports whether key is locked out and for how long.
func (l *loginLimiter) blocked(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.getNow()
	l.cleanupExpiredLocked(now)
	window, ok := l.failures[key]
	if !ok {
		return false, 0
	}
	if !now.Before(window.resetAt) {
		delete(l.failures, key)
		return false, 0
	}
	if window.count >= l.max {
		return true, window.resetAt.Sub(now)
	}
	return false, 0
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.getNow()
	l.cleanupExpiredLocked(now)
	window, ok := l.failures[key]
	if !ok || !now.Before(window.resetAt) {
		l.failures[key] = &failureWindow{count: 1, resetAt: now.Add(l.window)}
		return
	}
	window.count++
}

func (l *loginLimiter) cleanupExpiredLocked(now time.Time) {
	if !l.lastCleanup.IsZero() && now.Sub(l.lastCleanup) < time.Minute {
		return
	}
	for key, window := range l.failures {
		if !now.Before(window.resetAt) {
			delete(l.failures, key)
		}
	}
	l.lastCleanup = now
}

// success deletes the key.
func (l *loginLimiter) success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, key)
}

// Blocked implements LoginAttemptStore.
func (l *loginLimiter) Blocked(_ context.Context, key string) (bool, time.Duration, error) {
	blocked, remaining := l.blocked(key)
	return blocked, remaining, nil
}

// Failed implements LoginAttemptStore.
func (l *loginLimiter) Failed(_ context.Context, key string) error {
	l.fail(key)
	return nil
}

// Succeeded implements LoginAttemptStore.
func (l *loginLimiter) Succeeded(_ context.Context, key string) error {
	l.success(key)
	return nil
}
