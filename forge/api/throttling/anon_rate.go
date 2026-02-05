package throttling

import (
	"net/http"
	"time"
)

// AnonRateThrottle throttles anonymous (unauthenticated) requests
type AnonRateThrottle struct {
	// Rate is the rate limit (e.g., "100/hour")
	Rate string
	// Cache is the cache backend for storing throttle data
	Cache CacheBackend
	// Scope is the scope identifier for this throttle
	Scope string
}

// NewAnonRateThrottle creates a new anonymous rate throttle
func NewAnonRateThrottle(rate string, cache CacheBackend) *AnonRateThrottle {
	return &AnonRateThrottle{
		Rate:  rate,
		Cache: cache,
		Scope: "anon",
	}
}

// AllowRequest checks if the anonymous request should be allowed
func (t *AnonRateThrottle) AllowRequest(r *http.Request, view interface{}) (bool, time.Duration, error) {
	// Only throttle if user is not authenticated
	// In a real implementation, check if user is authenticated
	// For now, always apply to anonymous requests

	scope := t.GetScope(r, view)
	key := "throttle_anon_" + scope

	return t.checkRate(key, t.Rate)
}

// GetScope returns the scope identifier
func (t *AnonRateThrottle) GetScope(r *http.Request, view interface{}) string {
	// Use IP address as scope for anonymous users
	return getClientIP(r)
}

// checkRate checks if the rate limit is exceeded
func (t *AnonRateThrottle) checkRate(key, rate string) (bool, time.Duration, error) {
	if t.Cache == nil {
		// No cache, allow all requests
		return true, 0, nil
	}

	// Parse rate (e.g., "100/hour")
	limit, duration, err := parseRate(rate)
	if err != nil {
		return true, 0, err
	}

	// Get current count
	count, err := t.Cache.GetInt(key)
	if err != nil {
		count = 0
	}

	// Check if limit exceeded
	if count >= limit {
		// Get TTL to calculate wait duration
		ttl := t.Cache.GetTTL(key)
		return false, ttl, nil
	}

	// Increment count
	newCount := count + 1
	if err := t.Cache.Set(key, newCount, duration); err != nil {
		return true, 0, err
	}

	return true, 0, nil
}

// getClientIP gets the client IP address from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	// Check X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}
