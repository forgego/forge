package stores_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// forEachBackend runs fn against SQLite and PostgreSQL.
func forEachBackend(t *testing.T, fn func(t *testing.T, backend storestest.Backend)) {
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) { fn(t, backend) })
	}
}

// twoInstances opens two connections to one migrated database, as two
// application instances would, and returns a Stores for each sharing clock.
func twoInstances(t *testing.T, backend storestest.Backend, clock func() time.Time) (*stores.Stores, *stores.Stores) {
	t.Helper()
	first, dsn := backend.OpenMigrated(t)
	second := storestest.Open(t, backend.Driver, dsn)
	a, err := stores.New(first)
	require.NoError(t, err)
	b, err := stores.New(second)
	require.NoError(t, err)
	if clock != nil {
		a.SetClock(clock)
		b.SetClock(clock)
	}
	return a, b
}

func TestParseKind(t *testing.T) {
	for value, want := range map[string]string{"": "memory", "memory": "memory", " Database ": "database"} {
		got, err := stores.ParseKind(value)
		require.NoError(t, err)
		assert.Equal(t, want, got, value)
	}
	_, err := stores.ParseKind("redis")
	assert.ErrorContains(t, err, `invalid server.stores "redis"`)
}

func TestMigrate_CreatesTablesOnceAndKeepsTheConnectionOpen(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		ctx := context.Background()
		database := storestest.Open(t, backend.Driver, backend.NewDSN(t))

		err := stores.CheckMigrated(ctx, database)
		require.ErrorIs(t, err, stores.ErrNotMigrated)

		require.NoError(t, stores.Migrate(ctx, database))
		require.NoError(t, stores.Migrate(ctx, database), "a second run applies nothing")
		require.NoError(t, stores.CheckMigrated(ctx, database))
		require.NoError(t, database.PingContext(ctx), "migrating must not close the application's database")

		version, dirty, err := stores.Version(ctx, database)
		require.NoError(t, err)
		latest, err := stores.LatestVersion(database.Driver)
		require.NoError(t, err)
		assert.Equal(t, latest, version)
		assert.False(t, dirty)

		for _, table := range []string{"forge_sessions", "forge_rate_limits", "forge_admin_tokens", "forge_admin_saved_views", "forge_admin_log"} {
			var count int
			require.NoError(t, database.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count), table)
		}
		// The application's own migration table is untouched.
		var appTables int
		query := "SELECT count(*) FROM information_schema.tables WHERE table_name = 'schema_migrations' AND table_schema = current_schema()"
		if backend.Driver == "sqlite3" {
			query = "SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'"
		}
		require.NoError(t, database.QueryRowContext(ctx, query).Scan(&appTables))
		assert.Zero(t, appTables)
	})
}

func TestSessionStore_SharedBetweenInstances(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		ctx := context.Background()
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		clock := func() time.Time { return now }
		a, b := twoInstances(t, backend, clock)

		require.NoError(t, a.Sessions().CommitCtx(ctx, "tok", []byte("payload"), now.Add(time.Hour)))
		data, found, err := b.Sessions().FindCtx(ctx, "tok")
		require.NoError(t, err)
		require.True(t, found, "a session committed on A must be found on B")
		assert.Equal(t, []byte("payload"), data)

		require.NoError(t, b.Sessions().Commit("tok", []byte("updated"), now.Add(2*time.Hour)))
		data, found, err = a.Sessions().Find("tok")
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, []byte("updated"), data)

		now = now.Add(3 * time.Hour)
		_, found, err = a.Sessions().Find("tok")
		require.NoError(t, err)
		assert.False(t, found, "an expired session is not found")

		require.NoError(t, a.Sessions().Commit("gone", []byte("x"), now.Add(time.Hour)))
		require.NoError(t, b.Sessions().Delete("gone"))
		_, found, err = a.Sessions().Find("gone")
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func TestRateLimiter_CountsAtomicallyAcrossInstances(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		a, b := twoInstances(t, backend, nil)
		const limit = 7
		limiters := []*stores.RateLimiter{
			a.RateLimiter("test", limit, time.Minute),
			b.RateLimiter("test", limit, time.Minute),
		}
		var allowed atomic.Int32
		var wg sync.WaitGroup
		for i := range 40 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, _, err := limiters[i%2].AllowContext(context.Background(), "client")
				assert.NoError(t, err)
				if ok {
					allowed.Add(1)
				}
			}()
		}
		wg.Wait()
		assert.Equal(t, int32(limit), allowed.Load(), "exactly limit calls pass, whichever instance serves them")

		other, _ := limiters[0].Allow("other-client")
		assert.True(t, other, "keys are counted separately")
		otherName, _ := a.RateLimiter("other-name", limit, time.Minute).Allow("client")
		assert.True(t, otherName, "limiter names are counted separately")
	})
}

