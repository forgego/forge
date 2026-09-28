package exceptions

import "time"

// Throttled represents a rate limit exceeded error
type Throttled struct {
	*APIException
	RetryAfter time.Duration
}

// NewThrottled creates a new throttled exception
func NewThrottled(message string, retryAfter time.Duration) *Throttled {
	if message == "" {
		message = "Request was throttled"
	}
	return &Throttled{
		APIException: NewAPIException(
			429,
			"throttled",
			message,
			nil,
		),
		RetryAfter: retryAfter,
	}
}
