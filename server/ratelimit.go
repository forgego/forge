package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	internalratelimit "github.com/forgego/forge/internal/ratelimit"
	"github.com/forgego/forge/netutil"
)

// RateLimiter is the interface for rate limiters.
type RateLimiter interface {
	Allow(ctx context.Context) bool
	Wait(ctx context.Context) error
}

// KeyedLimiter applies independent token buckets to arbitrary keys.
type KeyedLimiter = internalratelimit.KeyedLimiter

// NewKeyedLimiter creates a keyed limiter with the supplied rate.
func NewKeyedLimiter(requests int, window time.Duration) *KeyedLimiter {
	return internalratelimit.NewKeyedLimiter(requests, window)
}

func getClientIP(r *http.Request) string {
	return netutil.ClientIP(r, netutil.TrustedProxies())
}

// RateLimitByIP creates middleware that rate limits each client IP.
func RateLimitByIP(requests int, window time.Duration) Middleware {
	limiter := NewKeyedLimiter(requests, window)
	return limitRequests(limiter, func(r *http.Request) string { return getClientIP(r) })
}

// RateLimitByUser creates middleware that rate limits each authenticated user.
func RateLimitByUser(requests int, window time.Duration) Middleware {
	limiter := NewKeyedLimiter(requests, window)
	return limitRequests(limiter, userRateLimitKey)
}

func limitRequests(limiter *KeyedLimiter, keyFor func(*http.Request) string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFor(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			if allowed, _ := limiter.Reserve(key); !allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func userRateLimitKey(r *http.Request) string {
	user := GetUser(r)
	if user == nil {
		return ""
	}
	return getUserID(user)
}

func getUserID(user interface{}) string {
	if u, ok := user.(interface{ GetID() string }); ok {
		return u.GetID()
	}
	if u, ok := user.(interface{ ID() string }); ok {
		return u.ID()
	}
	if u, ok := user.(interface{ GetID() int }); ok {
		return fmt.Sprintf("%d", u.GetID())
	}
	return fmt.Sprintf("%v", user)
}

// RateLimitGeneral creates an IP-based rate-limiting middleware.
func RateLimitGeneral(requests int, window time.Duration) Middleware {
	return RateLimitByIP(requests, window)
}
