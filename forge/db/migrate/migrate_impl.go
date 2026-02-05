// This file implements functions that require importing sub-packages.
// It's separated from migrate.go to break import cycles.
package migrate

import (
	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/generate"
	"github.com/forgego/forge/db/migrate/state"
	"github.com/forgego/forge/db/migrate/sql"
)

// Generate creates a new migration generator and generates migration files
// from model definitions. This is the main entry point for generating migrations.
//
// Example:
//
//	err := migrate.Generate("add_user_table", "./models", "./migrations")
func Generate(name, modelsDir, migrationsDir string) error {
	gen, err := generate.NewMigrationGeneratorWithDefaults(modelsDir, migrationsDir)
	if err != nil {
		return err
	}
	return gen.GenerateMigrations(name)
}

// LoadState loads the current database schema state from existing migration files.
// This is useful for comparing current models against the existing state.
//
// Example:
//
//	state, err := migrate.LoadState("./migrations")
//	if err != nil {
//		return err
//	}
//	previousDefs := state.ToModelDefinitions()
func LoadState(migrationsDir string) (*State, error) {
	schemaState, err := state.LoadStateFromFiles(migrationsDir)
	if err != nil {
		return nil, err
	}
	return &State{
		schemaState: schemaState,
	}, nil
}

// State wraps SchemaState and provides a simplified API
type State struct {
	schemaState *state.SchemaState
}

// ToModelDefinitions converts the state back to ModelDefinitions for comparison
func (s *State) ToModelDefinitions() []*generator.ModelDefinition {
	return s.schemaState.ToModelDefinitions()
}

// ApplyMigration applies a migration's changes to the state
func (s *State) ApplyMigration(changes []Change) error {
	manager := state.NewInMemoryState()
	managerState := manager.GetState()
	*managerState = *s.schemaState
	return manager.Apply(changes)
}

// DetectChanges compares current models to previous state and returns all changes.
// This is useful for previewing what migrations would be generated.
//
// Example:
//
//	changes, err := migrate.DetectChanges(currentDefs, previousDefs)
func DetectChanges(current, previous []*generator.ModelDefinition) ([]Change, error) {
	detector := generate.NewDetector()
	return detector.DetectChanges(current, previous)
}

// NewDetector creates a new change detector
func NewDetector() ChangeDetector {
	return generate.NewDetector()
}

// NewSQLGenerator creates a new SQL generator for the given driver.
// The driver should be "postgres" or "sqlite"/"sqlite3".
//
// Example:
//
//	sqlGen, err := migrate.NewSQLGenerator("postgres")
//	if err != nil {
//		return err
//	}
//	upSQL, err := sqlGen.GenerateUpSQL(changes)
func NewSQLGenerator(driver string) (*SQLGenerator, error) {
	driverType := Driver(driver)
	builder, err := sql.NewSQLBuilder(driverType)
	if err != nil {
		return nil, err
	}
	return &SQLGenerator{
		builder: builder,
		driver:  driver,
	}, nil
}

// SQLGenerator generates SQL DDL statements from changes
type SQLGenerator struct {
	builder sql.SQLBuilder
	driver  string
}

// GenerateUpSQL generates the up migration SQL for a list of changes
func (g *SQLGenerator) GenerateUpSQL(changes []Change) (string, error) {
	return g.builder.BuildUpSQL(changes)
}

// GenerateDownSQL generates the down migration SQL for a list of changes
func (g *SQLGenerator) GenerateDownSQL(changes []Change) (string, error) {
	return g.builder.BuildDownSQL(changes)
}

// NewGenerator creates a new migration generator with default dependencies.
// This is a convenience function that wraps NewMigrationGeneratorWithDefaults.
//
// Example:
//
//	gen, err := migrate.NewGenerator("./models", "./migrations")
//	if err != nil {
//		return err
//	}
//	err = gen.GenerateMigrations("add_user_table")
func NewGenerator(modelsDir, migrationsDir string) (*Generator, error) {
	gen, err := generate.NewMigrationGeneratorWithDefaults(modelsDir, migrationsDir)
	if err != nil {
		return nil, err
	}
	return &Generator{
		generator: gen,
	}, nil
}

// Generator wraps the migration generator
type Generator struct {
	generator *generate.MigrationGenerator
}

// GenerateMigrations generates migration files from model definitions
func (g *Generator) GenerateMigrations(name string) error {
	return g.generator.GenerateMigrations(name)
}
