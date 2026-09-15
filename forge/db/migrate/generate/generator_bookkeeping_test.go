package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateMigrations_FirstMigrationOmitsBookkeepingTable(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()

	modelSrc := `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Item struct {
	schema.BaseSchema
}

func (Item) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
	}
}

func (Item) Meta() schema.Meta {
	return schema.Meta{
		TableName: "items",
	}
}
`
	modelFile := filepath.Join(modelsDir, "item.go")
	require.NoError(t, os.WriteFile(modelFile, []byte(modelSrc), 0644))

	gen, err := NewMigrationGeneratorForDriver(modelsDir, migrationsDir, core.DriverSQLite)
	require.NoError(t, err)

	err = gen.GenerateMigrations("create_items")
	require.NoError(t, err)

	upBytes, err := os.ReadFile(filepath.Join(migrationsDir, "000001_create_items.up.sql"))
	require.NoError(t, err)
	assert.NotContains(t, string(upBytes), "schema_migrations")

	downBytes, err := os.ReadFile(filepath.Join(migrationsDir, "000001_create_items.down.sql"))
	require.NoError(t, err)
	assert.NotContains(t, string(downBytes), "schema_migrations")
}
