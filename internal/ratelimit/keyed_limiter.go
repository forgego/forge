package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64
}

// KeyedLimiter applies independent token buckets to arbitrary keys.
type KeyedLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
	rate     rate.Limit
	burst    int
	cleanup  *time.Ticker
	done     chan struct{}
	stopOnce sync.Once
	idleTTL  time.Duration
}

// NewKeyedLimiter creates a keyed limiter with a burst equal to requests.
func NewKeyedLimiter(requests int, window time.Duration) *KeyedLimiter {
	limiter := &KeyedLimiter{
		limiters: make(map[string]*limiterEntry), burst: requests,
		cleanup: time.NewTicker(5 * time.Minute), done: make(chan struct{}),
		idleTTL: 10 * time.Minute,
	}
	if requests > 0 && window > 0 {
		limiter.rate = rate.Every(window / time.Duration(requests))
	}
	go limiter.cleanupExpired()
	return limiter
}

// Reserve uses one token for key and reports when a denied request may retry.
func (l *KeyedLimiter) Reserve(key string) (bool, time.Duration) {
	now := time.Now()
	reservation := l.getLimiter(key).ReserveN(now, 1)
	if !reservation.OK() {
		return false, time.Second
	}
	retryAfter := reservation.DelayFrom(now)
	if retryAfter <= 0 {
		return true, 0
	}
	reservation.CancelAt(now)
	return false, retryAfter
}

// Allow reports whether an event may happen at the current time for key and the retry duration if denied.
func (l *KeyedLimiter) Allow(key string) (bool, time.Duration) {
	return l.Reserve(key)
}

func (l *KeyedLimiter) getLimiter(key string) *rate.Limiter {
	now := time.Now().UnixNano()
	l.mu.RLock()
	entry := l.limiters[key]
	l.mu.RUnlock()
	if entry != nil {
		entry.lastSeen.Store(now)
		return entry.limiter
	}
	return l.createLimiter(key, now)
}

func (l *KeyedLimiter) createLimiter(key string, now int64) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if entry := l.limiters[key]; entry != nil {
		entry.lastSeen.Store(now)
		return entry.limiter
	}
	entry := &limiterEntry{limiter: rate.NewLimiter(l.rate, l.burst)}
	entry.lastSeen.Store(now)
	l.limiters[key] = entry
	return entry.limiter
}

func (l *KeyedLimiter) cleanupExpired() {
	defer l.cleanup.Stop()
	for {
		select {
		case <-l.cleanup.C:
			l.evictIdle(time.Now())
		case <-l.done:
			return
		}
	}
}

func (l *KeyedLimiter) evictIdle(now time.Time) {
	cutoff := now.Add(-l.idleTTL).UnixNano()
	l.mu.Lock()
	defer l.mu.Unlock()
	for key, entry := range l.limiters {
		if entry.lastSeen.Load() < cutoff {
			delete(l.limiters, key)
		}
	}
}

// Close stops the background cleanup worker.
func (l *KeyedLimiter) Close() {
	l.stopOnce.Do(func() { close(l.done) })
}
