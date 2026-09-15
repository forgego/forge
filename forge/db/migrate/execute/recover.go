package execute

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/forgego/forge/db/migrate/checksum"
	"github.com/forgego/forge/db/migrate/verify"
)

// BaselineStatus represents the status of an applied migration compared against the baseline.
type BaselineStatus string

const (
	BaselineVerified   BaselineStatus = "verified"
	BaselineMismatched BaselineStatus = "mismatched"
	BaselineMissing    BaselineStatus = "missing"
	BaselineUnverified BaselineStatus = "unverified"

	BaselineStatusVerified   = BaselineVerified
	BaselineStatusMismatched = BaselineMismatched
	BaselineStatusMissing    = BaselineMissing
	BaselineStatusUnverified = BaselineUnverified
)

// BaselineEntry represents the baseline verification result for a single migration version.
type BaselineEntry struct {
	Version uint
	Status  BaselineStatus
	Detail  string
}

// BaselineReport represents the aggregate report of baseline verification.
type BaselineReport struct {
	Entries []BaselineEntry
}

// AllVerified returns true if every applied migration entry is verified.
func (r *BaselineReport) AllVerified() bool {
	if r == nil || len(r.Entries) == 0 {
		return true
	}
	for _, e := range r.Entries {
		if e.Status != BaselineStatusVerified {
			return false
		}
	}
	return true
}

// HasBlocking returns true when any entry is mismatched or missing.
func (r *BaselineReport) HasBlocking() bool {
	if r == nil {
		return false
	}
	for _, e := range r.Entries {
		if e.Status == BaselineStatusMismatched || e.Status == BaselineStatusMissing {
			return true
		}
	}
	return false
}

// Counts returns the count of entries by status.
func (r *BaselineReport) Counts() map[BaselineStatus]int {
	counts := map[BaselineStatus]int{
		BaselineStatusVerified:   0,
		BaselineStatusMismatched: 0,
		BaselineStatusMissing:    0,
		BaselineStatusUnverified: 0,
	}
	if r == nil {
		return counts
	}
	for _, e := range r.Entries {
		counts[e.Status]++
	}
	return counts
}

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
	Version    uint
	Applied    bool
	ErrorMsg   string
	Statements []string
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
		Version:  version,
		Applied:  true,
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

// VerifyAgainstBaseline verifies applied migrations against the baseline in forge_migration_checksums.
// It is strictly read-only and never creates or alters database tables.
func (r *Recovery) VerifyAgainstBaseline(ctx context.Context, migrationsDir string) (*BaselineReport, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	report := &BaselineReport{
		Entries: []BaselineEntry{},
	}

	// 1. Query schema_migrations for the current version.
	var currentVersion uint
	var dirty bool
	err := r.db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&currentVersion, &dirty)
	if err != nil {
		if err == sql.ErrNoRows || isTableNotExist(err) {
			// No migrations have been applied yet: "When there is no applied version, it is all verified."
			return report, nil
		}
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}

	if dirty {
		return nil, fmt.Errorf("database is in a dirty state (version %d)", currentVersion)
	}

	if currentVersion == 0 {
		return report, nil
	}

	// 2. Query forge_migration_checksums for recorded baseline checksums.
	type baselineRow struct {
		upSha256   string
		downSha256 sql.NullString
	}
	baselineRows := make(map[uint]baselineRow)

	rows, err := r.db.QueryContext(ctx, `SELECT version, up_sha256, down_sha256 FROM forge_migration_checksums WHERE version <= $1`, currentVersion)
	if err != nil {
		if !isTableNotExist(err) {
			return nil, fmt.Errorf("failed to query forge_migration_checksums: %w", err)
		}
		// If table does not exist, baselineRows remains empty (all versions will be unverified)
	} else {
		defer rows.Close()
		for rows.Next() {
			var v uint
			var row baselineRow
			if err := rows.Scan(&v, &row.upSha256, &row.downSha256); err != nil {
				return nil, fmt.Errorf("failed to scan migration checksum row: %w", err)
			}
			baselineRows[v] = row
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error reading migration checksum rows: %w", err)
		}
	}

	// 3. Collect applied versions:
	// "applied = every migration file version <= the current schema_migrations version,
	// plus the current version itself even if its file is gone"
	appliedSet := make(map[uint]bool)
	appliedSet[currentVersion] = true

	index, err := checksum.IndexMigrationFiles(migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for v := range index {
		if v <= currentVersion {
			appliedSet[v] = true
		}
	}

	// Also include any version present in forge_migration_checksums <= currentVersion
	for v := range baselineRows {
		if v <= currentVersion {
			appliedSet[v] = true
		}
	}

	sortedVersions := make([]uint, 0, len(appliedSet))
	for v := range appliedSet {
		sortedVersions = append(sortedVersions, v)
	}
	sort.Slice(sortedVersions, func(i, j int) bool { return sortedVersions[i] < sortedVersions[j] })

	// 4. Check each applied version against file and baseline row
	for _, v := range sortedVersions {
		upHash, downHashPtr, fileErr := index.Checksums(v)
		if fileErr != nil {
			if !errors.Is(fileErr, checksum.ErrUpFileNotFound) && !errors.Is(fileErr, os.ErrNotExist) {
				return nil, fileErr
			}
			// .up.sql file does not exist
			report.Entries = append(report.Entries, BaselineEntry{
				Version: v,
				Status:  BaselineStatusMissing,
				Detail:  fmt.Sprintf(".up.sql file for version %d missing", v),
			})
			continue
		}

		row, hasRow := baselineRows[v]
		if !hasRow {
			report.Entries = append(report.Entries, BaselineEntry{
				Version: v,
				Status:  BaselineStatusUnverified,
				Detail:  "no checksum baseline recorded",
			})
			continue
		}

		upMatches := (upHash == strings.TrimSpace(row.upSha256))

		dbHasDown := row.downSha256.Valid && strings.TrimSpace(row.downSha256.String) != ""
		fileHasDown := (downHashPtr != nil)
		downMatches := true
		if dbHasDown || fileHasDown {
			if dbHasDown && fileHasDown {
				downMatches = (strings.TrimSpace(row.downSha256.String) == *downHashPtr)
			} else {
				downMatches = false
			}
		}

		if upMatches && downMatches {
			report.Entries = append(report.Entries, BaselineEntry{
				Version: v,
				Status:  BaselineStatusVerified,
				Detail:  "",
			})
		} else {
			var detail string
			if !upMatches && !downMatches {
				detail = "up and down hash mismatch"
			} else if !upMatches {
				detail = "up hash mismatch"
			} else {
				detail = "down hash mismatch"
			}
			report.Entries = append(report.Entries, BaselineEntry{
				Version: v,
				Status:  BaselineStatusMismatched,
				Detail:  detail,
			})
		}
	}

	return report, nil
}

