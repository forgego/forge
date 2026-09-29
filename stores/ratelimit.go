package stores

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

// hitStatement counts one hit on a fixed-window bucket in one atomic
// statement: it starts a new window when there is no row or the row's
// window is over, and adds one to the current window otherwise. Concurrent
// callers on any instance serialize on the row, so no hit is lost.
const hitStatement = `INSERT INTO forge_rate_limits (bucket, hits, reset_at) VALUES ($1, 1, $2)
ON CONFLICT (bucket) DO UPDATE SET
	hits = CASE WHEN forge_rate_limits.reset_at <= $3 THEN 1 ELSE forge_rate_limits.hits + 1 END,
	reset_at = CASE WHEN forge_rate_limits.reset_at <= $3 THEN $2 ELSE forge_rate_limits.reset_at END
RETURNING hits, reset_at`

// hit records one hit on bucket and returns the hits in the current window
// and when that window ends.
func (s *Stores) hit(ctx context.Context, bucket string, window time.Duration) (int64, time.Time, error) {
	now := s.Now()
	var hits, resetAt int64
	err := s.db.QueryRowContext(ctx, hitStatement, bucket, unixMillis(now.Add(window)), unixMillis(now)).Scan(&hits, &resetAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	s.cleanupExpired(ctx, cleanupRateLimits)
	return hits, time.UnixMilli(resetAt), nil
}

// current returns the hits in bucket's current window, and when it ends.
// A missing or finished window has no hits.
func (s *Stores) current(ctx context.Context, bucket string) (int64, time.Time, error) {
	var hits, resetAt int64
	err := s.db.QueryRowContext(ctx,
		"SELECT hits, reset_at FROM forge_rate_limits WHERE bucket = $1 AND reset_at > $2",
		bucket, unixMillis(s.Now()),
	).Scan(&hits, &resetAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, time.Time{}, nil
	}
	if err != nil {
		return 0, time.Time{}, err
	}
	return hits, time.UnixMilli(resetAt), nil
}

// RateLimiter is a fixed-window counter shared through the database: at
// most limit calls per key are allowed in each window. It satisfies
// throttling.Store.
type RateLimiter struct {
	stores *Stores
	name   string
	limit  int
	window time.Duration
}

// RateLimiter returns a limiter allowing limit calls per key per window.
// name namespaces its keys: limiters with the same name, limit and window
// on any instance share their counts.
func (s *Stores) RateLimiter(name string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{stores: s, name: name, limit: limit, window: window}
}

// Allow implements throttling.Store. If the database cannot be reached the
// call is allowed and the error is logged, so a database outage does not
// turn into refused requests.
func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	allowed, retryAfter, err := l.AllowContext(context.Background(), key)
	if err != nil {
		log.Printf("forge/stores: rate limiter %s: %v", l.name, err)
		return true, 0
	}
	return allowed, retryAfter
}

// AllowContext counts one call for key and reports whether it is within
// the limit and, if not, how long until the window ends.
func (l *RateLimiter) AllowContext(ctx context.Context, key string) (bool, time.Duration, error) {
	if l.limit <= 0 {
		return false, l.window, nil
	}
	hits, resetAt, err := l.stores.hit(ctx, "rate:"+l.name+":"+key, l.window)
	if err != nil {
		return false, 0, err
	}
	if hits <= int64(l.limit) {
		return true, 0, nil
	}
	retryAfter := resetAt.Sub(l.stores.Now())
	if retryAfter < 0 {
		retryAfter = 0
	}
	return false, retryAfter, nil
}

// LoginAttempts counts failed logins per key in the database and locks a
// key out once it has max failures in one window. The admin uses it for its
// login lockout when server.stores is database.
type LoginAttempts struct {
	stores *Stores
	max    int
	window time.Duration
}

// LoginAttempts returns a lockout that blocks a key after max failures
// within window, until that window ends.
func (s *Stores) LoginAttempts(max int, window time.Duration) *LoginAttempts {
	return &LoginAttempts{stores: s, max: max, window: window}
}

func loginBucket(key string) string { return "login:" + key }

// Blocked reports whether key is locked out and for how long.
func (a *LoginAttempts) Blocked(ctx context.Context, key string) (bool, time.Duration, error) {
	hits, resetAt, err := a.stores.current(ctx, loginBucket(key))
	if err != nil {
		return false, 0, err
	}
	if hits < int64(a.max) {
		return false, 0, nil
	}
	return true, resetAt.Sub(a.stores.Now()), nil
}

// Failed records a failed login for key.
func (a *LoginAttempts) Failed(ctx context.Context, key string) error {
	_, _, err := a.stores.hit(ctx, loginBucket(key), a.window)
	return err
}

// Succeeded clears key's failures.
func (a *LoginAttempts) Succeeded(ctx context.Context, key string) error {
	_, err := a.stores.db.ExecContext(ctx, "DELETE FROM forge_rate_limits WHERE bucket = $1", loginBucket(key))
	return err
}
