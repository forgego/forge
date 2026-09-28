package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// createMigrationFiles creates empty migration files using sequential versioning.
// Returns paths to the created up and down migration files.
func createMigrationFiles(migrationsDir, name string) (upPath, downPath string, err error) {
	const seqDigits = 6 // Use 6-digit zero-padded sequential numbers

	// Ensure directory exists
	if err = os.MkdirAll(migrationsDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Find next sequential version
	version, err := nextSeqVersion(migrationsDir, seqDigits)
	if err != nil {
		return "", "", fmt.Errorf("failed to get next migration version: %w", err)
	}

	// Create file paths
	basename := fmt.Sprintf("%s_%s", version, name)
	upPath = filepath.Join(migrationsDir, fmt.Sprintf("%s.up.sql", basename))
	downPath = filepath.Join(migrationsDir, fmt.Sprintf("%s.down.sql", basename))

	// Check for duplicate version
	versionGlob := filepath.Join(migrationsDir, version+"_*.sql")
	matches, err := filepath.Glob(versionGlob)
	if err != nil {
		return "", "", fmt.Errorf("failed to check for existing migrations: %w", err)
	}
	if len(matches) > 0 {
		return "", "", fmt.Errorf("duplicate migration version: %s", version)
	}

	// Create empty up migration file
	upFile, err := os.OpenFile(upPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return "", "", fmt.Errorf("failed to create up migration file: %w", err)
	}
	if err = upFile.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close up migration file: %w", err)
	}

	// Create empty down migration file
	downFile, err := os.OpenFile(downPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		// Clean up up file if down file creation fails
		_ = os.Remove(upPath)
		return "", "", fmt.Errorf("failed to create down migration file: %w", err)
	}
	if err = downFile.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close down migration file: %w", err)
	}

	return upPath, downPath, nil
}

// nextSeqVersion finds the next sequential version number by scanning existing migration files.
func nextSeqVersion(migrationsDir string, seqDigits int) (string, error) {
	if seqDigits <= 0 {
		return "", fmt.Errorf("sequence digits must be positive")
	}

	// Find all migration files matching the pattern
	pattern := filepath.Join(migrationsDir, "*_*.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	nextSeq := uint64(1)

	if len(matches) > 0 {
		// Extract version numbers from filenames
		versions := make([]uint64, 0, len(matches))
		for _, filename := range matches {
			basename := filepath.Base(filename)
			idx := strings.Index(basename, "_")
			if idx < 1 {
				// Skip malformed filenames
				continue
			}

			versionStr := basename[0:idx]
			version, err := strconv.ParseUint(versionStr, 10, 64)
			if err != nil {
				// Skip non-numeric prefixes
				continue
			}
			versions = append(versions, version)
		}

		// Find maximum version
		if len(versions) > 0 {
			sort.Slice(versions, func(i, j int) bool {
				return versions[i] < versions[j]
			})
			nextSeq = versions[len(versions)-1] + 1
		}
	}

	// Format with zero-padding
	version := fmt.Sprintf("%0*d", seqDigits, nextSeq)

	// Check for overflow
	if len(version) > seqDigits {
		return "", fmt.Errorf("next sequence number %s too large, at most %d digits are allowed", version, seqDigits)
	}

	return version, nil
}

// detectConflicts detects migration conflicts (multiple migrations with the same parent)
func detectConflicts(migrationsDir string) ([]Conflict, error) {
	// Find all migration files
	pattern := filepath.Join(migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Parse versions and build dependency graph
	versions := make(map[string]string) // version -> filename

	for _, match := range matches {
		basename := filepath.Base(match)
		idx := strings.Index(basename, "_")
		if idx < 1 {
			continue
		}
		version := basename[0:idx]
		versions[version] = basename
	}

	// Simple conflict detection: if we have multiple migrations with sequential versions
	// that don't follow a strict linear order, we have a conflict
	// For now, detect if there are gaps or branches
	var conflicts []Conflict

	// Sort versions
	var sortedVersions []string
	for v := range versions {
		sortedVersions = append(sortedVersions, v)
	}
	sort.Slice(sortedVersions, func(i, j int) bool {
		vi, _ := strconv.ParseUint(sortedVersions[i], 10, 64)
		vj, _ := strconv.ParseUint(sortedVersions[j], 10, 64)
		return vi < vj
	})

	// Check for branches (multiple migrations that could have the same parent)
	// This is a simplified check - in practice, we'd parse migration dependencies
	for i := 1; i < len(sortedVersions); i++ {
		prev, _ := strconv.ParseUint(sortedVersions[i-1], 10, 64)
		curr, _ := strconv.ParseUint(sortedVersions[i], 10, 64)

		// If there's a gap of more than 1, or if we detect branching
		if curr > prev+1 {
			// Potential conflict - multiple migrations created independently
			conflicts = append(conflicts, Conflict{
				Version1: sortedVersions[i-1],
				Version2: sortedVersions[i],
			})
		}
	}

	return conflicts, nil
}

// Conflict represents a migration conflict
type Conflict struct {
	Version1 string
	Version2 string
}

// createMergeMigration creates a merge migration that depends on both conflicting migrations
func createMergeMigration(migrationsDir, name string, conflict Conflict) (upPath, downPath string, err error) {
	const seqDigits = 6

	// Get next version after both conflicting versions
	v1, _ := strconv.ParseUint(conflict.Version1, 10, 64)
	v2, _ := strconv.ParseUint(conflict.Version2, 10, 64)
	maxVersion := v1
	if v2 > v1 {
		maxVersion = v2
	}

	version := fmt.Sprintf("%0*d", seqDigits, maxVersion+1)

	basename := fmt.Sprintf("%s_%s", version, name)
	upPath = filepath.Join(migrationsDir, fmt.Sprintf("%s.up.sql", basename))
	downPath = filepath.Join(migrationsDir, fmt.Sprintf("%s.down.sql", basename))

	// Create merge migration SQL with dependencies as comments
	upSQL := fmt.Sprintf(`-- Merge migration: resolves conflict between %s and %s
-- This migration depends on both:
--   - %s
--   - %s

-- Add your merge logic here
`, conflict.Version1, conflict.Version2, conflict.Version1, conflict.Version2)

	downSQL := fmt.Sprintf(`-- Rollback for merge migration
-- This reverses the merge between %s and %s

-- Add your rollback logic here
`, conflict.Version1, conflict.Version2)

	// Create files
	if err = os.WriteFile(upPath, []byte(upSQL), 0666); err != nil {
		return "", "", fmt.Errorf("failed to create merge up migration: %w", err)
	}
	if err = os.WriteFile(downPath, []byte(downSQL), 0666); err != nil {
		_ = os.Remove(upPath)
		return "", "", fmt.Errorf("failed to create merge down migration: %w", err)
	}

	return upPath, downPath, nil
}
