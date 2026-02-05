package throttling

import (
	"net/http"
	"time"
)

// Throttle is the interface for throttle classes
type Throttle interface {
	// AllowRequest checks if the request should be allowed
	// Returns (allowed, waitDuration, error)
	// If allowed is false, waitDuration indicates how long to wait before retrying
	AllowRequest(r *http.Request, view interface{}) (bool, time.Duration, error)

	// GetScope returns a unique identifier for this throttle scope
	// Used to identify different rate limit buckets
	GetScope(r *http.Request, view interface{}) string
}

// CheckThrottles checks a list of throttles
// Returns error if any throttle denies the request
func CheckThrottles(r *http.Request, view interface{}, throttles []Throttle) error {
	for _, throttle := range throttles {
		allowed, waitDuration, err := throttle.AllowRequest(r, view)
		if err != nil {
			return err
		}
		if !allowed {
			return NewThrottledError(waitDuration)
		}
	}
	return nil
}

// ThrottledError represents a throttled request
type ThrottledError struct {
	WaitDuration time.Duration
}

// NewThrottledError creates a new throttled error
func NewThrottledError(waitDuration time.Duration) *ThrottledError {
	return &ThrottledError{WaitDuration: waitDuration}
}

// Error implements the error interface
func (e *ThrottledError) Error() string {
	return "Request was throttled"
}
