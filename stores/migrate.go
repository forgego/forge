package stores

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/forgego/forge/db"
	"github.com/golang-migrate/migrate/v4"
	migratedb "github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// MigrationsTable records which framework store migrations are applied. It
// is separate from the application's schema_migrations table, so framework
// and application migration versions never collide.
const MigrationsTable = "forge_framework_migrations"

//go:embed migrations
var migrationFiles embed.FS

// ErrNotMigrated is returned by CheckMigrated when the store tables are
// missing or older than this version of Forge expects.
var ErrNotMigrated = errors.New("framework store tables are not migrated")

func migrationsDir(driver string) (string, error) {
	switch driver {
	case "postgres", "postgresql":
		return "migrations/postgres", nil
	case "sqlite", "sqlite3":
		return "migrations/sqlite3", nil
	default:
		return "", fmt.Errorf("stores: unsupported database driver %q", driver)
	}
}

// LatestVersion is the newest framework store migration version for driver.
func LatestVersion(driver string) (uint64, error) {
	dir, err := migrationsDir(driver)
	if err != nil {
		return 0, err
	}
	entries, err := fs.ReadDir(migrationFiles, dir)
	if err != nil {
		return 0, err
	}
	var latest uint64
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		prefix, _, _ := strings.Cut(name, "_")
		version, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("stores: bad migration file name %q", name)
		}
		latest = max(latest, version)
	}
	if latest == 0 {
		return 0, fmt.Errorf("stores: no migrations for %s", driver)
	}
	return latest, nil
}

// Migrate applies the pending framework store migrations to database. It
// is what `forge migrate up` runs when server.stores is database, and it is
// safe to run again: applied migrations are skipped.
func Migrate(ctx context.Context, database *db.DB) error {
	return withMigrator(ctx, database, func(m *migrate.Migrate) error {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("applying framework store migrations: %w", err)
		}
		return nil
	})
}

// Version reports the applied framework store migration version (0 when
// none is applied) and whether the last migration failed half-way.
func Version(ctx context.Context, database *db.DB) (version uint64, dirty bool, err error) {
	err = withMigrator(ctx, database, func(m *migrate.Migrate) error {
		v, d, verr := m.Version()
		if errors.Is(verr, migrate.ErrNilVersion) {
			return nil
		}
		version, dirty = uint64(v), d
		return verr
	})
	return version, dirty, err
}

// CheckMigrated returns an error wrapping ErrNotMigrated unless every
// framework store migration is applied to database.
func CheckMigrated(ctx context.Context, database *db.DB) error {
	if database == nil || database.DB == nil {
		return errors.New("stores: database is nil")
	}
	latest, err := LatestVersion(database.Driver)
	if err != nil {
		return err
	}
	var version uint64
	var dirty bool
	row := database.QueryRowContext(ctx, "SELECT version, dirty FROM forge_framework_migrations LIMIT 1")
	if err := row.Scan(&version, &dirty); err != nil {
		return fmt.Errorf("%w (run `forge migrate up` with server.stores: database): %v", ErrNotMigrated, err)
	}
	if dirty {
		return fmt.Errorf("%w: version %d is dirty", ErrNotMigrated, version)
	}
	if version < latest {
		return fmt.Errorf("%w: version %d, want %d (run `forge migrate up`)", ErrNotMigrated, version, latest)
	}
	return nil
}

// withMigrator runs fn with a golang-migrate instance over the embedded
// migrations. It never closes database: for PostgreSQL it works on one
// dedicated connection, which it closes itself.
func withMigrator(ctx context.Context, database *db.DB, fn func(*migrate.Migrate) error) error {
	if database == nil || database.DB == nil {
		return errors.New("stores: database is nil")
	}
	dir, err := migrationsDir(database.Driver)
	if err != nil {
		return err
	}
	source, err := iofs.New(migrationFiles, dir)
	if err != nil {
		return fmt.Errorf("stores: reading embedded migrations: %w", err)
	}
	defer source.Close()

	var driver migratedb.Driver
	var driverName string
	switch database.Driver {
	case "postgres", "postgresql":
		conn, err := database.DB.Conn(ctx)
		if err != nil {
			return fmt.Errorf("stores: opening a migration connection: %w", err)
		}
		defer conn.Close()
		driver, err = postgres.WithConnection(ctx, conn, &postgres.Config{MigrationsTable: MigrationsTable})
		if err != nil {
			return fmt.Errorf("stores: preparing migrations: %w", err)
		}
		driverName = "postgres"
	default:
		// sqlite3.WithInstance keeps a reference to database but only
		// closes it from Close, which is never called here.
		driver, err = sqlite3.WithInstance(database.DB, &sqlite3.Config{MigrationsTable: MigrationsTable})
		if err != nil {
			return fmt.Errorf("stores: preparing migrations: %w", err)
		}
		driverName = "sqlite3"
	}

	m, err := migrate.NewWithInstance("iofs", source, driverName, driver)
	if err != nil {
		return fmt.Errorf("stores: preparing migrations: %w", err)
	}
	return fn(m)
}
