// Package docker exposes the shared database test helpers for integration tests.
package docker

import (
	"context"
	"database/sql"
	"time"

	"github.com/forgego/forge/tests/testhelpers"
)

// PostgresOpts uses the shared URL precedence and DSN construction.
type PostgresOpts = testhelpers.PostgresOpts

func DefaultPostgresOpts() PostgresOpts { return testhelpers.DefaultPostgresOpts() }
func DefaultPostgresOptsWithTest(testName string) PostgresOpts {
	return testhelpers.DefaultPostgresOptsWithTest(testName)
}
func GetDockerEndpoint() string { return testhelpers.GetDockerEndpoint() }
func StartPostgresContainer(ctx context.Context, opts PostgresOpts) (*sql.DB, string, func() error, error) {
	return testhelpers.StartPostgresContainer(ctx, opts)
}
func StartSQLiteMemory(dsn string) (*sql.DB, error) { return testhelpers.StartSQLiteMemory(dsn) }
func WaitForDBReady(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	return testhelpers.WaitForDBReady(ctx, db, timeout)
}
