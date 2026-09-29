package throttling

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/forgego/forge/api/authentication"
)

// UserRateThrottle throttles authenticated user requests.
type UserRateThrottle struct {
	Rate     string
	Scope    string
	store    Store // set by WithStore; nil uses the default factory
	defaults *defaultStore
	parseErr error
}

// NewUserRateThrottle creates a new user rate throttle.
func NewUserRateThrottle(rate string) *UserRateThrottle {
	limit, window, err := parseRate(rate)
	throttle := &UserRateThrottle{
		Rate:     rate,
		Scope:    "user",
		parseErr: err,
	}
	if err == nil {
		throttle.defaults = &defaultStore{name: "user/" + rate, limit: limit, window: window}
	}
	return throttle
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
	store := t.store
	if store == nil && t.defaults != nil {
		store = t.defaults.get()
	}
	if store == nil {
		return true, 0, nil
	}
	key := "throttle_user_" + t.GetScope(r, view)
	allowed, retryAfter := allow(store, r, key)
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
