package state

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// sortMigrationFiles sorts migration files by version number
func sortMigrationFiles(files []string) []string {
	// Parse version from filename and sort numerically
	// Filename format: VERSION_NAME.up.sql (e.g., 000001_create_users.up.sql)
	sort.Slice(files, func(i, j int) bool {
		vi := extractVersionFromFilename(files[i])
		vj := extractVersionFromFilename(files[j])
		return vi < vj
	})
	return files
}

// extractVersionFromFilename extracts the version number from a migration filename
// Returns 0 if version cannot be parsed
func extractVersionFromFilename(filename string) uint64 {
	basename := filepath.Base(filename)
	// Extract version prefix (e.g., "000001" from "000001_create_users.up.sql")
	parts := strings.Split(basename, "_")
	if len(parts) == 0 {
		return 0
	}

	versionStr := parts[0]
	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return 0
	}
	return version
}

// sortChangesByType sorts changes to ensure CREATE TABLE comes before other changes
func sortChangesByType(changes []core.Change) []core.Change {
	// Create separate slices for different change types
	var createTables []core.Change
	var addColumns []core.Change
	var addForeignKeys []core.Change
	var otherChanges []core.Change

	for _, change := range changes {
		switch change.Type() {
		case core.ChangeTypeCreateTable:
			createTables = append(createTables, change)
		case core.ChangeTypeAddColumn:
			addColumns = append(addColumns, change)
		case core.ChangeTypeAddForeignKey:
			addForeignKeys = append(addForeignKeys, change)
		default:
			otherChanges = append(otherChanges, change)
		}
	}

	// Combine in order: CREATE TABLE, ADD COLUMN, ADD FOREIGN KEY, others
	result := make([]core.Change, 0, len(changes))
	result = append(result, createTables...)
	result = append(result, addColumns...)
	result = append(result, addForeignKeys...)
	result = append(result, otherChanges...)

	return result
}
