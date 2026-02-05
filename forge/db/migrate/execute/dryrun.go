package execute

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DryRunExecutor executes migrations in dry-run mode (preview only)
type DryRunExecutor struct {
	migrationsPath string
}

// NewDryRunExecutor creates a new dry-run executor
func NewDryRunExecutor(migrationsPath string) *DryRunExecutor {
	return &DryRunExecutor{
		migrationsPath: migrationsPath,
	}
}

// PreviewMigrations returns the SQL that would be executed without actually running it
func (e *DryRunExecutor) PreviewMigrations(ctx context.Context) ([]MigrationPreview, error) {
	var previews []MigrationPreview

	// Find all pending migration files
	pattern := filepath.Join(e.migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Sort by version
	sortedFiles := sortMigrationFilesByVersion(matches)

	// Read each pending migration
	for _, file := range sortedFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		basename := filepath.Base(file)
		version, name := extractVersionAndName(basename)

		preview := MigrationPreview{
			Version: version,
			Name:    name,
			SQL:     string(content),
		}

		previews = append(previews, preview)
	}

	return previews, nil
}

// MigrationPreview represents a migration that would be applied
type MigrationPreview struct {
	Version string
	Name    string
	SQL     string
}

// sortMigrationFilesByVersion sorts migration files by version number
func sortMigrationFilesByVersion(files []string) []string {
	// Parse versions and sort
	type fileWithVersion struct {
		path    string
		version uint
	}

	var filesWithVersions []fileWithVersion
	for _, file := range files {
		basename := filepath.Base(file)
		version, name := extractVersionAndName(basename)
		if version != "" {
			v, err := strconv.ParseUint(version, 10, 64)
			if err == nil {
				filesWithVersions = append(filesWithVersions, fileWithVersion{
					path:    file,
					version: uint(v),
				})
			}
		}
		_ = name // Suppress unused variable warning
	}

	// Sort by version
	for i := 0; i < len(filesWithVersions)-1; i++ {
		for j := i + 1; j < len(filesWithVersions); j++ {
			if filesWithVersions[i].version > filesWithVersions[j].version {
				filesWithVersions[i], filesWithVersions[j] = filesWithVersions[j], filesWithVersions[i]
			}
		}
	}

	// Extract sorted paths
	result := make([]string, len(filesWithVersions))
	for i, f := range filesWithVersions {
		result[i] = f.path
	}

	return result
}

// extractVersionAndName extracts version and name from migration filename
func extractVersionAndName(filename string) (version, name string) {
	// Format: "000001_create_users.up.sql"
	basename := strings.TrimSuffix(filename, ".up.sql")
	basename = strings.TrimSuffix(basename, ".down.sql")

	parts := strings.SplitN(basename, "_", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", basename
}
