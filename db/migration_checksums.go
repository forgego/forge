package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/forgego/forge/db/migrate/checksum"
	"github.com/golang-migrate/migrate/v4"
)

// migrationChecksumLockKey identifies Forge's migration-and-checksum critical section.
// It serializes Forge MigrationRunner operations only; external golang-migrate tools
// are not covered. It is distinct from golang-migrate's per-step advisory lock.
const migrationChecksumLockKey int64 = 0x466f7267654d6967

const postgresMigrationChecksumsDDL = `CREATE TABLE IF NOT EXISTS forge_migration_checksums (version BIGINT PRIMARY KEY, up_sha256 CHAR(64) NOT NULL, down_sha256 CHAR(64), applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`
const sqliteMigrationChecksumsDDL = `CREATE TABLE IF NOT EXISTS forge_migration_checksums (version INTEGER PRIMARY KEY, up_sha256 TEXT NOT NULL, down_sha256 TEXT, applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`

type checksumExecutor interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (mr *MigrationRunner) withMigrationChecksums(ctx context.Context, run func(checksumExecutor) error) (err error) {
	var executor checksumExecutor = mr.db
	ddl := sqliteMigrationChecksumsDDL
	if mr.db.Driver == "postgres" || mr.db.Driver == "postgresql" {
		if mr.db.Stats().MaxOpenConnections == 1 {
			return fmt.Errorf("PostgreSQL migrations need at least 2 connections")
		}
		ddl = postgresMigrationChecksumsDDL
		conn, connErr := mr.db.DB.Conn(ctx)
		if connErr != nil {
			return connErr
		}
		defer conn.Close()
		executor = conn
		if _, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationChecksumLockKey); err != nil {
			return err
		}
		defer func() {
			unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, unlockErr := conn.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", migrationChecksumLockKey)
			if unlockErr != nil {
				// Discard the session so a failed unlock cannot leak a lock into the pool.
				_ = conn.Raw(func(any) error { return driver.ErrBadConn })
				err = errors.Join(err, fmt.Errorf("release migration checksum lock: %w", unlockErr))
			}
		}()
	}
	// SQLite serializes writers itself; no additional advisory lock is needed.
	if _, err := executor.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("create migration checksum table: %w", err)
	}
	if err := mr.reconcileMigrationChecksums(ctx, executor); err != nil {
		return err
	}
	return run(executor)
}

type migrationFilePair struct{ up, down string }
type migrationFileIndex map[uint]migrationFilePair

func indexMigrationFiles(path string) (migrationFileIndex, error) {
	files, err := checksum.IndexMigrationFiles(path)
	if err != nil {
		return nil, err
	}
	index := make(migrationFileIndex, len(files))
	for version, pair := range files {
		index[version] = migrationFilePair{up: pair.Up, down: pair.Down}
	}
	return index, nil
}

// MigrationFileChecksums computes hashes for one migration version.
func MigrationFileChecksums(path string, version uint) (string, *string, error) {
	return checksum.MigrationFileChecksums(path, version)
}

func migrationFileChecksums(path string, version uint) (string, *string, error) {
	return MigrationFileChecksums(path, version)
}

func (index migrationFileIndex) checksums(version uint) (string, *string, error) {
	pair := index[version]
	return (checksum.FileIndex{version: {Up: pair.up, Down: pair.down}}).Checksums(version)
}

func (mr *MigrationRunner) applyChecksumSteps(ctx context.Context, executor checksumExecutor, target *uint) error {
	index, err := indexMigrationFiles(mr.migrationsPath)
	if err != nil {
		return err
	}
	if target != nil {
		if _, _, err := index.checksums(*target); err != nil {
			return err
		}
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		version, _, err := mr.migrate.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return err
		}
		if target != nil && err == nil && version == *target {
			return nil
		}
		err = mr.migrate.Steps(1)
		var short migrate.ErrShortLimit
		// Steps(1) returns bare os.ErrNotExist when the source has no next version.
		// File read failures are wrapped and must still be reported.
		if err == os.ErrNotExist || errors.Is(err, migrate.ErrNoChange) || errors.As(err, &short) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}
		version, _, err = mr.migrate.Version()
		if err != nil {
			return err
		}
		up, down, err := index.checksums(version)
		if err != nil {
			return err
		}
		_, err = executor.ExecContext(ctx, `INSERT INTO forge_migration_checksums (version, up_sha256, down_sha256) VALUES ($1, $2, $3) ON CONFLICT (version) DO UPDATE SET up_sha256 = excluded.up_sha256, down_sha256 = excluded.down_sha256, applied_at = CURRENT_TIMESTAMP`, version, up, down)
		if err != nil {
			return fmt.Errorf("record migration version %d: %w", version, err)
		}
	}
}

func (mr *MigrationRunner) rollbackChecksumStep(ctx context.Context, executor checksumExecutor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stepErr := mr.migrate.Steps(-1)
	reconcileErr := mr.reconcileMigrationChecksums(ctx, executor)
	return errors.Join(stepErr, reconcileErr)
}

func (mr *MigrationRunner) reconcileMigrationChecksums(ctx context.Context, executor checksumExecutor) error {
	version, dirty, err := mr.migrate.Version()
	if err == migrate.ErrNilVersion {
		_, err = executor.ExecContext(ctx, "DELETE FROM forge_migration_checksums")
		return err
	}
	if err != nil {
		return err
	}
	if dirty {
		return nil
	}
	_, err = executor.ExecContext(ctx, "DELETE FROM forge_migration_checksums WHERE version > $1", version)
	return err
}
