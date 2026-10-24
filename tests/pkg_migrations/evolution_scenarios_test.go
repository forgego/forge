package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

const scenarioModelsV1 = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}
`

const scenarioModelsRenameField = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("full_name", schema.Required(), schema.MaxLength(100)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}
`

const scenarioDestructiveV1 = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}

type Widget struct {
	schema.BaseSchema
}

func (Widget) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(50)),
	}
}

func (Widget) Meta() schema.Meta {
	return schema.Meta{TableName: "widgets"}
}
`

const scenarioDestructiveRemoveField = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}

type Widget struct {
	schema.BaseSchema
}

func (Widget) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required(), schema.MaxLength(50)),
	}
}

func (Widget) Meta() schema.Meta {
	return schema.Meta{TableName: "widgets"}
}
`

const scenarioDestructiveRemoveModel = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}
`

const scenarioDataTransformV1 = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("status", schema.Required(), schema.Default("pending")),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}
`

const scenarioDataTransformV2 = `package models

import "github.com/forgego/forge/schema"

type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(100)),
		schema.StringField("status", schema.Required(), schema.Default("pending")),
		schema.StringField("tier", schema.Required(), schema.Default("standard")),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers"}
}
`

// TestUnchangedModelsProduceNoChanges verifies that after an initial migration,
// running detection 100 times on unchanged models produces no new migration files.
func TestUnchangedModelsProduceNoChanges(t *testing.T) {
	work := t.TempDir()
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	writeModels := func(source string) {
		require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o644))
	}

	writeModels(scenarioModelsV1)
	out, err := forgeRun(t, "makemigrations", "initial", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, filepath.Join(migrationsDir, "000001_initial.up.sql"))

	initialFiles := listMigrationFiles(t, migrationsDir)
	require.Equal(t, []string{"000001_initial.down.sql", "000001_initial.up.sql"}, initialFiles)

	for i := 0; i < 100; i++ {
		noopOut, noopErr := forgeRun(t, "makemigrations", fmt.Sprintf("noop_%03d", i), "--auto", "--models", modelsDir, "--path", migrationsDir)
		require.NoError(t, noopErr, noopOut)
		require.Contains(t, noopOut, "No changes detected")
		require.Equal(t, initialFiles, listMigrationFiles(t, migrationsDir), "run %d produced new migration files", i)
	}
}

// TestProbableRenameWarnsAboutDataLoss pins current behavior: a rename is generated
// as ADD COLUMN plus DROP COLUMN with a data-risk warning. Generation does not yet
// ask for an explicit rename decision (spec T035 gap, tracked as T035a).
func TestProbableRenameWarnsAboutDataLoss(t *testing.T) {
	work := t.TempDir()
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	writeModels := func(source string) {
		require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o644))
	}

	// 1. Initial migration with field "name".
	writeModels(scenarioModelsV1)
	out, err := forgeRun(t, "makemigrations", "initial", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, filepath.Join(migrationsDir, "000001_initial.up.sql"))

	// 2. Rename field "name" to "full_name".
	writeModels(scenarioModelsRenameField)
	out, err = forgeRun(t, "makemigrations", "rename_field", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)

	// CLI prints a data-risk warning for the destructive DropColumn operation.
	require.Contains(t, out, "Warning:")
	require.Contains(t, out, "contains destructive operation DropColumn customers.name")

	// Generated migration contains both ADD COLUMN and DROP COLUMN.
	upSQL, err := os.ReadFile(filepath.Join(migrationsDir, "000002_rename_field.up.sql"))
	require.NoError(t, err)
	require.Contains(t, string(upSQL), "full_name")
	require.Contains(t, string(upSQL), "DROP COLUMN")
}

// TestDestructiveChangeWarns pins current behavior: dropping a field or model
// prints a destructive-operation warning, and `migrate up` refuses to apply it
// without --allow-destructive.
func TestDestructiveChangeWarns(t *testing.T) {
	setupScenarioDB(t, "destructive")
	work := t.TempDir()
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	writeModels := func(source string) {
		require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o644))
	}

	// 1. Initial models with Customer and Widget.
	writeModels(scenarioDestructiveV1)
	out, err := forgeRun(t, "makemigrations", "initial", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, filepath.Join(migrationsDir, "000001_initial.up.sql"))
	out, err = forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.NoError(t, err, out)

	// 2. Remove field "email" from Customer: makemigrations warns about DropColumn.
	writeModels(scenarioDestructiveRemoveField)
	out, err = forgeRun(t, "makemigrations", "remove_field", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "Warning:")
	require.Contains(t, out, "contains destructive operation DropColumn customers.email")

	// 3. Remove model Widget: makemigrations warns about DropTable.
	writeModels(scenarioDestructiveRemoveModel)
	out, err = forgeRun(t, "makemigrations", "remove_model", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "Warning:")
	require.Contains(t, out, "contains destructive operation DropTable widgets")

	// 4. Applying the destructive migrations needs explicit intent.
	out, err = forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.Error(t, err, out)
	require.Contains(t, err.Error()+out, "--allow-destructive")
	out, err = forgeRun(t, "migrate", "up", "--allow-destructive", "--path", migrationsDir)
	require.NoError(t, err, out)
}

