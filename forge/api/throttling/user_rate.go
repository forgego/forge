package throttling

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/forgego/forge/api/authentication"
	internalratelimit "github.com/forgego/forge/internal/ratelimit"
)

// UserRateThrottle throttles authenticated user requests.
type UserRateThrottle struct {
	Rate     string
	Scope    string
	store    Store
	parseErr error
}

// NewUserRateThrottle creates a new user rate throttle.
func NewUserRateThrottle(rate string) *UserRateThrottle {
	limit, window, err := parseRate(rate)
	var store Store
	if err == nil {
		store = internalratelimit.NewFixedWindowCounter(limit, window)
	}
	return &UserRateThrottle{
		Rate:     rate,
		Scope:    "user",
		store:    store,
		parseErr: err,
	}
}

// NewUserRateThrottleWithStore creates a new user rate throttle with a custom store.
func NewUserRateThrottleWithStore(rate string, store Store) *UserRateThrottle {
	throttle := NewUserRateThrottle(rate)
	throttle.store = store
	return throttle
}

// WithStore sets the rate limit store.
func (t *UserRateThrottle) WithStore(store Store) *UserRateThrottle {
	t.store = store
	return t
}

// AllowRequest checks whether the authenticated request should be allowed.
func (t *UserRateThrottle) AllowRequest(r *http.Request, view interface{}) (bool, time.Duration, error) {
	if t.parseErr != nil {
		return true, 0, t.parseErr
	}
	if t.store == nil {
		return true, 0, nil
	}
	key := "throttle_user_" + t.GetScope(r, view)
	allowed, retryAfter := t.store.Allow(key)
	return allowed, retryAfter, nil
}

// GetScope returns the authenticated user's ID.
func (t *UserRateThrottle) GetScope(r *http.Request, view interface{}) string {
	user, ok := authentication.GetUserFromRequest(r)
	if !ok {
		return getClientIP(r)
	}
	return formatID(getUserID(user))
}

func getUserID(user interface{}) interface{} {
	value := reflect.ValueOf(user)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if method := value.MethodByName("GetID"); method.IsValid() {
		results := method.Call(nil)
		if len(results) > 0 {
			return results[0].Interface()
		}
	}
	if field := value.FieldByName("ID"); field.IsValid() {
		return field.Interface()
	}
	return nil
}

func formatID(id interface{}) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("%v", id)
}
