package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate"
)

// MigrationInfo represents information about a migration
type MigrationInfo struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

// CreateMigrationFromModels generates migrations from model definitions
func CreateMigrationFromModels(t *testing.T, modelsDir, migrationsDir, migrationName string) error {
	gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to create migration generator: %w", err)
	}

	if err := gen.GenerateMigrations(migrationName); err != nil {
		return fmt.Errorf("failed to generate migrations: %w", err)
	}

	// Verify both up and down migration files were created
	migrationFiles, err := GetMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to verify migration files: %w", err)
	}

	// Check if the generated migration has both up and down files
	found := false
	for _, mig := range migrationFiles {
		if mig.Name == migrationName || strings.Contains(mig.Name, migrationName) {
			if mig.UpSQL == "" {
				return fmt.Errorf("up migration file missing for %s", migrationName)
			}
			if mig.DownSQL == "" {
				return fmt.Errorf("down migration file missing for %s", migrationName)
			}
			found = true
			break
		}
	}

	if !found && len(migrationFiles) > 0 {
		// Migration might have been generated with a different name (with version prefix)
		// Check the most recent migration
		latest := migrationFiles[len(migrationFiles)-1]
		if latest.UpSQL == "" || latest.DownSQL == "" {
			return fmt.Errorf("latest migration %s is missing up or down SQL", latest.Name)
		}
	}

	return nil
}

// ApplyMigrationSequence applies multiple migrations in order
// If runner is provided, it will be reused; otherwise a new one will be created (and NOT closed)
func ApplyMigrationSequence(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string, runner ...*db.MigrationRunner) error {
	var mr *db.MigrationRunner
	var err error
	shouldClose := false

	if len(runner) > 0 && runner[0] != nil {
		mr = runner[0]
	} else {
		mr, err = db.NewMigrationRunner(database, migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to create migration runner: %w", err)
		}
		shouldClose = false // Don't close - it closes the database connection
	}

	if err := mr.Migrate(ctx); err != nil {
		if shouldClose {
			mr.Close()
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Note: We don't close the runner here because it would close the database connection
	// The caller is responsible for managing the runner lifecycle
	return nil
}

// RollbackMigrationSequence rolls back migrations in reverse order
// If runner is provided, it will be reused; otherwise a new one will be created (and NOT closed)
func RollbackMigrationSequence(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string, steps int, runner ...*db.MigrationRunner) error {
	var mr *db.MigrationRunner
	var err error
	shouldClose := false

	if len(runner) > 0 && runner[0] != nil {
		mr = runner[0]
	} else {
		mr, err = db.NewMigrationRunner(database, migrationsPath)
		if err != nil {
			return fmt.Errorf("failed to create migration runner: %w", err)
		}
		shouldClose = false // Don't close - it closes the database connection
	}

	for i := 0; i < steps; i++ {
		if err := mr.Rollback(ctx); err != nil {
			if shouldClose {
				mr.Close()
			}
			return fmt.Errorf("failed to rollback migration step %d: %w", i+1, err)
		}
	}

	// Note: We don't close the runner here because it would close the database connection
	// The caller is responsible for managing the runner lifecycle
	return nil
}

// AssertMigrationState verifies migration state matches expected
// If runner is nil, a new runner will be created (and closed after use)
func AssertMigrationState(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string, expectedVersion uint, expectedDirty bool) {
	AssertMigrationStateWithRunner(ctx, t, database, migrationsPath, expectedVersion, expectedDirty, nil)
}

// AssertMigrationStateWithRunner verifies migration state matches expected using an optional runner
// If runner is nil, a new runner will be created (and closed after use)
func AssertMigrationStateWithRunner(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string, expectedVersion uint, expectedDirty bool, runner *db.MigrationRunner) {
	var err error
	shouldClose := false

	if runner == nil {
		// Ensure database connection is still valid
		if err := database.Ping(ctx); err != nil {
			t.Fatalf("database connection is closed or invalid: %v", err)
		}

		runner, err = db.NewMigrationRunner(database, migrationsPath)
		if err != nil {
			t.Fatalf("failed to create migration runner: %v", err)
		}
		shouldClose = true
	}

	if shouldClose {
		defer runner.Close()
	}

	version, dirty, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}

	if version != expectedVersion {
		t.Errorf("expected migration version %d, got %d", expectedVersion, version)
	}

	if dirty != expectedDirty {
		t.Errorf("expected dirty state %v, got %v", expectedDirty, dirty)
	}
}

// AssertMigrationFilesExist verifies that both up and down migration files exist for a given version
// NOTE: Currently unused but kept for potential future use
func AssertMigrationFilesExist(t *testing.T, migrationsDir, version string) {
	upPattern := filepath.Join(migrationsDir, version+"_*.up.sql")
	downPattern := filepath.Join(migrationsDir, version+"_*.down.sql")

	upMatches, err := filepath.Glob(upPattern)
	if err != nil {
		t.Fatalf("failed to glob up migration files: %v", err)
	}
	if len(upMatches) == 0 {
		t.Fatalf("up migration file not found for version %s (pattern: %s)", version, upPattern)
	}

	downMatches, err := filepath.Glob(downPattern)
	if err != nil {
		t.Fatalf("failed to glob down migration files: %v", err)
	}
	if len(downMatches) == 0 {
		t.Fatalf("down migration file not found for version %s (pattern: %s)", version, downPattern)
	}
}

// CreateUniqueTestDB generates a unique database name for a test
// This is a convenience wrapper around DefaultPostgresOptsWithTest
// NOTE: Currently unused but kept for potential future use
func CreateUniqueTestDB(t *testing.T, prefix string) string {
	testName := t.Name()
	if prefix != "" {
		testName = prefix + "_" + testName
	}
	opts := DefaultPostgresOptsWithTest(testName)
	return opts.DBName
}

// AssertSchemaMatchesModels verifies database schema matches model definitions
// This is a high-level check that validates tables, columns, and basic constraints exist
func AssertSchemaMatchesModels(ctx context.Context, t *testing.T, database *sql.DB, dialect string, expectedTables []string) {
	for _, tableName := range expectedTables {
		AssertTableExists(ctx, t, database, dialect, tableName)
	}
}

// GetMigrationFiles returns all migration files in a directory
func GetMigrationFiles(migrationsDir string) ([]MigrationInfo, error) {
	var migrations []MigrationInfo

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrationMap := make(map[string]*MigrationInfo)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}

		// Parse migration file name: {version}_{name}.{up|down}.sql
		var version, migrationName, direction string
		baseName := filepath.Base(name)
		ext := filepath.Ext(baseName) // .sql
		baseWithoutExt := baseName[:len(baseName)-len(ext)]

		// Split by . to get direction
		parts := filepath.Ext(baseWithoutExt) // .up or .down
		if parts == "" {
			continue
		}
		direction = parts[1:] // Remove the dot
		baseWithoutDirection := baseWithoutExt[:len(baseWithoutExt)-len(parts)]

		// Split by _ to get version and name
		underscoreIdx := strings.IndexRune(baseWithoutDirection, '_')

		if underscoreIdx == -1 {
			continue
		}

		version = baseWithoutDirection[:underscoreIdx]
		migrationName = baseWithoutDirection[underscoreIdx+1:]

		key := version + "_" + migrationName
		if migrationMap[key] == nil {
			migrationMap[key] = &MigrationInfo{
				Version: version,
				Name:    migrationName,
			}
		}

		// Read SQL content
		filePath := filepath.Join(migrationsDir, name)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if direction == "up" {
			migrationMap[key].UpSQL = string(content)
		} else if direction == "down" {
			migrationMap[key].DownSQL = string(content)
		}
	}

	// Convert map to slice
	for _, mig := range migrationMap {
		migrations = append(migrations, *mig)
	}

	return migrations, nil
}