// TestDataTransformationBetweenSchemaSteps verifies that an ordered data migration
// runs between two schema migrations, and its effect is visible after migrate up.
func TestDataTransformationBetweenSchemaSteps(t *testing.T) {
	conn := setupScenarioDB(t, "transform")

	work := t.TempDir()
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))

	writeModels := func(source string) {
		require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o644))
	}

	// 1. First schema migration: Customer model with status defaulting to 'pending'.
	writeModels(scenarioDataTransformV1)
	out, err := forgeRun(t, "makemigrations", "initial", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, filepath.Join(migrationsDir, "000001_initial.up.sql"))

	// Apply initial schema so we can insert rows.
	out, err = forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.NoError(t, err, out)

	_, err = conn.Exec(`INSERT INTO customers (name, status) VALUES ('Alice', 'pending'), ('Bob', 'pending')`)
	require.NoError(t, err)

	// 2. Ordered data migration: empty migration template populated with data transformation.
	out, err = forgeRun(t, "makemigrations", "backfill_status", "--empty", "--path", migrationsDir)
	require.NoError(t, err, out)

	dataUp := filepath.Join(migrationsDir, "000002_backfill_status.up.sql")
	dataDown := filepath.Join(migrationsDir, "000002_backfill_status.down.sql")
	require.NoError(t, os.WriteFile(dataUp, []byte("UPDATE customers SET status = 'active' WHERE status = 'pending';\n"), 0o644))
	require.NoError(t, os.WriteFile(dataDown, []byte("UPDATE customers SET status = 'pending' WHERE status = 'active';\n"), 0o644))

	// 3. Second schema migration: add column "tier" with default 'standard'.
	writeModels(scenarioDataTransformV2)
	out, err = forgeRun(t, "makemigrations", "add_tier", "--auto", "--models", modelsDir, "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, filepath.Join(migrationsDir, "000003_add_tier.up.sql"))

	// 4. Apply migrations (runs data migration 000002 then schema migration 000003).
	out, err = forgeRun(t, "migrate", "up", "--path", migrationsDir)
	require.NoError(t, err, out)

	// 5. Verify that the effect of the data transformation is visible.
	rows, err := conn.Query(`SELECT name, status, tier FROM customers ORDER BY name`)
	require.NoError(t, err)
	defer rows.Close()

	type customerRow struct {
		name   string
		status string
		tier   string
	}
	var results []customerRow
	for rows.Next() {
		var r customerRow
		require.NoError(t, rows.Scan(&r.name, &r.status, &r.tier))
		results = append(results, r)
	}
	require.NoError(t, rows.Err())

	require.Equal(t, []customerRow{
		{name: "Alice", status: "active", tier: "standard"},
		{name: "Bob", status: "active", tier: "standard"},
	}, results)
}

// setupScenarioDB creates a fresh Postgres database for one scenario and points
// the forge CLI at it through FORGE_DATABASE_* variables.
func setupScenarioDB(t *testing.T, prefix string) *sql.DB {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	opts := testhelpers.LocalPostgresOpts(t.Name())
	opts.DBName = fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	conn, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	if err != nil {
		testhelpers.SkipOrFailNoDB(t, "Postgres unavailable: %v", err)
	}
	t.Cleanup(func() { require.NoError(t, cleanup()) })

	u, err := url.Parse(dsn)
	require.NoError(t, err)
	password, _ := u.User.Password()
	for key, value := range map[string]string{
		"DRIVER": "postgres", "HOST": u.Hostname(), "PORT": u.Port(), "USER": u.User.Username(),
		"PASSWORD": password, "NAME": strings.TrimPrefix(u.Path, "/"), "SSLMODE": "disable",
	} {
		t.Setenv("FORGE_DATABASE_"+key, value)
	}
	return conn
}
