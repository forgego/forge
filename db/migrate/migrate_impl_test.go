package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db/migrate/core"
	"github.com/stretchr/testify/require"
)

func writeTestModel(t *testing.T, dir string) {
	t.Helper()
	modelSrc := `package testmodels

import "github.com/forgego/forge/schema"

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
	return schema.Meta{TableName: "items"}
}
`
	err := os.WriteFile(filepath.Join(dir, "item.go"), []byte(modelSrc), 0644)
	require.NoError(t, err)
}

func TestGenerate_DefaultDriver(t *testing.T) {
	modelsDir1 := t.TempDir()
	migrationsDir1 := t.TempDir()
	writeTestModel(t, modelsDir1)

	modelsDir2 := t.TempDir()
	migrationsDir2 := t.TempDir()
	writeTestModel(t, modelsDir2)

	err1 := Generate("create_items", modelsDir1, migrationsDir1)
	require.NoError(t, err1)

	defaultDriver := core.Driver(config.NewConfig().GetDriver())
	err2 := GenerateForDriver("create_items", modelsDir2, migrationsDir2, defaultDriver)
	require.NoError(t, err2)

	up1, err := os.ReadFile(filepath.Join(migrationsDir1, "000001_create_items.up.sql"))
	require.NoError(t, err)
	up2, err := os.ReadFile(filepath.Join(migrationsDir2, "000001_create_items.up.sql"))
	require.NoError(t, err)
	require.Equal(t, string(up2), string(up1))
}

func TestNewGenerator_DefaultDriver(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()

	gen, err := NewGenerator(modelsDir, migrationsDir)
	require.NoError(t, err)
	require.NotNil(t, gen)

	defaultDriver := core.Driver(config.NewConfig().GetDriver())
	driverGen, err := NewGeneratorForDriver(modelsDir, migrationsDir, defaultDriver)
	require.NoError(t, err)
	require.NotNil(t, driverGen)
}
