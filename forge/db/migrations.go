package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/forgego/forge/db/migrate/verify"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationRunner handles migration execution using golang-migrate
type MigrationRunner struct {
	db             *DB
	migrate        *migrate.Migrate
	migrationsPath string
}

// NewMigrationRunner creates a new migration runner (package-level function)
func NewMigrationRunner(db *DB, migrationsPath string) (*MigrationRunner, error) {
	// Validate database connection is still open
	if db.DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Ping the database to ensure connection is still valid
	if err := db.DB.Ping(); err != nil {
		return nil, fmt.Errorf("database connection is closed or invalid: %w", err)
	}

	// Detect driver from DSN or use config
	cfg := config.NewConfig()
	driverName := cfg.GetDriver()

	var driver database.Driver
	var err error

	if driverName == "sqlite" || driverName == "sqlite3" {
		driver, err = sqlite3.WithInstance(db.DB, &sqlite3.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create sqlite driver: %w", err)
		}
		driverName = "sqlite3" // golang-migrate uses "sqlite3" as the driver name
	} else {
		driver, err = postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres driver: %w", err)
		}
		driverName = "postgres"
	}

	// Always convert to absolute path to avoid issues with working directory changes
	// This ensures the migration files can be found regardless of where the code is called from
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
	// golang-migrate's file source driver expects file:// URLs
	// On Windows, there's a known issue with file:// URLs, so we try multiple formats
	var migrationsURL string
	if len(urlPath) >= 2 && urlPath[1] == ':' {
		// Windows path with drive letter (e.g., E:/path/to/migrations)
		// Try different formats that golang-migrate might accept on Windows
		// Format 1: file:///E:/path (three slashes - standard but may fail)
		// Format 2: file://E:/path (two slashes - alternative)
		// Format 3: file:///E|/path (pipe notation - sometimes works)

		// First try the standard format
		migrationsURL = "file:///" + urlPath

		// Test if this URL format works by checking if golang-migrate can parse it
		// If it fails, we'll catch it in the migrate.NewWithDatabaseInstance call
	} else {
		// Unix-style absolute path
		migrationsURL = "file://" + urlPath
	}

	// Try to create migrate instance
	// On Windows, if the standard file:// URL format fails, try alternatives
	m, err := migrate.NewWithDatabaseInstance(
		migrationsURL,
		driverName,
		driver,
	)
	if err != nil {
		// On Windows, try alternative URL formats if the standard one fails
		if len(urlPath) >= 2 && urlPath[1] == ':' {
			// Try format: file://E:/path (two slashes instead of three)
			altURL := "file://" + urlPath
			m, err = migrate.NewWithDatabaseInstance(altURL, driverName, driver)
			if err != nil {
				// Try format: file:///E|/path (pipe notation)
				pipePath := "/" + string(urlPath[0]) + "|" + urlPath[2:]
				altURL2 := "file://" + pipePath
				m, err = migrate.NewWithDatabaseInstance(altURL2, driverName, driver)
				if err != nil {
					// Last resort: try using relative path from current working directory
					wd, wdErr := os.Getwd()
					if wdErr == nil {
						relPath, relErr := filepath.Rel(wd, absPath)
						if relErr == nil && !filepath.IsAbs(relPath) {
							relURL := "file://" + filepath.ToSlash(relPath)
							m, err = migrate.NewWithDatabaseInstance(relURL, driverName, driver)
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

	return &MigrationRunner{
		db:             db,
		migrate:        m,
		migrationsPath: migrationsPath,
	}, nil
}

// Migrate applies all pending migrations
func (mr *MigrationRunner) Migrate(ctx context.Context) error {
	// Check current version before migrating
	currentVersion, dirty, err := mr.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	// Validate checksums for pending migrations before applying
	if err := mr.validatePendingMigrationChecksums(ctx, currentVersion); err != nil {
		return fmt.Errorf("checksum validation failed: %w", err)
	}

	// Validate pending migrations before applying
	if err := mr.validatePendingMigrations(ctx, currentVersion); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	// Apply migrations
	if err := mr.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			// No pending migrations - this is fine
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// validatePendingMigrations validates pending migrations before execution
func (mr *MigrationRunner) validatePendingMigrations(ctx context.Context, currentVersion uint) error {
	// Get all migration files
	pattern := filepath.Join(mr.migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Create migration validator
	validator := execute.NewMigrationValidator()

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

// validatePendingMigrationChecksums validates checksums for all pending migrations
func (mr *MigrationRunner) validatePendingMigrationChecksums(ctx context.Context, currentVersion uint) error {
	// Get all migration files
	pattern := filepath.Join(mr.migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Create checksum validator
	validator := verify.NewChecksumValidator(mr.migrationsPath)

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
		// If currentVersion is 0 (no migrations applied), all migrations are pending
		// Otherwise, only migrations with version > currentVersion are pending
		if currentVersion == 0 || uint(version) > currentVersion {
			pendingFiles = append(pendingFiles, match)
		}
	}

	// Validate checksums for pending migrations
	for _, filePath := range pendingFiles {
		// Validate that file exists and is readable by calculating checksum
		_, err := validator.CalculateChecksum(filePath)
		if err != nil {
			return fmt.Errorf("failed to validate migration file %s: %w", filepath.Base(filePath), err)
		}

		// Also validate corresponding down migration exists
		downPath := strings.Replace(filePath, ".up.sql", ".down.sql", 1)
		if _, err := os.Stat(downPath); os.IsNotExist(err) {
			// Down migration is optional, but log a warning
			// For now, we'll allow it but could make it strict
		} else if err == nil {
			// Validate down migration checksum if it exists
			_, err := validator.CalculateChecksum(downPath)
			if err != nil {
				return fmt.Errorf("failed to validate down migration file %s: %w", filepath.Base(downPath), err)
			}
		}
	}

	return nil
}

// MigrateTo applies migrations up to a specific version
func (mr *MigrationRunner) MigrateTo(ctx context.Context, version uint) error {
	// Check current version
	currentVersion, dirty, err := mr.migrate.Version()
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
	if err := mr.migrate.Migrate(version); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}
	return nil
}

// Rollback rolls back the last migration
func (mr *MigrationRunner) Rollback(ctx context.Context) error {
	// Check current version before rolling back
	currentVersion, dirty, err := mr.migrate.Version()
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
	if err := mr.migrate.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// RollbackTo rolls back to a specific version
func (mr *MigrationRunner) RollbackTo(ctx context.Context, version uint) error {
	// Check current version
	currentVersion, dirty, err := mr.migrate.Version()
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
	if err := mr.migrate.Migrate(version); err != nil {
		return fmt.Errorf("failed to rollback to version %d: %w", version, err)
	}
	return nil
}

// Version returns the current migration version
func (mr *MigrationRunner) Version(ctx context.Context) (version uint, dirty bool, err error) {
	version, dirty, err = mr.migrate.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Status returns migration status information
func (mr *MigrationRunner) Status(ctx context.Context) (*MigrationStatus, error) {
	version, dirty, err := mr.Version(ctx)
	if err != nil {
		return nil, err
	}

	return &MigrationStatus{
		Version: version,
		Dirty:   dirty,
	}, nil
}

// GetDetailedStatus returns detailed migration status
func (mr *MigrationRunner) GetDetailedStatus(ctx context.Context) (*DetailedMigrationStatus, error) {
	// Use StatusReporter for detailed status
	// Pass database connection so it can query schema_migrations table
	reporter := execute.NewStatusReporter(mr.migrationsPath, mr.migrate, mr.db.DB)
	detailed, err := reporter.GetDetailedStatus(ctx)
	if err != nil {
		// Fallback to basic status if detailed status fails
		status, statusErr := mr.Status(ctx)
		if statusErr != nil {
			return nil, fmt.Errorf("failed to get status: %w", statusErr)
		}
		return &DetailedMigrationStatus{
			Current: fmt.Sprintf("%d", status.Version),
			Status:  "OK",
			Dirty:   status.Dirty,
		}, nil
	}

	// Convert execute.DetailedStatus to db.DetailedMigrationStatus
	result := &DetailedMigrationStatus{
		Current:    detailed.Current,
		Next:       detailed.Next,
		Status:     detailed.Status,
		Dirty:      detailed.Status == "DIRTY",
		Error:      detailed.Error,
		Applied:    make([]string, len(detailed.Applied)),
		Pending:    make([]string, len(detailed.Pending)),
		OutOfOrder: make([]string, len(detailed.OutOfOrder)),
	}

	for i, mig := range detailed.Applied {
		result.Applied[i] = fmt.Sprintf("[%s] %s", mig.Version, mig.Name)
	}
	for i, mig := range detailed.Pending {
		result.Pending[i] = fmt.Sprintf("[%s] %s", mig.Version, mig.Name)
	}
	for i, mig := range detailed.OutOfOrder {
		result.OutOfOrder[i] = fmt.Sprintf("[%s] %s", mig.Version, mig.Name)
	}

	return result, nil
}

// MigrationStatus represents the status of migrations
type MigrationStatus struct {
	Version uint
	Dirty   bool
}

// DetailedMigrationStatus represents detailed migration status
type DetailedMigrationStatus struct {
	Current    string
	Next       string
	Applied    []string
	Pending    []string
	OutOfOrder []string
	Status     string
	Dirty      bool
	Error      string
}

// Force sets a migration version and marks it as clean (for dirty state recovery)
// WARNING: Use with caution - only after manually fixing a failed migration
func (mr *MigrationRunner) Force(ctx context.Context, version uint) error {
	if err := mr.migrate.Force(int(version)); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	return nil
}

// Close closes the migration runner
// Note: This closes the migrate instance, but NOT the underlying database connection.
// The database connection should be closed separately via db.Close()
func (mr *MigrationRunner) Close() error {
	sourceErr, _ := mr.migrate.Close()

	// The second return value from migrate.Close() refers to the migrate driver's internal connection,
	// not our underlying *sql.DB connection. We ignore it as it doesn't affect our DB.
	// The underlying database connection remains open and can be reused.

	if sourceErr != nil {
		return fmt.Errorf("failed to close migration source: %w", sourceErr)
	}

	// Verify database connection is still open after closing migrate instance
	if mr.db != nil && mr.db.DB != nil {
		if err := mr.db.DB.Ping(); err != nil {
			// Connection was closed - this shouldn't happen but log it
			return fmt.Errorf("database connection was closed unexpectedly: %w", err)
		}
	}

	return nil
}
