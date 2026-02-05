package generate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/config"
	codegen "github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/sql"
	"github.com/forgego/forge/db/migrate/state"
)

// MigrationGenerator generates migrations from model definitions
type MigrationGenerator struct {
	modelsDir     string
	migrationsDir string
	driver        core.Driver
	detector      ChangeDetector
	sqlBuilder    sql.SQLBuilder
	stateManager  state.StateManager
}

// NewMigrationGenerator creates a new migration generator with dependency injection
func NewMigrationGenerator(
	modelsDir, migrationsDir string,
	detector ChangeDetector,
	sqlBuilder sql.SQLBuilder,
	stateManager state.StateManager,
) (*MigrationGenerator, error) {
	cfg := config.NewConfig()
	driverName := cfg.GetDriver()
	driver := core.Driver(driverName)

	return &MigrationGenerator{
		modelsDir:     modelsDir,
		migrationsDir: migrationsDir,
		driver:        driver,
		detector:      detector,
		sqlBuilder:    sqlBuilder,
		stateManager:  stateManager,
	}, nil
}

// NewMigrationGeneratorWithDefaults creates a new migration generator with default dependencies
func NewMigrationGeneratorWithDefaults(modelsDir, migrationsDir string) (*MigrationGenerator, error) {
	cfg := config.NewConfig()
	driverName := cfg.GetDriver()
	driver := core.Driver(driverName)

	// Create dependencies
	detector := NewDetector()
	sqlBuilder, err := sql.NewSQLBuilder(driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL builder: %w", err)
	}
	// Use FileStateLoader to load state from existing migration files
	// This ensures incremental migrations work correctly
	stateManager := state.NewFileStateLoader(migrationsDir)

	return NewMigrationGenerator(modelsDir, migrationsDir, detector, sqlBuilder, stateManager)
}

// GenerateMigrations generates migration files from model definitions
func (g *MigrationGenerator) GenerateMigrations(name string) error {
	// Parse current models
	parser := codegen.NewASTParser()
	currentDefs, err := parser.ParseDirectory(g.modelsDir)
	if err != nil {
		return core.NewMigrationError(
			core.ErrParseFailed,
			"failed to parse models",
			err,
		)
	}

	if len(currentDefs) == 0 {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			fmt.Sprintf("no model definitions found in %s", g.modelsDir),
			nil,
		)
	}

	// Load previous state from state manager
	// If stateManager is a FileStateLoader, it will load from migration files
	// Otherwise, it will use in-memory state
	schemaState, err := g.stateManager.Load()
	if err != nil {
		return core.NewMigrationError(
			core.ErrStateMismatch,
			"failed to load state",
			err,
		)
	}

	// If state is nil or empty, create a new one
	if schemaState == nil {
		schemaState = &state.SchemaState{
			Tables: make(map[string]*state.TableState),
		}
	}

	// Convert state to model definitions for comparison
	previousDefs := schemaState.ToModelDefinitions()

	// Detect changes
	changes, err := g.detector.DetectChanges(currentDefs, previousDefs)
	if err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to detect changes",
			err,
		)
	}

	// If no changes, return early (this is not an error, just no migration needed)
	if len(changes) == 0 {
		return nil // Return nil to indicate no migration needed
	}

	// Validate changes before generating SQL
	if err := validateChanges(changes, currentDefs); err != nil {
		return core.NewMigrationError(
			core.ErrValidationFailed,
			"validation failed",
			err,
		)
	}

	// Check if this is the first migration
	isFirstMigration := false
	if entries, err := os.ReadDir(g.migrationsDir); err == nil {
		sqlFileCount := 0
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && (strings.HasSuffix(name, ".up.sql") || strings.HasSuffix(name, ".down.sql")) {
				sqlFileCount++
			}
		}
		isFirstMigration = sqlFileCount == 0
	}

	// Generate SQL
	upSQL, err := g.sqlBuilder.BuildUpSQL(changes)
	if err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to generate up SQL",
			err,
		)
	}

	downSQL, err := g.sqlBuilder.BuildDownSQL(changes)
	if err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to generate down SQL",
			err,
		)
	}

	// Prepend bookkeeping table to first migration
	if isFirstMigration {
		bookkeepingSQL := generateBookkeepingTable(g.driver)
		upSQL = bookkeepingSQL + "\n\n" + upSQL
		downSQL = downSQL + "\n\n" + "DROP TABLE IF EXISTS schema_migrations;"
	}

	// Get next version
	version, err := getNextVersion(g.migrationsDir)
	if err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to get next version",
			err,
		)
	}

	// Checksum will be calculated and validated when migration is applied
	// See execute/checksum.go for checksum validation

	// Auto-detect dependencies from foreign keys and SQL references
	dependencyDetector := NewDependencyDetector(g.migrationsDir, g.stateManager)
	dependencies, err := dependencyDetector.DetectDependencies(changes, upSQL)
	if err != nil {
		// Log error but continue - dependencies are optional
		// In production, you might want to make this stricter
	}

	// Validate dependencies exist
	if err := dependencyDetector.ValidateDependencies(dependencies); err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"dependency validation failed",
			err,
		)
	}

	// Add dependency comments to SQL if dependencies were detected
	if len(dependencies) > 0 {
		var depComments []string
		depComments = append(depComments, "-- Auto-detected dependencies:")
		for _, dep := range dependencies {
			if dep.App != "" {
				depComments = append(depComments, fmt.Sprintf("-- DEPENDS: %s:%s", dep.App, dep.Version))
			} else {
				depComments = append(depComments, fmt.Sprintf("-- DEPENDS: %s", dep.Version))
			}
		}
		upSQL = strings.Join(depComments, "\n") + "\n\n" + upSQL
	}

	// Create migration files
	migrationName := fmt.Sprintf("%s_%s", version, name)
	upPath := filepath.Join(g.migrationsDir, fmt.Sprintf("%s.up.sql", migrationName))
	downPath := filepath.Join(g.migrationsDir, fmt.Sprintf("%s.down.sql", migrationName))

	// Ensure directory exists
	if err := os.MkdirAll(g.migrationsDir, 0755); err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to create migrations directory",
			err,
		)
	}

	// Write up migration
	if err := os.WriteFile(upPath, []byte(upSQL+"\n"), 0644); err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to write up migration",
			err,
		)
	}

	// Write down migration
	if err := os.WriteFile(downPath, []byte(downSQL+"\n"), 0644); err != nil {
		return core.NewMigrationError(
			core.ErrInvalidChange,
			"failed to write down migration",
			err,
		)
	}

	// Update state
	if err := g.stateManager.Apply(changes); err != nil {
		return core.NewMigrationError(
			core.ErrStateMismatch,
			"failed to update state",
			err,
		)
	}

	// Checksum is calculated and can be stored in schema_migrations table when migration is applied
	// The checksum calculation is available via calculateChecksum function

	return nil
}

