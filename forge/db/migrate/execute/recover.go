package execute

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Recovery handles migration recovery operations
type Recovery struct {
	db *sql.DB
}

// NewRecovery creates a new recovery handler
func NewRecovery(db *sql.DB) *Recovery {
	return &Recovery{
		db: db,
	}
}

// RecoveryMigrationInfo represents basic migration information for recovery
type RecoveryMigrationInfo struct {
	Version uint
	Dirty   bool
}

// DirtyMigration represents a dirty migration that needs recovery
type DirtyMigration struct {
	Version     uint
	Applied     bool
	ErrorMsg    string
	Statements  []string
}

// RecoverDirtyState attempts to recover from a dirty migration state
func (r *Recovery) RecoverDirtyState(ctx context.Context, migrationsDir string) (*DirtyMigration, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// Query the schema_migrations table to check for dirty migrations
	var version uint
	var dirty bool
	
	query := `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query).Scan(&version, &dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no migrations found in schema_migrations table")
		}
		return nil, fmt.Errorf("failed to query migration status: %w", err)
	}

	if !dirty {
		return nil, nil // No dirty migration
	}

	// Found a dirty migration
	dirtyMigration := &DirtyMigration{
		Version: version,
		Applied: true,
		ErrorMsg: fmt.Sprintf("Migration %d is marked as dirty", version),
	}

	return dirtyMigration, nil
}

// MarkMigrationClean marks a migration as clean (not dirty)
func (r *Recovery) MarkMigrationClean(ctx context.Context, version uint) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `UPDATE schema_migrations SET dirty = false WHERE version = $1`
	_, err := r.db.ExecContext(ctx, query, version)
	if err != nil {
		return fmt.Errorf("failed to mark migration as clean: %w", err)
	}

	return nil
}

// GetDirtyMigrationInfo retrieves detailed information about a dirty migration
func (r *Recovery) GetDirtyMigrationInfo(ctx context.Context) (*RecoveryMigrationInfo, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	info := &RecoveryMigrationInfo{}
	query := `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`
	err := r.db.QueryRowContext(ctx, query).Scan(&info.Version, &info.Dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return &RecoveryMigrationInfo{Version: 0, Dirty: false}, nil
		}
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}

	return info, nil
}

// ValidateMigrationIntegrity validates that migration files haven't been modified
func (r *Recovery) ValidateMigrationIntegrity(migrationsDir string) (map[uint]string, error) {
	// Read all migration files and compute their checksums
	checksums := make(map[uint]string)
	
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse version from filename (format: YYYYMMDDHHMMSS_name.up.sql)
		name := file.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		var version uint
		_, err := fmt.Sscanf(name, "%d_", &version)
		if err != nil {
			continue
		}

		// Compute checksum
		filePath := filepath.Join(migrationsDir, name)
		checksum, err := computeFileChecksum(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to compute checksum for %s: %w", name, err)
		}

		checksums[version] = checksum
	}

	return checksums, nil
}

// computeFileChecksum computes SHA256 checksum of a file
func computeFileChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// CompareChecksums compares current file checksums with stored checksums
func (r *Recovery) CompareChecksums(migrationsDir string, storedChecksums map[uint]string) ([]uint, error) {
	currentChecksums, err := r.ValidateMigrationIntegrity(migrationsDir)
	if err != nil {
		return nil, err
	}

	var modified []uint
	for version, storedChecksum := range storedChecksums {
		currentChecksum, exists := currentChecksums[version]
		if !exists {
			modified = append(modified, version)
			continue
		}

		if currentChecksum != storedChecksum {
			modified = append(modified, version)
		}
	}

	return modified, nil
}

// GetRecoverySteps returns recovery steps for a failed migration
func (r *Recovery) GetRecoverySteps(version string, errorMsg string) []string {
	steps := []string{
		"1. Check the error message above for details",
		"2. Review the migration file: " + version,
		"3. Check database logs for detailed error information",
		"4. Verify database state matches expected state",
	}

	if strings.Contains(errorMsg, "constraint") {
		steps = append(steps, "5. Check for foreign key constraints that may need to be dropped first")
	}

	if strings.Contains(errorMsg, "column") {
		steps = append(steps, "5. Verify column doesn't already exist or has different type")
	}

	if strings.Contains(errorMsg, "table") {
		steps = append(steps, "5. Verify table doesn't already exist or has different structure")
	}

	steps = append(steps,
		"6. Fix the issue manually if needed",
		fmt.Sprintf("7. Mark migration as clean: UPDATE schema_migrations SET dirty = false WHERE version = %s", version),
		"8. Retry the migration",
	)

	return steps
}

// RollbackPartialMigration attempts to rollback a partially applied migration
func (r *Recovery) RollbackPartialMigration(ctx context.Context, version uint, downSQL string) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Start a transaction for rollback
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Split downSQL into individual statements
	statements := splitSQL(downSQL)

	// Execute each statement
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return fmt.Errorf("failed to execute down statement %d: %w\nStatement: %s", i+1, err, stmt)
		}
	}

	// Remove the migration from schema_migrations
	_, err = tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version)
	if err != nil {
		return fmt.Errorf("failed to remove migration from schema_migrations: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback transaction: %w", err)
	}

	return nil
}

// splitSQL splits SQL text into individual statements
func splitSQL(sql string) []string {
	// Simple split by semicolon
	// Note: This is a basic implementation and may not handle all cases
	// (e.g., semicolons in strings or comments)
	statements := strings.Split(sql, ";")
	
	result := make([]string, 0, len(statements))
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			result = append(result, stmt)
		}
	}
	
	return result
}

// ForceCleanState forces the database to a clean state (use with caution!)
func (r *Recovery) ForceCleanState(ctx context.Context, version uint) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `UPDATE schema_migrations SET dirty = false WHERE version = $1`
	result, err := r.db.ExecContext(ctx, query, version)
	if err != nil {
		return fmt.Errorf("failed to force clean state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no migration found with version %d", version)
	}

	return nil
}

// GetAppliedMigrations retrieves all applied migrations
func (r *Recovery) GetAppliedMigrations(ctx context.Context) ([]RecoveryMigrationInfo, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	query := `SELECT version, dirty FROM schema_migrations ORDER BY version ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	var migrations []RecoveryMigrationInfo
	for rows.Next() {
		var m RecoveryMigrationInfo
		if err := rows.Scan(&m.Version, &m.Dirty); err != nil {
			return nil, fmt.Errorf("failed to scan migration: %w", err)
		}
		migrations = append(migrations, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migrations: %w", err)
	}

	return migrations, nil
}
