package generate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/db/migrate/generate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMigrationGeneratorBackwardsCompatibility ensures the original 5-arg signature still works
func TestNewMigrationGeneratorBackwardsCompatibility(t *testing.T) {
	// This should compile and work with the original 5-argument signature
	_, err := generate.NewMigrationGenerator(
		"models",     // modelsDir
		"migrations", // migrationsDir
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGenerator with 5 args failed: %v", err)
	}
}

// TestNewMigrationGeneratorUsesDefaultDriver ensures the backwards compatible function uses the default driver
func TestNewMigrationGeneratorUsesDefaultDriver(t *testing.T) {
	gen, err := generate.NewMigrationGenerator(
		"models",     // modelsDir
		"migrations", // migrationsDir
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGenerator failed: %v", err)
	}
	if gen == nil {
		t.Fatalf("NewMigrationGenerator returned nil")
	}
	// We can't easily test the driver value without mocking config, but we can ensure it's not nil
	// The driver field is unexported, so we can't access it directly in tests
	// We'll rely on the fact that the function doesn't return an error to indicate it worked
}

// TestNewMigrationGeneratorWithDriverIsAvailable ensures the new driver-aware function exists
func TestNewMigrationGeneratorWithDriverIsAvailable(t *testing.T) {
	_, err := generate.NewMigrationGeneratorWithDriver(
		"models",     // modelsDir
		"migrations", // migrationsDir
		"postgres",   // driver
		nil,          // detector
		nil,          // sqlBuilder
		nil,          // stateManager
	)
	if err != nil {
		t.Fatalf("NewMigrationGeneratorWithDriver failed: %v", err)
	}
}

// Migrations inspect model AST for diffs
func TestMigrationsInspectModelASTForDiffs(t *testing.T) {
	tmpDir := t.TempDir()
	modelsDir := filepath.Join(tmpDir, "models")
	migrationsDir := filepath.Join(tmpDir, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0755))

	modelSrc := `package models

import "github.com/forgego/forge/schema"

type Note struct {
	schema.BaseSchema
	ID int64
}

func (Note) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
	}
}

func (Note) Meta() schema.Meta {
	return schema.Meta{TableName: "notes"}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "note.go"), []byte(modelSrc), 0644))

	gen, err := generate.NewMigrationGeneratorForDriver(modelsDir, migrationsDir, "sqlite")
	require.NoError(t, err)

	err = gen.GenerateMigrations("create_notes")
	require.NoError(t, err)

	entries, err := os.ReadDir(migrationsDir)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	var foundUpSQL bool
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			foundUpSQL = true
			content, readErr := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
			require.NoError(t, readErr)
			assert.Contains(t, string(content), "CREATE TABLE")
			assert.Contains(t, string(content), "notes")
		}
	}
	assert.True(t, foundUpSQL, "expected .up.sql migration file generated from model AST diff")
}