func isTableNotExist(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "no such table") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "undefined table")
}

// ValidateMigrationIntegrity computes migration file hashes.
// Note: this only computes hashes of files and does not compare with what was applied to the database.
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
		checksum, err := verify.NewChecksumValidator(migrationsDir).CalculateChecksum(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to compute checksum for %s: %w", name, err)
		}

		checksums[version] = checksum
	}

	return checksums, nil
}

// CompareChecksums compares file hashes with caller-supplied hashes.
// Note: this only computes hashes of files and does not compare with what was applied to the database.
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
	var statements []string
	start := 0
	n := len(sql)
	i := 0

	for i < n {
		switch {
		case sql[i] == '\'':
			// Single-quoted string: scan until closing unescaped single quote
			i++
			for i < n {
				if sql[i] == '\'' {
					if i+1 < n && sql[i+1] == '\'' {
						i += 2 // Escaped quote ''
					} else {
						i++
						break
					}
				} else {
					i++
				}
			}
		case sql[i] == '"':
			// Double-quoted identifier: scan until closing unescaped double quote
			i++
			for i < n {
				if sql[i] == '"' {
					if i+1 < n && sql[i+1] == '"' {
						i += 2 // Escaped quote ""
					} else {
						i++
						break
					}
				} else {
					i++
				}
			}
		case sql[i] == '-' && i+1 < n && sql[i+1] == '-':
			// Line comment: scan until newline
			i += 2
			for i < n && sql[i] != '\n' {
				i++
			}
		case sql[i] == '/' && i+1 < n && sql[i+1] == '*':
			// Block comment: scan until */
			i += 2
			closed := false
			for i+1 < n {
				if sql[i] == '*' && sql[i+1] == '/' {
					i += 2
					closed = true
					break
				}
				i++
			}
			if !closed {
				i = n
			}
		case sql[i] == '$':
			tag, ok := scanDollarTag(sql, i)
			if ok {
				i += len(tag)
				closed := false
				for i <= n-len(tag) {
					if sql[i:i+len(tag)] == tag {
						i += len(tag)
						closed = true
						break
					}
					i++
				}
				if !closed {
					i = n
				}
			} else {
				i++
			}
		case sql[i] == ';':
			stmt := strings.TrimSpace(sql[start:i])
			if stmt != "" {
				statements = append(statements, stmt)
			}
			i++
			start = i
		default:
			i++
		}
	}

	if start < n {
		stmt := strings.TrimSpace(sql[start:])
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// scanDollarTag checks if sql[i:] starts a PostgreSQL dollar quote tag ($$ or $tag$).
// Tag chars are letters, digits, underscore; must not start with a digit.
func scanDollarTag(sql string, i int) (string, bool) {
	n := len(sql)
	if i >= n || sql[i] != '$' {
		return "", false
	}
	if i+1 < n && sql[i+1] == '$' {
		return "$$", true
	}
	if i+1 >= n {
		return "", false
	}
	first := sql[i+1]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return "", false
	}
	j := i + 2
	for j < n {
		c := sql[j]
		if c == '$' {
			return sql[i : j+1], true
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return "", false
		}
		j++
	}
	return "", false
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
