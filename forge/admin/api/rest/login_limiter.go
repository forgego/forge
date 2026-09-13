package rest

import (
	"sync"
	"time"
)

type loginLimiter struct {
	mu       sync.Mutex
	failures map[string]*failureWindow // key -> window
	max      int                       // default 5
	window   time.Duration             // default 15 * time.Minute
	now      func() time.Time
}

type failureWindow struct {
	count   int
	resetAt time.Time
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{
		failures: make(map[string]*failureWindow),
		max:      5,
		window:   15 * time.Minute,
		now:      time.Now,
	}
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
	w, ok := l.failures[key]
	if !ok {
		return false, 0
	}

	if !now.Before(w.resetAt) {
		delete(l.failures, key)
		return false, 0
	}

	if w.count >= l.max {
		return true, w.resetAt.Sub(now)
	}

	return false, 0
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.getNow()

	if len(l.failures) > 10000 {
		for k, w := range l.failures {
			if !now.Before(w.resetAt) {
				delete(l.failures, k)
			}
		}
	}

	w, ok := l.failures[key]
	if !ok || !now.Before(w.resetAt) {
		l.failures[key] = &failureWindow{
			count:   1,
			resetAt: now.Add(l.window),
		}
	} else {
		w.count++
	}
}

// success deletes the key
func (l *loginLimiter) success(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.failures, key)
}
