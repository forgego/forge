package execute

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	migrate "github.com/golang-migrate/migrate/v4"
)

// DetailedStatus provides detailed migration status information
type DetailedStatus struct {
	Current    string
	Next       string
	Applied    []MigrationInfo
	Pending    []MigrationInfo
	OutOfOrder []MigrationInfo
	Status     string // "OK", "PENDING", "DIRTY"
	Error      string
}

// MigrationInfo contains information about a single migration
type MigrationInfo struct {
	Version   string
	Name      string
	Applied   bool
	Checksum  string
	AppliedAt string
}

// StatusReporter collects detailed migration status
type StatusReporter struct {
	migrationsPath string
	migrate        *migrate.Migrate
	db             *sql.DB
}

// NewStatusReporter creates a new status reporter
func NewStatusReporter(migrationsPath string, m *migrate.Migrate, db *sql.DB) *StatusReporter {
	return &StatusReporter{
		migrationsPath: migrationsPath,
		migrate:        m,
		db:             db,
	}
}

// GetDetailedStatus returns detailed migration status
func (r *StatusReporter) GetDetailedStatus(ctx context.Context) (*DetailedStatus, error) {
	status := &DetailedStatus{
		Applied:    []MigrationInfo{},
		Pending:    []MigrationInfo{},
		OutOfOrder: []MigrationInfo{},
		Status:     "PENDING",
	}

	// Get current version
	var version uint
	var dirty bool
	var err error
	if r.migrate == nil {
		status.Current = "Unknown (migration engine unavailable)"
		status.Error = "Migration engine unavailable - detailed DB version status could not be determined"
	} else {
		version, dirty, err = r.migrate.Version()
		if err != nil && err != migrate.ErrNilVersion {
			return nil, fmt.Errorf("failed to get migration version: %w", err)
		}

		if err == migrate.ErrNilVersion {
			status.Current = "No migration applied yet"
		} else {
			status.Current = fmt.Sprintf("%d", version)
			if dirty {
				status.Status = "DIRTY"
				status.Error = "Migration is in a dirty state - manual intervention required"
			} else {
				status.Status = "OK"
			}
		}
	}

	// Get all migration files
	allMigrations, err := r.getAllMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations from database
	appliedVersions, err := r.getAppliedVersions(ctx)
	if err != nil {
		// If we can't get applied versions, assume none are applied
		appliedVersions = make(map[uint]bool)
	}
	appliedVersions = mergeAppliedVersions(version, dirty, appliedVersions)

	// Categorize migrations
	for _, mig := range allMigrations {
		migVersion, _ := parseVersion(mig.Version)
		isApplied := appliedVersions[migVersion]

		info := MigrationInfo{
			Version: mig.Version,
			Name:    mig.Name,
			Applied: isApplied,
		}

		if isApplied {
			status.Applied = append(status.Applied, info)
		} else {
			// Any unapplied migration lower than the current DB version is out-of-order.
			if isOutOfOrderMigration(migVersion, version, isApplied) {
				status.OutOfOrder = append(status.OutOfOrder, info)
			} else {
				status.Pending = append(status.Pending, info)
			}
		}
	}

	// Sort applied migrations
	sort.Slice(status.Applied, func(i, j int) bool {
		vi, _ := parseVersion(status.Applied[i].Version)
		vj, _ := parseVersion(status.Applied[j].Version)
		return vi < vj
	})

	// Sort pending migrations
	sort.Slice(status.Pending, func(i, j int) bool {
		vi, _ := parseVersion(status.Pending[i].Version)
		vj, _ := parseVersion(status.Pending[j].Version)
		return vi < vj
	})

	// Set next migration
	if len(status.Pending) > 0 {
		status.Next = status.Pending[0].Version
	} else {
		status.Next = "Already at latest version"
	}

	return status, nil
}

// MigrationFile represents a migration file
type MigrationFile struct {
	Version string
	Name    string
	Path    string
}

// getAllMigrations gets all migration files
func (r *StatusReporter) getAllMigrations() ([]MigrationFile, error) {
	pattern := filepath.Join(r.migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var migrations []MigrationFile
	for _, match := range matches {
		basename := filepath.Base(match)
		// Extract version and name from filename like "000001_create_users.up.sql"
		parts := strings.Split(basename, "_")
		if len(parts) < 2 {
			continue
		}
		version := parts[0]
		if _, err := parseVersion(version); err != nil {
			// Ignore malformed migration files with non-numeric version prefixes.
			continue
		}
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		migrations = append(migrations, MigrationFile{
			Version: version,
			Name:    name,
			Path:    match,
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		vi, _ := parseVersion(migrations[i].Version)
		vj, _ := parseVersion(migrations[j].Version)
		return vi < vj
	})

	return migrations, nil
}

// getAppliedVersions gets applied migration versions from database
func (r *StatusReporter) getAppliedVersions(ctx context.Context) (map[uint]bool, error) {
	applied := make(map[uint]bool)

	// If no database connection, fall back to core.Version()
	if r.db == nil {
		if r.migrate == nil {
			return applied, nil
		}
		version, _, err := r.migrate.Version()
		if err == nil {
			applied[version] = true
		}
		return applied, nil
	}

	// Query schema_migrations table (golang-migrate's internal table)
	// The table structure is: version (bigint), dirty (boolean)
	query := `SELECT version FROM schema_migrations WHERE dirty = false ORDER BY version`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		// If table doesn't exist or query fails, fall back to core.Version()
		if r.migrate == nil {
			return applied, nil
		}
		version, _, err := r.migrate.Version()
		if err == nil {
			applied[version] = true
		}
		return applied, nil
	}
	defer rows.Close()

	for rows.Next() {
		var version uint
		if err := rows.Scan(&version); err != nil {
			continue
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return applied, fmt.Errorf("error reading applied versions: %w", err)
	}

	return applied, nil
}

// parseVersion parses a version string to uint
func parseVersion(versionStr string) (uint, error) {
	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(version), nil
}

// mergeAppliedVersions merges explicit applied versions with inferred history from current version.
// golang-migrate stores only the current version in schema_migrations, so older applied versions must
// be inferred for accurate status output.
func mergeAppliedVersions(currentVersion uint, dirty bool, explicit map[uint]bool) map[uint]bool {
	merged := make(map[uint]bool, len(explicit))
	for version := range explicit {
		if dirty && version == currentVersion {
			// A dirty current version is not successfully applied yet.
			continue
		}
		merged[version] = true
	}

	if currentVersion == 0 {
		return merged
	}

	maxApplied := currentVersion
	if dirty && maxApplied > 0 {
		maxApplied--
	}

	for v := uint(1); v <= maxApplied; v++ {
		merged[v] = true
	}

	return merged
}

func isOutOfOrderMigration(migrationVersion, currentVersion uint, applied bool) bool {
	if applied {
		return false
	}
	if currentVersion == 0 {
		return false
	}
	return migrationVersion < currentVersion
}
