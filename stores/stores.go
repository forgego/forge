// Package stores keeps state that the server and the admin would otherwise
// hold in process memory in the application database, so that it survives a
// restart and is shared by every instance that uses the same database.
//
// It backs:
//
//   - scs sessions (the forge_session cookie): Stores.Sessions
//   - API throttling counters: Stores.RateLimiter
//   - admin bearer tokens: Stores.AdminTokens
//   - the admin login lockout: Stores.LoginAttempts
//   - admin saved views and change history: package stores/adminstore
//
// server.NewServer (with server.WithDatabase) and admin.Site.UseDatabaseStores
// wire them in when server.stores is database.
//
// The tables are framework-owned migrations embedded in this package. With
// server.stores set to database, `forge migrate up` applies them before the
// application's migrations and records them in forge_framework_migrations,
// separately from the application's schema_migrations. Call Migrate to apply
// them from Go instead. Nothing creates tables at runtime.
package stores

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/forgego/forge/db"
)

// Setting values for server.stores.
const (
	// KindMemory keeps every store in process memory (the default).
	KindMemory = "memory"
	// KindDatabase keeps the stores in the application database.
	KindDatabase = "database"
)

// ParseKind normalizes a server.stores value. An empty value is memory.
func ParseKind(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", KindMemory:
		return KindMemory, nil
	case KindDatabase:
		return KindDatabase, nil
	default:
		return "", fmt.Errorf("invalid server.stores %q: use %q or %q", value, KindDatabase, KindMemory)
	}
}

// cleanupInterval is how often one process deletes expired rows from a
// table. Cleanup runs inline on the write path, so no cron job is needed.
const cleanupInterval = time.Minute

// Stores is a handle on the framework store tables of one database. It is
// safe for concurrent use.
type Stores struct {
	db  *db.DB
	now func() time.Time

	mu          sync.Mutex
	lastCleanup map[string]time.Time
}

// New returns the stores kept in database, which must be PostgreSQL or
// SQLite. It does not touch the database; call CheckMigrated to verify the
// tables exist.
func New(database *db.DB) (*Stores, error) {
	if database == nil || database.DB == nil {
		return nil, errors.New("stores: database is nil")
	}
	if _, err := migrationsDir(database.Driver); err != nil {
		return nil, err
	}
	return &Stores{db: database, now: time.Now, lastCleanup: make(map[string]time.Time)}, nil
}

// DB returns the database the stores use.
func (s *Stores) DB() *db.DB { return s.db }

// Now returns the current time used for expiries.
func (s *Stores) Now() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// SetClock replaces the clock used for expiries. It is meant for tests.
func (s *Stores) SetClock(now func() time.Time) { s.now = now }

// unixMillis converts t to the Unix milliseconds stored in expiry columns.
func unixMillis(t time.Time) int64 { return t.UnixMilli() }

// Expired-row cleanup statements, one per table with an expiry column.
const (
	cleanupSessions   = "DELETE FROM forge_sessions WHERE expires_at <= $1"
	cleanupRateLimits = "DELETE FROM forge_rate_limits WHERE reset_at <= $1"
	cleanupTokens     = "DELETE FROM forge_admin_tokens WHERE expires_at <= $1"
)

// cleanupExpired runs one of the cleanup statements, at most once per
// minute per statement and process. Stores call it on their write paths,
// so expired rows go away without a cron job.
func (s *Stores) cleanupExpired(ctx context.Context, statement string) {
	now := s.Now()
	s.mu.Lock()
	last := s.lastCleanup[statement]
	if !last.IsZero() && now.Sub(last) < cleanupInterval && !now.Before(last) {
		s.mu.Unlock()
		return
	}
	s.lastCleanup[statement] = now
	s.mu.Unlock()

	if _, err := s.db.ExecContext(ctx, statement, unixMillis(now)); err != nil {
		log.Printf("forge/stores: deleting expired rows: %v", err)
	}
}
