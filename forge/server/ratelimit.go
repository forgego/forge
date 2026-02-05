package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is the interface for rate limiters
type RateLimiter interface {
	Allow(ctx context.Context) bool
	Wait(ctx context.Context) error
}

// rateLimitStore stores rate limiters for different keys
type rateLimitStore struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
	cleanup  *time.Ticker
}

// newRateLimitStore creates a new rate limit store
func newRateLimitStore(r rate.Limit, burst int) *rateLimitStore {
	store := &rateLimitStore{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
		cleanup:  time.NewTicker(5 * time.Minute),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// getLimiter gets or creates a limiter for a key
func (s *rateLimitStore) getLimiter(key string) *rate.Limiter {
	s.mu.RLock()
	limiter, exists := s.limiters[key]
	s.mu.RUnlock()

	if exists {
		return limiter
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := s.limiters[key]; exists {
		return limiter
	}

	limiter = rate.NewLimiter(s.rate, s.burst)
	s.limiters[key] = limiter
	return limiter
}

// cleanupExpired periodically cleans up old limiters
func (s *rateLimitStore) cleanupExpired() {
	for range s.cleanup.C {
		s.mu.Lock()
		// Simple cleanup: remove limiters that haven't been used recently
		// In production, you might want a more sophisticated approach
		if len(s.limiters) > 1000 {
			// Clear half of the limiters (simple strategy)
			// In production, use LRU or time-based expiration
			newLimiters := make(map[string]*rate.Limiter, len(s.limiters)/2)
			count := 0
			for k, v := range s.limiters {
				if count < len(s.limiters)/2 {
					newLimiters[k] = v
					count++
				}
			}
			s.limiters = newLimiters
		}
		s.mu.Unlock()
	}
}

// stop stops the cleanup ticker
func (s *rateLimitStore) stop() {
	s.cleanup.Stop()
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
	return string(rune(requests)) + "-" + window.String()
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (original client)
		ips := splitIPs(xff)
		if len(ips) > 0 {
			return ips[0]
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if host, _, err := splitHostPort(ip); err == nil {
		return host
	}

	return ip
}

// splitIPs splits comma-separated IPs
func splitIPs(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// splitHostPort splits host:port (simplified)
func splitHostPort(hostport string) (host, port string, err error) {
	// Simple implementation - in production use net.SplitHostPort
	idx := strings.LastIndex(hostport, ":")
	if idx < 0 {
		return hostport, "", nil
	}
	return hostport[:idx], hostport[idx+1:], nil
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
