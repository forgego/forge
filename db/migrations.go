package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/forgego/forge/db/migrate/checksum"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/forgego/forge/db/migrate/verify"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// ErrChecksumBaselineBlocked is returned when migration execution is blocked by a baseline verification failure.
var ErrChecksumBaselineBlocked = errors.New("checksum baseline verification blocked")

// verifyBaselineBeforeMigration checks the baseline for all applied versions and blocks
// if any applied version is mismatched or missing against forge_migration_checksums.
// Unverified versions do not block. Dirty state and lookup failures do not block here;
// the caller reports dirty state separately. Rollback and Force never call this.
// It uses the single hashing implementation in forge/db/migrate/checksum and the
// runner's migration file index, avoiding a dependency on the execute package's
// recovery helper. Returns an error wrapping ErrChecksumBaselineBlocked if blocked.
func (mr *MigrationRunner) verifyBaselineBeforeMigration(ctx context.Context, executor checksumExecutor) error {
	// 1. Current applied version from schema_migrations.
	var currentVersion uint
	var dirty bool
	rows, err := executor.QueryContext(ctx, `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`)
	if err != nil {
		// No applied versions (or unreadable state): nothing to verify.
		return nil
	}
	if !rows.Next() {
		_ = rows.Close()
		return nil
	}
	if err := rows.Scan(&currentVersion, &dirty); err != nil {
		_ = rows.Close()
		return nil
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil
	}
	// Close before any further query: on PostgreSQL the executor may be a
	// single dedicated *sql.Conn which cannot run a second query while rows
	// are still open ("conn busy").
	if err := rows.Close(); err != nil {
		return nil
	}
	if dirty || currentVersion == 0 {
		return nil
	}

	// 2. Recorded baseline checksums for applied versions.
	type baselineRow struct {
		upSha256   string
		downSha256 sql.NullString
	}
	baselineRows := make(map[uint]baselineRow)
	baselineQuery, err := executor.QueryContext(ctx, `SELECT version, up_sha256, down_sha256 FROM forge_migration_checksums WHERE version <= $1`, currentVersion)
	if err != nil {
		return fmt.Errorf("query migration checksum baseline: %w", err)
	}
	for baselineQuery.Next() {
		var v uint
		var row baselineRow
		if err := baselineQuery.Scan(&v, &row.upSha256, &row.downSha256); err != nil {
			_ = baselineQuery.Close()
			return fmt.Errorf("scan migration checksum baseline: %w", err)
		}
		baselineRows[v] = row
	}
	if err := baselineQuery.Err(); err != nil {
		_ = baselineQuery.Close()
		return fmt.Errorf("read migration checksum baseline: %w", err)
	}
	if err := baselineQuery.Close(); err != nil {
		return fmt.Errorf("close migration checksum baseline rows: %w", err)
	}

	// 3. Applied versions: every indexed file version <= current, plus the current
	// version itself even if its file is gone, plus recorded baseline versions.
	index := mr.sourceIndex
	if index == nil {
		var err error
		index, err = indexMigrationFiles(mr.migrationsPath)
		if err != nil {
			return nil
		}
	}
	appliedSet := map[uint]bool{currentVersion: true}
	for v := range index {
		if v <= currentVersion {
			appliedSet[v] = true
		}
	}
	for v := range baselineRows {
		if v <= currentVersion {
			appliedSet[v] = true
		}
	}

	// 4. Compare each applied version against its file hashes and baseline row.
	appliedVersions := make([]uint, 0, len(appliedSet))
	for v := range appliedSet {
		appliedVersions = append(appliedVersions, v)
	}
	sort.Slice(appliedVersions, func(i, j int) bool { return appliedVersions[i] < appliedVersions[j] })
	for _, v := range appliedVersions {
		upHash, downHashPtr, fileErr := index.checksums(v)
		if fileErr != nil {
			if errors.Is(fileErr, checksum.ErrUpFileNotFound) || errors.Is(fileErr, os.ErrNotExist) {
				return fmt.Errorf("%w: version %d is missing (.up.sql file for version %d missing)", ErrChecksumBaselineBlocked, v, v)
			}
			return nil
		}
		row, hasRow := baselineRows[v]
		if !hasRow {
			continue
		}
		upMatches := upHash == strings.TrimSpace(row.upSha256)
		dbHasDown := row.downSha256.Valid && strings.TrimSpace(row.downSha256.String) != ""
		fileHasDown := downHashPtr != nil
		downMatches := true
		if dbHasDown || fileHasDown {
			if dbHasDown && fileHasDown {
				downMatches = strings.TrimSpace(row.downSha256.String) == *downHashPtr
			} else {
				downMatches = false
			}
		}
		if !upMatches || !downMatches {
			mismatchDetail := "up hash mismatch"
			if !upMatches && !downMatches {
				mismatchDetail = "up and down hash mismatch"
			} else if upMatches {
				mismatchDetail = "down hash mismatch"
			}
			return fmt.Errorf("%w: version %d is mismatched (%s)", ErrChecksumBaselineBlocked, v, mismatchDetail)
		}
	}
	return nil
}

