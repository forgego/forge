package throttling

import "time"

// Store represents a backend store for rate limiting.
// Implementations must be safe for concurrent use by multiple goroutines.
type Store interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}