func TestRateLimiter_WindowResetAndRetryAfter(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		a, b := twoInstances(t, backend, func() time.Time { return now })
		la := a.RateLimiter("w", 2, time.Minute)
		lb := b.RateLimiter("w", 2, time.Minute)

		ok, _ := la.Allow("k")
		require.True(t, ok)
		ok, _ = lb.Allow("k")
		require.True(t, ok)
		now = now.Add(20 * time.Second)
		ok, retry := la.Allow("k")
		assert.False(t, ok)
		assert.Equal(t, 40*time.Second, retry)

		now = now.Add(40 * time.Second)
		ok, retry = lb.Allow("k")
		assert.True(t, ok, "a new window starts when the old one ends")
		assert.Zero(t, retry)

		ok, retry = a.RateLimiter("zero", 0, time.Minute).Allow("k")
		assert.False(t, ok)
		assert.Equal(t, time.Minute, retry)
	})
}

func TestLoginAttempts_LockoutSharedBetweenInstances(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		ctx := context.Background()
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		a, b := twoInstances(t, backend, func() time.Time { return now })
		la := a.LoginAttempts(5, 15*time.Minute)
		lb := b.LoginAttempts(5, 15*time.Minute)

		for i := range 4 {
			attempts := la
			if i%2 == 1 {
				attempts = lb
			}
			require.NoError(t, attempts.Failed(ctx, "ip:10.0.0.1"))
			blocked, _, err := lb.Blocked(ctx, "ip:10.0.0.1")
			require.NoError(t, err)
			assert.False(t, blocked, "failure %d does not lock out yet", i+1)
		}
		require.NoError(t, lb.Failed(ctx, "ip:10.0.0.1"))
		blocked, remaining, err := la.Blocked(ctx, "ip:10.0.0.1")
		require.NoError(t, err)
		assert.True(t, blocked, "five failures split across instances lock the key out")
		assert.Equal(t, 15*time.Minute, remaining)

		now = now.Add(10 * time.Minute)
		blocked, remaining, err = lb.Blocked(ctx, "ip:10.0.0.1")
		require.NoError(t, err)
		assert.True(t, blocked)
		assert.Equal(t, 5*time.Minute, remaining)

		now = now.Add(5*time.Minute + time.Second)
		blocked, _, err = la.Blocked(ctx, "ip:10.0.0.1")
		require.NoError(t, err)
		assert.False(t, blocked, "the lockout ends with its window")

		require.NoError(t, la.Failed(ctx, "user:alice"))
		require.NoError(t, lb.Succeeded(ctx, "user:alice"))
		for range 4 {
			require.NoError(t, la.Failed(ctx, "user:alice"))
		}
		blocked, _, err = la.Blocked(ctx, "user:alice")
		require.NoError(t, err)
		assert.False(t, blocked, "a successful login cleared the earlier failure")
	})
}