// validateChanges validates that all changes are valid before SQL generation
func validateChanges(changes []core.Change, defs []*codegen.ModelDefinition) error {
	// Build a map of table names to their definitions for quick lookup
	tableDefs := make(map[string]*codegen.ModelDefinition)
	for _, def := range defs {
		tableName := getTableNameFromDef(def)
		tableDefs[tableName] = def
	}

	for _, change := range changes {
		switch c := change.(type) {
		case *core.AddForeignKey:
			// Validate target table exists
			if c.TargetTable == "" {
				return fmt.Errorf("foreign key %s.%s has empty target table", c.Table, c.Relation.Name)
			}
			// Validate source table exists
			if _, exists := tableDefs[c.Table]; !exists {
				return fmt.Errorf("foreign key references non-existent table: %s", c.Table)
			}
			// Validate target table exists in definitions
			targetExists := false
			for _, def := range defs {
				if getTableNameFromDef(def) == c.TargetTable {
					targetExists = true
					break
				}
			}
			if !targetExists {
				return fmt.Errorf("foreign key %s.%s references non-existent target table: %s", c.Table, c.Relation.Name, c.TargetTable)
			}
		}
	}
	return nil
}

// getNextVersion gets the next migration version number
func getNextVersion(migrationsDir string) (string, error) {
	const seqDigits = 6

	// Find all migration files
	pattern := filepath.Join(migrationsDir, "*_*.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	nextSeq := uint64(1)

	if len(matches) > 0 {
		versions := make([]uint64, 0, len(matches))
		for _, filename := range matches {
			basename := filepath.Base(filename)
			idx := -1
			for i, r := range basename {
				if r == '_' {
					idx = i
					break
				}
			}
			if idx < 1 {
				continue
			}

			versionStr := basename[0:idx]
			version, err := parseUint(versionStr)
			if err != nil {
				continue
			}
			versions = append(versions, version)
		}

		if len(versions) > 0 {
			// Find maximum
			maxVersion := versions[0]
			for _, v := range versions[1:] {
				if v > maxVersion {
					maxVersion = v
				}
			}
			nextSeq = maxVersion + 1
		}
	}

	// Format with zero-padding
	version := fmt.Sprintf("%0*d", seqDigits, nextSeq)
	return version, nil
}

// generateBookkeepingTable generates SQL for schema_migrations table
func generateBookkeepingTable(driver core.Driver) string {
	if driver.IsSQLite() {
		return `-- Create schema_migrations table for tracking applied migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    checksum TEXT,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
	}

	return `-- Create schema_migrations table for tracking applied migrations
CREATE TABLE IF NOT EXISTS schema_migrations (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    checksum TEXT,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);`
}

// calculateChecksum calculates SHA256 checksum of SQL
func calculateChecksum(sql string) string {
	hash := sha256.Sum256([]byte(sql))
	return hex.EncodeToString(hash[:])
}

// parseUint helper
func parseUint(s string) (uint64, error) {
	var result uint64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		result = result*10 + uint64(r-'0')
	}
	return result, nil
}

// getTableNameFromDef gets the table name from a model definition (helper)
func getTableNameFromDef(def *codegen.ModelDefinition) string {
	if def.Meta.TableName != "" {
		return def.Meta.TableName
	}
	return fmt.Sprintf("%ss", toSnakeCaseFromDef(def.Name))
}

// toSnakeCaseFromDef converts CamelCase to snake_case (helper)
func toSnakeCaseFromDef(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
}
