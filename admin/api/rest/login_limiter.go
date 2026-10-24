package rest

import (
	"context"
	"time"
)

// LoginAttemptStore counts failed admin logins per key (the client IP and
// the username) and locks a key out after too many. stores.LoginAttempts
// keeps them in the database so the lockout holds across restarts and
// instances.
type LoginAttemptStore interface {
	// Blocked reports whether key is locked out and for how long.
	Blocked(ctx context.Context, key string) (bool, time.Duration, error)
	// Failed records a failed login for key.
	Failed(ctx context.Context, key string) error
	// Succeeded clears the failures of key.
	Succeeded(ctx context.Context, key string) error
}