func TestAdminTokens_SharedHashedAndRevocable(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		ctx := context.Background()
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		a, b := twoInstances(t, backend, func() time.Time { return now })

		token, err := a.AdminTokens().IssueToken(ctx, "alice", time.Hour)
		require.NoError(t, err)
		require.Len(t, token, 64)

		username, ok, err := b.AdminTokens().ValidateToken(ctx, token)
		require.NoError(t, err)
		require.True(t, ok, "a token issued on A is valid on B")
		assert.Equal(t, "alice", username)

		var stored int
		require.NoError(t, a.DB().QueryRowContext(ctx, "SELECT count(*) FROM forge_admin_tokens WHERE token_hash = $1", token).Scan(&stored))
		assert.Zero(t, stored, "the raw token is never stored")

		_, ok, err = b.AdminTokens().ValidateToken(ctx, "not-a-token")
		require.NoError(t, err)
		assert.False(t, ok)

		require.NoError(t, b.AdminTokens().RevokeToken(ctx, token))
		_, ok, err = a.AdminTokens().ValidateToken(ctx, token)
		require.NoError(t, err)
		assert.False(t, ok, "a token revoked on B is rejected on A")

		expiring, err := a.AdminTokens().IssueToken(ctx, "bob", time.Minute)
		require.NoError(t, err)
		now = now.Add(2 * time.Minute)
		_, ok, err = b.AdminTokens().ValidateToken(ctx, expiring)
		require.NoError(t, err)
		assert.False(t, ok, "an expired token is rejected")

		_, err = a.AdminTokens().IssueToken(ctx, "carol", 0)
		assert.Error(t, err)
	})
}

func TestExpiredRowsAreDeletedWithoutACronJob(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		ctx := context.Background()
		now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		a, _ := twoInstances(t, backend, func() time.Time { return now })

		require.NoError(t, a.Sessions().Commit("old", []byte("x"), now.Add(time.Minute)))
		_, _ = a.RateLimiter("r", 1, time.Minute).Allow("old")
		_, err := a.AdminTokens().IssueToken(ctx, "old", time.Minute)
		require.NoError(t, err)

		now = now.Add(time.Hour)
		require.NoError(t, a.Sessions().Commit("new", []byte("x"), now.Add(time.Hour)))
		_, _ = a.RateLimiter("r", 1, time.Minute).Allow("new")
		_, err = a.AdminTokens().IssueToken(ctx, "new", time.Hour)
		require.NoError(t, err)

		for table, want := range map[string]int{"forge_sessions": 1, "forge_rate_limits": 1, "forge_admin_tokens": 1} {
			var count int
			require.NoError(t, a.DB().QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count))
			assert.Equal(t, want, count, "%s keeps only the unexpired row", table)
		}
	})
}

func TestNew_RejectsUnsupportedDatabases(t *testing.T) {
	_, err := stores.New(nil)
	assert.Error(t, err)
	err = stores.CheckMigrated(context.Background(), nil)
	assert.Error(t, err)
	assert.False(t, errors.Is(err, stores.ErrNotMigrated))
}

func TestRateLimiter_FailsOpenWhenTheDatabaseStalls(t *testing.T) {
	forEachBackend(t, func(t *testing.T, backend storestest.Backend) {
		database, _ := backend.OpenMigrated(t)
		s, err := stores.New(database)
		require.NoError(t, err)
		limiter := s.RateLimiter("stall", 1, time.Minute)

		// Exhaust the pool: the limiter's query cannot get a connection.
		database.SetMaxOpenConns(1)
		held, err := database.Conn(context.Background())
		require.NoError(t, err)
		t.Cleanup(func() { _ = held.Close() })

		start := time.Now()
		ok, retry := limiter.Allow("k")
		assert.True(t, ok, "an unreachable store allows the request")
		assert.Zero(t, retry)
		assert.Less(t, time.Since(start), 10*time.Second, "Allow must not wait for the database forever")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, err = limiter.AllowContext(ctx, "k")
		assert.ErrorIs(t, err, context.Canceled, "a canceled request stops the query")
	})
}