// MigrationRunner handles migration execution using golang-migrate
type MigrationRunner struct {
	db             *DB
	migrate        *migrate.Migrate
	migrationsPath string
	sourceIndex    migrationFileIndex
}

// NewMigrationRunner creates a new migration runner (package-level function)
func NewMigrationRunner(db *DB, migrationsPath string) (*MigrationRunner, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Validate database connection is still open
	if db.DB == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Ping the database to ensure connection is still valid
	if err := db.DB.Ping(); err != nil {
		return nil, fmt.Errorf("database connection is closed or invalid: %w", err)
	}

	if db.Driver == "" {
		return nil, fmt.Errorf("database driver is unknown")
	}
	driverName := db.Driver

	var driver database.Driver
	var err error

	if driverName == "sqlite" || driverName == "sqlite3" {
		driver, err = sqlite3.WithInstance(db.DB, &sqlite3.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create sqlite driver: %w", err)
		}
		driverName = "sqlite3" // golang-migrate uses "sqlite3" as the driver name
	} else if driverName == "postgres" || driverName == "postgresql" {
		if db.Stats().MaxOpenConnections == 1 {
			return nil, fmt.Errorf("PostgreSQL migrations need at least 2 connections")
		}
		driver, err = postgres.WithInstance(db.DB, &postgres.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres driver: %w", err)
		}
		driverName = "postgres"
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", driverName)
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

	sourceIndex, err := indexMigrationFiles(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to index migration files: %w", err)
	}

	return &MigrationRunner{
		db:             db,
		migrate:        m,
		migrationsPath: absPath,
		sourceIndex:    sourceIndex,
	}, nil
}

// Migrate applies all pending migrations
func (mr *MigrationRunner) Migrate(ctx context.Context) error {
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error { return mr.migrateWithChecksums(ctx, executor) })
}

