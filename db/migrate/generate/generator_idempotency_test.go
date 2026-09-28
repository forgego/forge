package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/sql"
	"github.com/forgego/forge/db/migrate/state"
)

const idempotencyModelsSrc = `package tracker

import "github.com/forgego/forge/schema"

type Project struct{ schema.BaseSchema }

func (Project) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.TextField("description", schema.Optional()),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
	}
}

func (Project) Meta() schema.Meta { return schema.Meta{TableName: "projects"} }

type Task struct{ schema.BaseSchema }

func (Task) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("project_id", schema.Required()),
		schema.StringField("title", schema.Required(), schema.MaxLength(200)),
		schema.BoolField("done", schema.Default(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		// EXTRA
	}
}

func (Task) Meta() schema.Meta { return schema.Meta{TableName: "tasks"} }

func (Task) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("project_id", "Project", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
`

// TestGenerateMigrations_IdempotentAcrossRuns runs makemigrations the way the
// CLI does: every run is a new generator that reads the previous schema back
// from the migration files. Unchanged models must write no migration, and an
// added field must produce only that column. Auto timestamps, defaults and
// foreign keys used to be re-emitted on every run, and the churn dropped
// created_at's NOT NULL and DEFAULT now().
func TestGenerateMigrations_IdempotentAcrossRuns(t *testing.T) {
	for _, driver := range []core.Driver{core.DriverPostgreSQL, core.DriverSQLite} {
		t.Run(string(driver), func(t *testing.T) {
			modelsDir := t.TempDir()
			migrationsDir := t.TempDir()
			modelsPath := filepath.Join(modelsDir, "models.go")
			writeModels := func(extra string) {
				src := strings.Replace(idempotencyModelsSrc, "// EXTRA", extra, 1)
				require.NoError(t, os.WriteFile(modelsPath, []byte(src), 0o644))
			}
			generate := func(name string) []string {
				before := migrationFiles(t, migrationsDir)
				gen, err := NewMigrationGeneratorForDriver(modelsDir, migrationsDir, driver)
				require.NoError(t, err)
				require.NoError(t, gen.GenerateMigrations(name))
				var added []string
				for _, file := range migrationFiles(t, migrationsDir) {
					if !contains(before, file) {
						added = append(added, file)
					}
				}
				return added
			}

			writeModels("")
			require.Len(t, generate("initial"), 2)
			assert.Empty(t, generate("unchanged"), "unchanged models must not write a migration")

			writeModels(`schema.Int32Field("priority", schema.Default(0)),`)
			added := generate("add_priority")
			require.Equal(t, []string{"000002_add_priority.down.sql", "000002_add_priority.up.sql"}, added)
			up := readStatements(t, filepath.Join(migrationsDir, added[1]))
			down := readStatements(t, filepath.Join(migrationsDir, added[0]))
			require.Len(t, up, 1, "up: %v", up)
			assert.Contains(t, up[0], `ADD COLUMN "priority" INTEGER DEFAULT 0`)
			require.Len(t, down, 1, "down: %v", down)
			assert.Contains(t, down[0], "DROP COLUMN")

			assert.Empty(t, generate("unchanged_again"), "unchanged models must not write a migration")
		})
	}
}

// silentBuilder renders no SQL for any change set.
type silentBuilder struct{ sql.SQLBuilder }

func (silentBuilder) BuildUpSQL([]core.Change) (string, error)   { return "", nil }
func (silentBuilder) BuildDownSQL([]core.Change) (string, error) { return "\n", nil }

// TestGenerateMigrations_EmptySQLWritesNothing covers change sets that render
// no SQL, which used to write an empty migration pair on every run.
func TestGenerateMigrations_EmptySQLWritesNothing(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(idempotencyModelsSrc), 0o644))
	builder, err := sql.NewSQLBuilder(core.DriverPostgreSQL)
	require.NoError(t, err)
	gen, err := NewMigrationGeneratorWithDriver(modelsDir, migrationsDir, core.DriverPostgreSQL,
		NewDetectorForDriver(core.DriverPostgreSQL), silentBuilder{builder}, state.NewFileStateLoader(migrationsDir))
	require.NoError(t, err)
	require.NoError(t, gen.GenerateMigrations("empty"))
	entries, err := os.ReadDir(migrationsDir)
	if !os.IsNotExist(err) {
		require.NoError(t, err)
		assert.Empty(t, entries, "a change set without SQL must not write a migration")
	}
}

// TestGenerateMigrations_UnrenderableColumnChangeFails keeps the empty-SQL
// guard from hiding a real change: a column change the PostgreSQL builder
// cannot express fails instead of writing nothing.
func TestGenerateMigrations_UnrenderableColumnChangeFails(t *testing.T) {
	modelsDir := t.TempDir()
	migrationsDir := t.TempDir()
	modelsPath := filepath.Join(modelsDir, "models.go")
	require.NoError(t, os.WriteFile(modelsPath, []byte(idempotencyModelsSrc), 0o644))
	gen, err := NewMigrationGeneratorForDriver(modelsDir, migrationsDir, core.DriverPostgreSQL)
	require.NoError(t, err)
	require.NoError(t, gen.GenerateMigrations("initial"))

	unique := strings.Replace(idempotencyModelsSrc, `schema.StringField("title", schema.Required(), schema.MaxLength(200))`,
		`schema.StringField("title", schema.Required(), schema.MaxLength(200), schema.Unique())`, 1)
	require.NotEqual(t, idempotencyModelsSrc, unique)
	require.NoError(t, os.WriteFile(modelsPath, []byte(unique), 0o644))
	gen, err = NewMigrationGeneratorForDriver(modelsDir, migrationsDir, core.DriverPostgreSQL)
	require.NoError(t, err)
	err = gen.GenerateMigrations("unique_title")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tasks.title`)
	assert.Contains(t, err.Error(), "UNIQUE")
	assert.Len(t, migrationFiles(t, migrationsDir), 2, "a failed generation must not write a migration")
}

func migrationFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// readStatements returns the SQL statements in a migration file, without
// comment lines.
func readStatements(t *testing.T, path string) []string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var lines []string
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "--") {
			lines = append(lines, line)
		}
	}
	var statements []string
	for _, stmt := range strings.Split(strings.Join(lines, "\n"), ";") {
		if stmt = strings.TrimSpace(stmt); stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}
