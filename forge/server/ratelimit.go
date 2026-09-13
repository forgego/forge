package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/forgego/forge/netutil"
	"golang.org/x/time/rate"
)

// RateLimiter is the interface for rate limiters
type RateLimiter interface {
	Allow(ctx context.Context) bool
	Wait(ctx context.Context) error
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen atomic.Int64 // unix nanos
}

// rateLimitStore stores rate limiters for different keys
type rateLimitStore struct {
	mu       sync.RWMutex
	limiters map[string]*limiterEntry
	rate     rate.Limit
	burst    int
	cleanup  *time.Ticker
	done     chan struct{}
	stopOnce sync.Once
	idleTTL  time.Duration
}

// newRateLimitStore creates a new rate limit store
func newRateLimitStore(r rate.Limit, burst int) *rateLimitStore {
	store := &rateLimitStore{
		limiters: make(map[string]*limiterEntry),
		rate:     r,
		burst:    burst,
		cleanup:  time.NewTicker(5 * time.Minute),
		done:     make(chan struct{}),
		idleTTL:  10 * time.Minute,
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// getLimiter gets or creates a limiter for a key
func (s *rateLimitStore) getLimiter(key string) *rate.Limiter {
	now := time.Now().UnixNano()

	s.mu.RLock()
	entry, exists := s.limiters[key]
	if exists {
		entry.lastSeen.Store(now)
		s.mu.RUnlock()
		return entry.limiter
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists = s.limiters[key]; exists {
		entry.lastSeen.Store(now)
		return entry.limiter
	}

	entry = &limiterEntry{
		limiter: rate.NewLimiter(s.rate, s.burst),
	}
	entry.lastSeen.Store(now)
	s.limiters[key] = entry
	return entry.limiter
}

// evictIdle removes limiters that have been idle longer than idleTTL
func (s *rateLimitStore) evictIdle(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := now.Add(-s.idleTTL).UnixNano()
	for k, e := range s.limiters {
		if e.lastSeen.Load() < cutoff {
			delete(s.limiters, k)
		}
	}
}

// cleanupExpired periodically cleans up old limiters
func (s *rateLimitStore) cleanupExpired() {
	defer s.cleanup.Stop()
	for {
		select {
		case <-s.cleanup.C:
			s.evictIdle(time.Now())
		case <-s.done:
			return
		}
	}
}

// stop stops the cleanup ticker
func (s *rateLimitStore) stop() {
	s.stopOnce.Do(func() {
		close(s.done)
	})
}

// global stores for different rate limit types
var (
	ipRateLimitStores   = make(map[string]*rateLimitStore)
	userRateLimitStores = make(map[string]*rateLimitStore)
	rateLimitMu         sync.RWMutex
)

// getIPRateLimitStore gets or creates a rate limit store for IP-based limiting
func getIPRateLimitStore(requests int, window time.Duration) *rateLimitStore {
	key := rateLimitKey(requests, window)
	rateLimitMu.RLock()
	store, exists := ipRateLimitStores[key]
	rateLimitMu.RUnlock()

	if exists {
		return store
	}

	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	// Double-check
	if store, exists := ipRateLimitStores[key]; exists {
		return store
	}

	r := rate.Limit(float64(requests) / window.Seconds())
	store = newRateLimitStore(r, requests)
	ipRateLimitStores[key] = store
	return store
}

// getUserRateLimitStore gets or creates a rate limit store for user-based limiting
func getUserRateLimitStore(requests int, window time.Duration) *rateLimitStore {
	key := rateLimitKey(requests, window)
	rateLimitMu.RLock()
	store, exists := userRateLimitStores[key]
	rateLimitMu.RUnlock()

	if exists {
		return store
	}

	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	// Double-check
	if store, exists := userRateLimitStores[key]; exists {
		return store
	}

	r := rate.Limit(float64(requests) / window.Seconds())
	store = newRateLimitStore(r, requests)
	userRateLimitStores[key] = store
	return store
}

// rateLimitKey generates a key for rate limit store lookup
func rateLimitKey(requests int, window time.Duration) string {
	return fmt.Sprintf("%d-%s", requests, window.String())
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	return netutil.ClientIP(r, netutil.TrustedProxies())
}

// RateLimitByIP creates a middleware that rate limits by IP address
func RateLimitByIP(requests int, window time.Duration) Middleware {
	store := getIPRateLimitStore(requests, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			limiter := store.getLimiter(ip)

			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUser creates a middleware that rate limits by user ID
// The user ID is extracted from the request context (set by auth middleware)
func RateLimitByUser(requests int, window time.Duration) Middleware {
	store := getUserRateLimitStore(requests, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := GetUser(r)
			if user == nil {
				// No user, allow request (or use IP-based limiting)
				next.ServeHTTP(w, r)
				return
			}

			// Get user ID (assuming user has an ID field)
			// This is a simplified implementation - adjust based on your user type
			userID := getUserID(user)
			if userID == "" {
				next.ServeHTTP(w, r)
				return
			}

			limiter := store.getLimiter(userID)

			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getUserID extracts user ID from user object
// This is a simplified implementation - adjust based on your user type
func getUserID(user interface{}) string {
	// Try common methods
	if u, ok := user.(interface{ GetID() string }); ok {
		return u.GetID()
	}
	if u, ok := user.(interface{ ID() string }); ok {
		return u.ID()
	}
	if u, ok := user.(interface{ GetID() int }); ok {
		return fmt.Sprintf("%d", u.GetID())
	}
	// Fallback: use string representation
	return fmt.Sprintf("%v", user)
}

// RateLimitGeneral creates a general rate limiting middleware
// It uses IP-based limiting by default, but can be configured
// Note: This is a proper implementation. The RateLimit in middleware.go is a stub.
func RateLimitGeneral(requests int, window time.Duration) Middleware {
	return RateLimitByIP(requests, window)
}
