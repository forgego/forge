// Package storestest opens databases with the framework store tables for
// tests.
//
// PostgreSQL tests use the database named by FORGE_TEST_DATABASE_URL (or
// DATABASE_URL). Each call to PostgresDSN creates a fresh schema in that
// database and drops it when the test ends, so tests never need to create
// databases. Without a URL the test is skipped, or fails when
// FORGE_REQUIRE_DB=1.
package storestest

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/stores"
	"github.com/lib/pq"
)

var schemaCounter atomic.Uint64

// PostgresURL returns FORGE_TEST_DATABASE_URL (or DATABASE_URL). Without
// one it skips t, or fails it when FORGE_REQUIRE_DB=1.
func PostgresURL(t testing.TB) string {
	t.Helper()
	for _, name := range []string{"FORGE_TEST_DATABASE_URL", "DATABASE_URL"} {
		if u := os.Getenv(name); u != "" {
			return u
		}
	}
	if os.Getenv("FORGE_REQUIRE_DB") == "1" {
		t.Fatal("PostgreSQL not configured: set FORGE_TEST_DATABASE_URL (FORGE_REQUIRE_DB=1 is set)")
	}
	t.Skip("PostgreSQL not configured: set FORGE_TEST_DATABASE_URL (set FORGE_REQUIRE_DB=1 to turn into failure)")
	return ""
}

// PostgresDSN creates an empty schema in the test database and returns a
// connection string whose search_path is that schema. The schema is dropped
// when t ends. Every connection opened with the DSN, like two application
// instances, sees the same tables.
func PostgresDSN(t testing.TB) string {
	t.Helper()
	base := PostgresURL(t)
	admin, err := sql.Open("postgres", base)
	if err != nil {
		t.Fatalf("opening PostgreSQL: %v", err)
	}
	defer admin.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := admin.PingContext(ctx); err != nil {
		if os.Getenv("FORGE_REQUIRE_DB") == "1" {
			t.Fatalf("PostgreSQL not reachable: %v (FORGE_REQUIRE_DB=1 is set)", err)
		}
		t.Skipf("PostgreSQL not reachable: %v (set FORGE_REQUIRE_DB=1 to turn into failure)", err)
	}

	schema := fmt.Sprintf("forge_stores_test_%d_%d_%d", os.Getpid(), time.Now().UnixNano(), schemaCounter.Add(1))
	if _, err := admin.ExecContext(ctx, "CREATE SCHEMA "+pq.QuoteIdentifier(schema)); err != nil {
		t.Fatalf("creating schema: %v", err)
	}
	t.Cleanup(func() {
		cleanup, err := sql.Open("postgres", base)
		if err != nil {
			return
		}
		defer cleanup.Close()
		_, _ = cleanup.Exec("DROP SCHEMA IF EXISTS " + pq.QuoteIdentifier(schema) + " CASCADE")
	})
	return withSearchPath(base, schema)
}

func withSearchPath(dsn, schema string) string {
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		if u, err := url.Parse(dsn); err == nil {
			q := u.Query()
			q.Set("search_path", schema)
			u.RawQuery = q.Encode()
			return u.String()
		}
	}
	return dsn + " search_path=" + schema
}

// Open connects to dsn with driver and closes the connection when t ends.
func Open(t testing.TB, driver, dsn string) *db.DB {
	t.Helper()
	database, err := db.NewDBWithDriver(driver, dsn)
	if err != nil {
		t.Fatalf("connecting to %s: %v", driver, err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

// SQLitePath returns a new SQLite database file in a temporary directory.
func SQLitePath(t testing.TB) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "stores.sqlite")
}

// Migrate applies the framework store migrations to database.
func Migrate(t testing.TB, database *db.DB) {
	t.Helper()
	if err := stores.Migrate(context.Background(), database); err != nil {
		t.Fatalf("migrating framework store tables: %v", err)
	}
}

// Backend is a database to run store tests against.
type Backend struct {
	Name   string
	Driver string
	// NewDSN returns the connection string of a new, empty database.
	// Opening it twice gives two connections to the same database, as two
	// application instances would have.
	NewDSN func(t testing.TB) string
}

// Backends returns a SQLite and a PostgreSQL backend. PostgreSQL skips (or
// fails with FORGE_REQUIRE_DB=1) without a test database.
func Backends() []Backend {
	return []Backend{
		{Name: "sqlite", Driver: "sqlite3", NewDSN: SQLitePath},
		{Name: "postgres", Driver: "postgres", NewDSN: PostgresDSN},
	}
}

// OpenMigrated opens a new database of backend with the framework store
// tables applied, and returns it with its DSN.
func (b Backend) OpenMigrated(t testing.TB) (*db.DB, string) {
	t.Helper()
	dsn := b.NewDSN(t)
	database := Open(t, b.Driver, dsn)
	Migrate(t, database)
	return database, dsn
}
