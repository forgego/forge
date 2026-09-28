package execute

import (
	"context"

	migrate "github.com/golang-migrate/migrate/v4"
)

// MigrationRunner runs migrations against a database
type MigrationRunner interface {
	// Run applies a migration
	Run(ctx context.Context, migration *migrate.Migration) error

	// Status returns the current migration status
	Status(ctx context.Context) (*MigrationStatus, error)

	// Rollback rolls back a migration
	Rollback(ctx context.Context, version string) error
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	Version    uint
	Dirty      bool
	Applied    []string
	Pending    []string
	OutOfOrder []string
	Current    string
	Next       string
}
