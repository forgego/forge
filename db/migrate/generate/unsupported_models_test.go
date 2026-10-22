package generate_test

import (
	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/generate"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrationRefusesUnsupportedModelExpressionsBeforeWriting(t *testing.T) {
	for _, expression := range []string{`"prefix_"+"items"`, `tableName()`} {
		t.Run(expression, func(t *testing.T) {
			dir := t.TempDir()
			models := filepath.Join(dir, "models")
			migrations := filepath.Join(dir, "migrations")
			require.NoError(t, os.MkdirAll(models, 0700))
			source := `package models
import "github.com/forgego/forge/schema"
type Item struct{schema.BaseSchema}
func(Item)Fields()[]schema.Field{return []schema.Field{schema.Int64Field("id",schema.Primary()),schema.StringField("name")}}
func(Item)Meta()schema.Meta{return schema.Meta{TableName:TABLE_EXPR}}
func tableName()string{return "prefix_items"}
`
			modelFile := filepath.Join(models, "models.go")
			require.NoError(t, os.WriteFile(modelFile, []byte(strings.ReplaceAll(source, "TABLE_EXPR", `"prefix_items"`)), 0600))
			generator, err := generate.NewMigrationGeneratorForDriver(models, migrations, core.DriverPostgreSQL)
			require.NoError(t, err)
			require.NoError(t, generator.GenerateMigrations("initial"))
			files, err := filepath.Glob(filepath.Join(migrations, "*.sql"))
			require.NoError(t, err)
			require.Len(t, files, 2)
			before := make(map[string][]byte)
			for _, file := range files {
				data, err := os.ReadFile(file)
				require.NoError(t, err)
				before[file] = data
			}
			require.NoError(t, os.WriteFile(modelFile, []byte(strings.ReplaceAll(source, "TABLE_EXPR", expression)), 0600))
			require.Error(t, generator.GenerateMigrations("must_not_drop_table"))
			after, err := filepath.Glob(filepath.Join(migrations, "*.sql"))
			require.NoError(t, err)
			require.Len(t, after, 2)
			for file, original := range before {
				data, err := os.ReadFile(file)
				require.NoError(t, err)
				require.Equal(t, original, data)
			}
		})
	}
}