func (mr *MigrationRunner) migrateWithChecksums(ctx context.Context, executor checksumExecutor) error {
	// Check current version before migrating
	currentVersion, dirty, err := mr.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	// Verify baseline before migrating: block if any applied version is mismatched or missing
	if err := mr.verifyBaselineBeforeMigration(ctx, executor); err != nil {
		return err
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
	if err := mr.applyChecksumSteps(ctx, executor, nil); err != nil {
		if err == migrate.ErrNoChange {
			// No pending migrations - this is fine
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}

// Up applies all pending migrations (alias for Migrate)
func (mr *MigrationRunner) Up(ctx context.Context) error {
	return mr.Migrate(ctx)
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
		version, err := strconv.ParseUint(versionStr, 10, strconv.IntSize-1)
		if err != nil || version > math.MaxUint {
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
		version, err := strconv.ParseUint(versionStr, 10, strconv.IntSize-1)
		if err != nil || version > math.MaxUint {
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
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error { return mr.migrateToWithChecksums(ctx, executor, version) })
}

func (mr *MigrationRunner) migrateToWithChecksums(ctx context.Context, executor checksumExecutor, version uint) error {
	// Check current version
	currentVersion, dirty, err := mr.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to check current migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("cannot migrate: database is in a dirty state (version %d). Use Force() to resolve or manually fix the issue", currentVersion)
	}

	// Verify baseline before migrating: block if any applied version is mismatched or missing
	if err := mr.verifyBaselineBeforeMigration(ctx, executor); err != nil {
		return err
	}

	// Check if target version is valid
	if err == migrate.ErrNilVersion {
		currentVersion = 0
	}

	if version < currentVersion {
		return fmt.Errorf("target version %d is less than current version %d. Use RollbackTo() to rollback", version, currentVersion)
	}

	// Apply migrations
	if err := mr.applyChecksumSteps(ctx, executor, &version); err != nil {
		if err == migrate.ErrNoChange {
			return nil
		}
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}
	return nil
}

// Rollback rolls back the last migration
func (mr *MigrationRunner) Rollback(ctx context.Context) error {
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error { return mr.rollbackWithChecksums(ctx, executor) })
}

func (mr *MigrationRunner) rollbackWithChecksums(ctx context.Context, executor checksumExecutor) error {
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
	if err := mr.rollbackChecksumStep(ctx, executor); err != nil {
		if err == migrate.ErrNoChange {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	return nil
}

// RollbackSteps rolls back a specified number of migration steps
func (mr *MigrationRunner) RollbackSteps(ctx context.Context, steps int) error {
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error { return mr.rollbackStepsWithChecksums(ctx, executor, steps) })
}

func (mr *MigrationRunner) rollbackStepsWithChecksums(ctx context.Context, executor checksumExecutor, steps int) error {
	if steps <= 0 {
		return fmt.Errorf("steps must be greater than 0, got %d", steps)
	}

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

	// Rollback n steps
	for i := 0; i < steps; i++ {
		if err := mr.rollbackChecksumStep(ctx, executor); err != nil {
			if err == migrate.ErrNoChange {
				return fmt.Errorf("no migrations to rollback")
			}
			return fmt.Errorf("failed to rollback migration: %w", err)
		}
	}
	return nil
}

// RollbackTo rolls back to a specific version
func (mr *MigrationRunner) RollbackTo(ctx context.Context, version uint) error {
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error { return mr.rollbackToWithChecksums(ctx, executor, version) })
}

func (mr *MigrationRunner) rollbackToWithChecksums(ctx context.Context, executor checksumExecutor, version uint) error {
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

	// Validate target version against runner's source snapshot before stepping
	index := mr.sourceIndex
	if index == nil {
		var err error
		index, err = indexMigrationFiles(mr.migrationsPath)
		if err != nil {
			return err
		}
	}
	if _, ok := index[version]; !ok {
		return fmt.Errorf("cannot rollback to version %d: target version missing from migration source", version)
	}
	if _, _, err := index.checksums(version); err != nil {
		return fmt.Errorf("cannot rollback to version %d: %w", version, err)
	}

	// Rollback to target version
	for currentVersion > version {
		if err := mr.rollbackChecksumStep(ctx, executor); err != nil {
			return fmt.Errorf("failed to rollback to version %d: %w", version, err)
		}
		currentVersion, _, err = mr.migrate.Version()
		if err != nil {
			return err
		}
	}
	// Guard against overshoot: the runner's golang-migrate source snapshot may not
	// contain the target version (e.g., file added after runner construction), causing
	// the loop to exit at a version < target. Require exact match.
	if currentVersion != version {
		return fmt.Errorf("rollback ended at version %d, expected target %d — target version missing from migration source", currentVersion, version)
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
		return fallbackDetailedMigrationStatus(status), nil
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

func fallbackDetailedMigrationStatus(status *MigrationStatus) *DetailedMigrationStatus {
	fallbackStatus := "OK"
	if status != nil && status.Dirty {
		fallbackStatus = "DIRTY"
	}

	current := "0"
	if status != nil {
		current = fmt.Sprintf("%d", status.Version)
	}

	return &DetailedMigrationStatus{
		Current: current,
		Status:  fallbackStatus,
		Dirty:   status != nil && status.Dirty,
	}
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
	if version > math.MaxInt {
		return fmt.Errorf("migration version exceeds maximum supported value")
	}
	forced := int(version)
	return mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error {
		if err := mr.migrate.Force(forced); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
		return mr.reconcileMigrationChecksums(ctx, executor)
	})
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

// AdoptChecksumBaseline adopts existing applied migrations into the checksum baseline table.
// Under the same lock helper used for applying, it creates the table if missing and inserts rows
// for applied versions that have no row, using current file hashes. It never overwrites an existing row.
// Returns the slice of adopted versions. If a migration file is missing, it returns an error naming the version.
func (mr *MigrationRunner) AdoptChecksumBaseline(ctx context.Context) ([]uint, error) {
	var adopted []uint
	err := mr.withMigrationChecksums(ctx, func(executor checksumExecutor) error {
		currentVersion, dirty, err := mr.migrate.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				return nil
			}
			return fmt.Errorf("failed to get migration version: %w", err)
		}
		if dirty {
			return fmt.Errorf("database is in a dirty state (version %d). Resolve before adopting baseline", currentVersion)
		}
		if currentVersion == 0 {
			return nil
		}

		// Query existing versions in forge_migration_checksums
		rows, err := executor.QueryContext(ctx, "SELECT version FROM forge_migration_checksums")
		if err != nil {
			return fmt.Errorf("failed to query existing checksums: %w", err)
		}
		defer rows.Close()

		existing := make(map[uint]bool)
		for rows.Next() {
			var v uint
			if err := rows.Scan(&v); err != nil {
				return fmt.Errorf("failed to scan existing checksum version: %w", err)
			}
			existing[v] = true
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("error reading existing checksums: %w", err)
		}

		// Determine applied versions:
		// every migration file version <= current schema_migrations version,
		// plus current version itself even if its file is gone.
		appliedSet := make(map[uint]bool)
		appliedSet[currentVersion] = true

		index, err := indexMigrationFiles(mr.migrationsPath)
		if err != nil {
			return err
		}
		for v := range index {
			if v <= currentVersion {
				appliedSet[v] = true
			}
		}

		// Also check existing in forge_migration_checksums <= currentVersion
		for v := range existing {
			if v <= currentVersion {
				appliedSet[v] = true
			}
		}

		var sortedApplied []uint
		for v := range appliedSet {
			sortedApplied = append(sortedApplied, v)
		}
		sort.Slice(sortedApplied, func(i, j int) bool { return sortedApplied[i] < sortedApplied[j] })

		for _, v := range sortedApplied {
			upHash, downHash, err := index.checksums(v)
			if err != nil {
				return fmt.Errorf("cannot adopt baseline: migration file for version %d is missing: %w", v, err)
			}

			if existing[v] {
				continue
			}

			_, err = executor.ExecContext(ctx,
				`INSERT INTO forge_migration_checksums (version, up_sha256, down_sha256) VALUES ($1, $2, $3)`,
				v, upHash, downHash,
			)
			if err != nil {
				return fmt.Errorf("failed to insert checksum for version %d: %w", v, err)
			}
			adopted = append(adopted, v)
		}

		return nil
	})
	return adopted, err
}
