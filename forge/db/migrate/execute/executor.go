package execute

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/forgego/forge/config"
	migrate "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Executor wraps golang-migrate for migration execution
type Executor struct {
	migrate        *migrate.Migrate
	migrationsPath string
}

// NewExecutor creates a new migration executor
func NewExecutor(db database.Driver, migrationsPath string) (*Executor, error) {
	// Always convert to absolute path to avoid issues with working directory changes
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for migrations: %w", err)
	}

	// Clean the path to normalize it
	absPath = filepath.Clean(absPath)

	// Verify the migrations directory exists
	if stat, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s", absPath)
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat migrations directory %s: %w", absPath, err)
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("migrations path is not a directory: %s", absPath)
	}

	// Check if directory has any migration files (golang-migrate requires at least one)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", absPath, err)
	}
	hasMigrations := false
	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if (len(name) > 7 && name[len(name)-7:] == ".up.sql") ||
				(len(name) > 9 && name[len(name)-9:] == ".down.sql") {
				hasMigrations = true
				break
			}
		}
	}
	if !hasMigrations {
		return nil, fmt.Errorf("migrations directory %s exists but contains no migration files", absPath)
	}

	// Convert to forward slashes for URL
	urlPath := filepath.ToSlash(absPath)

	// Build file:// URL for golang-migrate
	var migrationsURL string
	if len(urlPath) >= 2 && urlPath[1] == ':' {
		// Windows path with drive letter
		migrationsURL = "file:///" + urlPath
	} else {
		// Unix-style absolute path
		migrationsURL = "file://" + urlPath
	}

	// Detect driver name from config
	cfg := config.NewConfig()
	driverName := cfg.GetDriver()
	if driverName == "sqlite" || driverName == "sqlite3" {
		driverName = "sqlite3" // golang-migrate uses "sqlite3" as the driver name
	} else {
		driverName = "postgres"
	}

	// Try to create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		driverName,
		db,
	)
	if err != nil {
		// On Windows, try alternative URL formats if the standard one fails
		if len(urlPath) >= 2 && urlPath[1] == ':' {
			// Try format: file://E:/path (two slashes instead of three)
			altURL := "file://" + urlPath
			m, err = migrate.NewWithDatabaseInstance(altURL, driverName, db)
			if err != nil {
				// Try format: file:///E|/path (pipe notation)
				pipePath := "/" + string(urlPath[0]) + "|" + urlPath[2:]
				altURL2 := "file://" + pipePath
				m, err = migrate.NewWithDatabaseInstance(altURL2, driverName, db)
				if err != nil {
					// Last resort: try using relative path from current working directory
					wd, wdErr := os.Getwd()
					if wdErr == nil {
						relPath, relErr := filepath.Rel(wd, absPath)
						if relErr == nil && !filepath.IsAbs(relPath) {
							relURL := "file://" + filepath.ToSlash(relPath)
							m, err = migrate.NewWithDatabaseInstance(relURL, driverName, db)
						}
					}
					if err != nil {
						return nil, fmt.Errorf("failed to create migrate instance with URL %q (tried alternatives): %w", migrationsURL, err)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("failed to create migrate instance: %w", err)
		}
	}

	return &Executor{
		migrate:        m,
		migrationsPath: migrationsPath,
	}, nil
}

// Migrate applies all pending migrations
func (e *Executor) Migrate(ctx context.Context) error {
	// Check current version before migrating
	currentVersion, dirty, err := e.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	// Apply migrations
	if err := e.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			// No pending migrations - this is fine
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// MigrateTo applies migrations up to a specific version
func (e *Executor) MigrateTo(ctx context.Context, version uint) error {
	// Check current version
	currentVersion, dirty, err := e.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("cannot migrate: database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	// Check if target version is valid
	if err == migrate.ErrNilVersion {
		currentVersion = 0
	}

	if version < currentVersion {
		return fmt.Errorf("target version %d is less than current version %d. Use RollbackTo() to rollback", version, currentVersion)
	}

	// Apply migrations
	if err := e.migrate.Migrate(version); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}
	return nil
}

// Rollback rolls back the last migration
func (e *Executor) Rollback(ctx context.Context) error {
	// Check current version before rolling back
	currentVersion, dirty, err := e.migrate.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return fmt.Errorf("no migrations to rollback: database is at version 0")
		}
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("cannot rollback: database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	if currentVersion == 0 {
		return fmt.Errorf("no migrations to rollback: database is at version 0")
	}

	// Rollback one step
	if err := e.migrate.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// RollbackTo rolls back to a specific version
func (e *Executor) RollbackTo(ctx context.Context, version uint) error {
	// Check current version
	currentVersion, dirty, err := e.migrate.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return fmt.Errorf("cannot rollback: database is at version 0")
		}
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("cannot rollback: database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	if version >= currentVersion {
		return fmt.Errorf("target version %d is greater than or equal to current version %d. Use MigrateTo() to migrate forward", version, currentVersion)
	}

	// Rollback to target version
	if err := e.migrate.Migrate(version); err != nil {
		return fmt.Errorf("failed to rollback to version %d: %w", version, err)
	}
	return nil
}

// Version returns the current migration version
func (e *Executor) Version(ctx context.Context) (version uint, dirty bool, err error) {
	version, dirty, err = e.migrate.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Force sets a migration version and marks it as clean (for dirty state recovery)
// WARNING: Use with caution - only after manually fixing a failed migration
func (e *Executor) Force(ctx context.Context, version uint) error {
	if err := e.migrate.Force(int(version)); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	return nil
}

// Close closes the migration executor
func (e *Executor) Close() error {
	sourceErr, _ := e.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}
	return nil
}

// ValidatePendingMigrations validates pending migrations before execution
func (e *Executor) ValidatePendingMigrations(ctx context.Context, currentVersion uint, validator *MigrationValidator) error {
	// Get all migration files
	pattern := filepath.Join(e.migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Find and validate pending migrations
	var pendingFiles []string
	for _, match := range matches {
		basename := filepath.Base(match)
		// Extract version from filename (e.g., "000001_create_users.up.sql" -> "000001")
		parts := strings.Split(basename, "_")
		if len(parts) == 0 {
			continue
		}

		versionStr := parts[0]
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			continue
		}

		// Check if this migration is pending
		if currentVersion == 0 || uint(version) > currentVersion {
			pendingFiles = append(pendingFiles, match)
		}
	}

	// Validate each pending migration
	for _, filePath := range pendingFiles {
		if err := validator.ValidateBeforeExecution(ctx, filePath); err != nil {
			return fmt.Errorf("validation failed for migration %s: %w", filepath.Base(filePath), err)
		}

		// Also validate migration pair (up and down)
		downPath := strings.Replace(filePath, ".up.sql", ".down.sql", 1)
		if err := validator.ValidateMigrationPair(filePath, downPath); err != nil {
			// Down migration validation is a warning, not an error
			// Log it but continue
		}
	}

	return nil
}