// ValidateMigrationFiles validates that migration files are properly formatted
func ValidateMigrationFiles(t *testing.T, migrationsDir string) error {
	migrations, err := GetMigrationFiles(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, mig := range migrations {
		if mig.UpSQL == "" {
			return fmt.Errorf("migration %s_%s is missing up SQL", mig.Version, mig.Name)
		}
		if mig.DownSQL == "" {
			return fmt.Errorf("migration %s_%s is missing down SQL", mig.Version, mig.Name)
		}
		if len(mig.Version) != 6 {
			return fmt.Errorf("migration %s_%s has invalid version format (expected 6 digits, got %d)", mig.Version, mig.Name, len(mig.Version))
		}
	}

	return nil
}

// AssertMigrationCanRollback verifies that a migration can be rolled back successfully
func AssertMigrationCanRollback(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string) {
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		t.Fatalf("failed to create migration runner: %v", err)
	}
	defer runner.Close()

	// Get current version
	currentVersion, dirty, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}

	if dirty {
		t.Fatalf("database is in dirty state, cannot test rollback")
	}

	if currentVersion == 0 {
		t.Skip("no migrations to rollback")
	}

	// Rollback
	if err := runner.Rollback(ctx); err != nil {
		t.Fatalf("failed to rollback migration: %v", err)
	}

	// Verify version decreased
	newVersion, newDirty, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get migration version after rollback: %v", err)
	}

	if newDirty {
		t.Errorf("database is in dirty state after rollback")
	}

	if newVersion >= currentVersion {
		t.Errorf("expected version to decrease after rollback, got %d (was %d)", newVersion, currentVersion)
	}
}

// AssertMigrationSequence validates that migrations can be applied and rolled back in sequence
func AssertMigrationSequence(ctx context.Context, t *testing.T, database *db.DB, migrationsPath string, expectedFinalVersion uint) {
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		t.Fatalf("failed to create migration runner: %v", err)
	}
	defer runner.Close()

	// Apply all migrations
	if err := runner.Migrate(ctx); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Verify final version
	version, dirty, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("failed to get migration version: %v", err)
	}

	if dirty {
		t.Errorf("database is in dirty state after migration")
	}

	if version != expectedFinalVersion {
		t.Errorf("expected final version %d, got %d", expectedFinalVersion, version)
	}

	// Rollback all migrations
	for version > 0 {
		if err := runner.Rollback(ctx); err != nil {
			t.Fatalf("failed to rollback migration at version %d: %v", version, err)
		}

		newVersion, newDirty, err := runner.Version(ctx)
		if err != nil {
			t.Fatalf("failed to get migration version after rollback: %v", err)
		}

		if newDirty {
			t.Errorf("database is in dirty state after rollback at version %d", version)
		}

		if newVersion >= version {
			t.Errorf("expected version to decrease after rollback, got %d (was %d)", newVersion, version)
		}

		version = newVersion
	}

	// Verify we're back at version 0
	if version != 0 {
		t.Errorf("expected to be at version 0 after rolling back all migrations, got %d", version)
	}
}
