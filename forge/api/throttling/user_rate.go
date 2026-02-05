package throttling

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/forgego/forge/api/authentication"
)

// UserRateThrottle throttles authenticated user requests
type UserRateThrottle struct {
	// Rate is the rate limit (e.g., "1000/day")
	Rate string
	// Cache is the cache backend for storing throttle data
	Cache CacheBackend
	// Scope is the scope identifier for this throttle
	Scope string
}

// NewUserRateThrottle creates a new user rate throttle
func NewUserRateThrottle(rate string, cache CacheBackend) *UserRateThrottle {
	return &UserRateThrottle{
		Rate:  rate,
		Cache: cache,
		Scope: "user",
	}
}

// AllowRequest checks if the authenticated request should be allowed
func (t *UserRateThrottle) AllowRequest(r *http.Request, view interface{}) (bool, time.Duration, error) {
	// Get authenticated user
	_, ok := authentication.GetUserFromRequest(r)
	if !ok {
		// Not authenticated, don't throttle (let AnonRateThrottle handle it)
		return true, 0, nil
	}

	scope := t.GetScope(r, view)
	key := "throttle_user_" + scope

	return t.checkRate(key, t.Rate)
}

// GetScope returns the scope identifier (user ID)
func (t *UserRateThrottle) GetScope(r *http.Request, view interface{}) string {
	user, ok := authentication.GetUserFromRequest(r)
	if !ok {
		return ""
	}

	// Try to get user ID
	if id := getUserID(user); id != nil {
		return formatID(id)
	}

	// User exists but no ID found - use empty string as scope
	return ""
}

// checkRate checks if the rate limit is exceeded
func (t *UserRateThrottle) checkRate(key, rate string) (bool, time.Duration, error) {
	if t.Cache == nil {
		return true, 0, nil
	}

	limit, duration, err := parseRate(rate)
	if err != nil {
		return true, 0, err
	}

	count, err := t.Cache.GetInt(key)
	if err != nil {
		count = 0
	}

	if count >= limit {
		ttl := t.Cache.GetTTL(key)
		return false, ttl, nil
	}

	newCount := count + 1
	if err := t.Cache.Set(key, newCount, duration); err != nil {
		return true, 0, err
	}

	return true, 0, nil
}

// getUserID gets the user ID from a user object
func getUserID(user interface{}) interface{} {
	// Use reflection to get ID
	v := reflect.ValueOf(user)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Try GetID method
	if method := v.MethodByName("GetID"); method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			return results[0].Interface()
		}
	}

	// Try ID field
	if field := v.FieldByName("ID"); field.IsValid() {
		return field.Interface()
	}

	return nil
}

// formatID formats an ID as a string
func formatID(id interface{}) string {
	if id == nil {
		return ""
	}
	// Convert to string
	return fmt.Sprintf("%v", id)
}
