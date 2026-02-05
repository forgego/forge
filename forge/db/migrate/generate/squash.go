package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Squasher squashes multiple migrations into a single migration
type Squasher struct {
	migrationsDir string
}

// NewSquasher creates a new migration squasher
func NewSquasher(migrationsDir string) *Squasher {
	return &Squasher{
		migrationsDir: migrationsDir,
	}
}

// SquashMigrations squashes migrations from startVersion to endVersion into a single migration
func (s *Squasher) SquashMigrations(startVersion, endVersion, newName string) error {
	// Find all migration files in range
	migrations, err := s.getMigrationsInRange(startVersion, endVersion)
	if err != nil {
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found in range")
	}

	// Read and combine up migrations
	var upSQLParts []string
	var downSQLParts []string

	for _, mig := range migrations {
		upContent, err := os.ReadFile(mig.UpPath)
		if err != nil {
			return fmt.Errorf("failed to read up migration %s: %w", mig.UpPath, err)
		}

		downContent, err := os.ReadFile(mig.DownPath)
		if err != nil {
			return fmt.Errorf("failed to read down migration %s: %w", mig.DownPath, err)
		}

		upSQLParts = append(upSQLParts, fmt.Sprintf("-- Migration: %s\n%s", mig.Name, string(upContent)))
		// Down migrations are reversed
		downSQLParts = append([]string{fmt.Sprintf("-- Migration: %s\n%s", mig.Name, string(downContent))}, downSQLParts...)
	}

	// Combine SQL
	combinedUpSQL := strings.Join(upSQLParts, "\n\n")
	combinedDownSQL := strings.Join(downSQLParts, "\n\n")

	// Get next version
	nextVersion, err := s.getNextVersion()
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}

	// Create new migration files
	newMigrationName := fmt.Sprintf("%s_%s", nextVersion, newName)
	upPath := filepath.Join(s.migrationsDir, fmt.Sprintf("%s.up.sql", newMigrationName))
	downPath := filepath.Join(s.migrationsDir, fmt.Sprintf("%s.down.sql", newMigrationName))

	if err := os.WriteFile(upPath, []byte(combinedUpSQL), 0644); err != nil {
		return fmt.Errorf("failed to write up migration: %w", err)
	}

	if err := os.WriteFile(downPath, []byte(combinedDownSQL), 0644); err != nil {
		return fmt.Errorf("failed to write down migration: %w", err)
	}

	// Note: Old migrations should be archived, not deleted
	// In production, you'd want to move them to an archive directory

	return nil
}

// MigrationFile represents a migration file pair
type MigrationFile struct {
	Name     string
	Version  string
	UpPath   string
	DownPath string
}

// getMigrationsInRange gets all migrations between start and end versions
func (s *Squasher) getMigrationsInRange(startVersion, endVersion string) ([]MigrationFile, error) {
	pattern := filepath.Join(s.migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var migrations []MigrationFile
	for _, match := range matches {
		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) < 2 {
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		// Check if in range
		if version >= startVersion && version <= endVersion {
			downPath := strings.Replace(match, ".up.sql", ".down.sql", 1)
			migrations = append(migrations, MigrationFile{
				Name:     name,
				Version:  version,
				UpPath:   match,
				DownPath: downPath,
			})
		}
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getNextVersion gets the next migration version
func (s *Squasher) getNextVersion() (string, error) {
	pattern := filepath.Join(s.migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	maxVersion := uint64(0)
	for _, match := range matches {
		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) < 1 {
			continue
		}

		versionStr := parts[0]
		version := parseVersion(versionStr)
		if version > maxVersion {
			maxVersion = version
		}
	}

	nextVersion := maxVersion + 1
	return fmt.Sprintf("%06d", nextVersion), nil
}

// parseVersion parses a version string to uint64
func parseVersion(versionStr string) uint64 {
	var result uint64
	for _, r := range versionStr {
		if r >= '0' && r <= '9' {
			result = result*10 + uint64(r-'0')
		}
	}
	return result
}
