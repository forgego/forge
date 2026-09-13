package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/state"
)

// DependencyDetector detects dependencies from changes and SQL
type DependencyDetector struct {
	migrationsDir string
	stateManager  state.StateManager
}

// NewDependencyDetector creates a new dependency detector
func NewDependencyDetector(migrationsDir string, stateManager state.StateManager) *DependencyDetector {
	return &DependencyDetector{
		migrationsDir: migrationsDir,
		stateManager:  stateManager,
	}
}

// DetectDependencies detects dependencies from changes
// Returns list of dependencies that should be added to the migration
func (d *DependencyDetector) DetectDependencies(changes []core.Change, upSQL string) ([]core.Dependency, error) {
	var dependencies []core.Dependency
	seen := make(map[string]bool)

	// 1. Detect dependencies from foreign key relationships
	fkDeps := d.detectFromForeignKeys(changes)
	for _, dep := range fkDeps {
		key := dep.Version
		if dep.App != "" {
			key = dep.App + ":" + dep.Version
		}
		if !seen[key] {
			dependencies = append(dependencies, dep)
			seen[key] = true
		}
	}

	// 2. Detect dependencies from SQL table references
	sqlDeps := d.detectFromSQL(upSQL)
	for _, dep := range sqlDeps {
		key := dep.Version
		if dep.App != "" {
			key = dep.App + ":" + dep.Version
		}
		if !seen[key] {
			dependencies = append(dependencies, dep)
			seen[key] = true
		}
	}

	return dependencies, nil
}

// detectFromForeignKeys detects dependencies from foreign key relationships
func (d *DependencyDetector) detectFromForeignKeys(changes []core.Change) []core.Dependency {
	var dependencies []core.Dependency

	// Get current state to find which migrations created which tables
	schemaState := d.stateManager.GetState()
	if schemaState == nil {
		return dependencies
	}

	// Build a map of table name to migration version
	tableToVersion := d.buildTableToVersionMap()

	for _, change := range changes {
		// Check for AddForeignKey changes
		if fkChange, ok := change.(*core.AddForeignKey); ok {
			targetTable := fkChange.TargetTable
			if targetTable == "" {
				continue
			}

			// Find which migration created the target table
			if version, exists := tableToVersion[targetTable]; exists {
				dependencies = append(dependencies, core.Dependency{
					Version: version,
				})
			}
		}

		// Check for CreateTable changes that have foreign keys in the table definition
		if createChange, ok := change.(*core.CreateTable); ok {
			if createChange.Table != nil {
				// Check relations in the model definition
				for _, rel := range createChange.Table.Relations {
					if rel.Type == "ForeignKey" || rel.Type == "OneToOne" {
						// Use the To field directly to find target table
						if rel.To != "" {
							// Convert model name to table name (PascalCase to snake_case)
							// Use lowercase version for consistency
							targetTable := strings.ToLower(toSnakeCase(rel.To))
							// Try plural form if singular doesn't exist
							if _, exists := schemaState.Tables[targetTable]; !exists {
								targetTable = targetTable + "s"
							}
							if version, exists := tableToVersion[targetTable]; exists {
								dependencies = append(dependencies, core.Dependency{
									Version: version,
								})
							}
						}
					}
				}
			}
		}
	}

	return dependencies
}

// detectFromSQL detects dependencies from table references in SQL
func (d *DependencyDetector) detectFromSQL(sql string) []core.Dependency {
	var dependencies []core.Dependency

	// Build table to version map
	tableToVersion := d.buildTableToVersionMap()

	// Find all table references in SQL
	// Pattern: REFERENCES table_name, ALTER TABLE table_name, etc.
	refPattern := regexp.MustCompile(`(?i)REFERENCES\s+["']?(\w+)["']?`)
	matches := refPattern.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			tableName := match[1]
			if version, exists := tableToVersion[tableName]; exists {
				dependencies = append(dependencies, core.Dependency{
					Version: version,
				})
			}
		}
	}

	// Also check for ALTER TABLE ... ADD CONSTRAINT FOREIGN KEY
	alterPattern := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?\s+ADD\s+CONSTRAINT\s+.*?FOREIGN\s+KEY\s+.*?REFERENCES\s+["']?(\w+)["']?`)
	alterMatches := alterPattern.FindAllStringSubmatch(sql, -1)
	for _, match := range alterMatches {
		if len(match) >= 3 {
			targetTable := match[2]
			if version, exists := tableToVersion[targetTable]; exists {
				dependencies = append(dependencies, core.Dependency{
					Version: version,
				})
			}
		}
	}

	return dependencies
}

// buildTableToVersionMap builds a map of table name to migration version
func (d *DependencyDetector) buildTableToVersionMap() map[string]string {
	tableToVersion := make(map[string]string)

	// Find all migration files
	pattern := filepath.Join(d.migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return tableToVersion
	}

	// Parse each migration to find which tables it creates
	for _, file := range matches {
		basename := filepath.Base(file)
		parts := strings.Split(basename, "_")
		if len(parts) == 0 {
			continue
		}
		version := parts[0]

		// Read migration file
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		// Find CREATE TABLE statements
		createPattern := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?`)
		createMatches := createPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range createMatches {
			if len(match) >= 2 {
				tableName := match[1]
				// Only set if not already set (first migration that creates the table)
				if _, exists := tableToVersion[tableName]; !exists {
					tableToVersion[tableName] = version
				}
			}
		}
	}

	return tableToVersion
}

// ValidateDependencies validates that all dependencies exist
func (d *DependencyDetector) ValidateDependencies(dependencies []core.Dependency) error {
	// Get all migration versions
	pattern := filepath.Join(d.migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	versions := make(map[string]bool)
	for _, match := range matches {
		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) > 0 {
			versions[parts[0]] = true
		}
	}

	// Validate each dependency
	for _, dep := range dependencies {
		if !versions[dep.Version] {
			return fmt.Errorf("dependency on non-existent migration version: %s", dep.Version)
		}
	}

	return nil
}
