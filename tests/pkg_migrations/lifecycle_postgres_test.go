package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	migrationcmds "github.com/forgego/forge/cli/commands/migrations"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/tests/testhelpers"
)

// Models for the lifecycle test, written in the functional DSL that compiles.
const lifecycleModelsV1 = `package models

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

type Order struct {
	schema.BaseSchema
}

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required()),
		schema.Float64Field("total", schema.Default(0)),
		schema.TextField("note", schema.Optional()),
	}
}

func (Order) Meta() schema.Meta {
	return schema.Meta{TableName: "orders"}
}

func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
`

// lifecycleModelsV2 adds a column with a default (customers.status), a new
// table with a relation to it (orders.coupon_id -> coupons), a CHECK constraint
// on orders, and a UNIQUE constraint on customers.
const lifecycleModelsV2 = `package models

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
		schema.StringField("status", schema.Required(), schema.Default("active")),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{
		TableName: "customers",
		Constraints: []schema.Constraint{
			{Name: "customers_name_email_key", Type: "UNIQUE", Fields: []string{"name", "email"}},
		},
	}
}

type Coupon struct {
	schema.BaseSchema
}

func (Coupon) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("code", schema.Required(), schema.Unique()),
	}
}

func (Coupon) Meta() schema.Meta {
	return schema.Meta{TableName: "coupons"}
}

type Order struct {
	schema.BaseSchema
}

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required()),
		schema.Float64Field("total", schema.Default(0)),
		schema.TextField("note", schema.Optional()),
		schema.Int64Field("coupon_id", schema.Optional()),
	}
}

func (Order) Meta() schema.Meta {
	return schema.Meta{
		TableName: "orders",
		Constraints: []schema.Constraint{
			{Name: "orders_total_nonnegative", Type: "CHECK", Condition: "total >= 0"},
		},
	}
}

func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("coupon_id", "Coupon", schema.OnDelete(schema.CascadeSET_NULL)),
	}
}
`

// A hand-written data migration whose second statement fails after the first
// has already changed rows.
const (
	backfillFailingUp = `UPDATE customers SET status = 'vip' WHERE email LIKE '%@vip.example';
INSERT INTO orders (customer_id, total) VALUES (999999, 1);
`
	backfillFixedUp = `UPDATE customers SET status = 'vip' WHERE email LIKE '%@vip.example';
`
	backfillDown = `UPDATE customers SET status = 'active' WHERE status = 'vip';
`
)

// forgeRun runs a forge migration CLI command in-process. It wires the
// command the way core.Registry.BuildRootCommand does for cmd/forge: a fresh
// config (read from FORGE_DATABASE_* here) and the command's Execute method.
// args are what follows `forge`, for example "migrate", "up", "--path", dir.
func forgeRun(t *testing.T, args ...string) (string, error) {
	t.Helper()
	commands := map[string]core.Command{
		"makemigrations": migrationcmds.NewMakeMigrationsCommand(),
		"up":             migrationcmds.NewUpCommand(),
		"force":          migrationcmds.NewForceCommand(),
		"recover":        migrationcmds.NewRecoverCommand(),
	}
	if args[0] == "migrate" {
		args = args[1:]
	}
	command, ok := commands[args[0]]
	require.True(t, ok, "unknown command %q", args[0])
	definition := command.Definition()
	var out bytes.Buffer
	definition.SetOut(&out)
	definition.SetErr(&out)
	require.NoError(t, definition.ParseFlags(args[1:]))
	positional := definition.Flags().Args()
	if definition.Args != nil {
		require.NoError(t, definition.Args(definition, positional))
	}
	ctx := core.NewContextWithConfig(config.NewConfig())
	ctx.Cmd = definition
	err := command.Execute(ctx, positional)
	return out.String(), err
}

func listMigrationFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names
}

type migrationState struct {
	version uint
	dirty   bool
}

func readMigrationState(t *testing.T, conn *sql.DB) migrationState {
	t.Helper()
	var state migrationState
	require.NoError(t, conn.QueryRow(`SELECT version, dirty FROM schema_migrations`).Scan(&state.version, &state.dirty))
	return state
}

// assertChecksumsMatchFiles requires one checksum row per applied version,
// each equal to the file currently on disk.
func assertChecksumsMatchFiles(t *testing.T, conn *sql.DB, dir string, versions ...uint) {
	t.Helper()
	rows, err := conn.Query(`SELECT version, up_sha256, down_sha256 FROM forge_migration_checksums ORDER BY version`)
	require.NoError(t, err)
	defer rows.Close()
	var got []uint
	for rows.Next() {
		var version uint
		var up string
		var down sql.NullString
		require.NoError(t, rows.Scan(&version, &up, &down))
		wantUp, wantDown, err := db.MigrationFileChecksums(dir, version)
		require.NoError(t, err)
		require.Equal(t, wantUp, up, "up checksum of version %d", version)
		require.NotNil(t, wantDown)
		require.Equal(t, *wantDown, down.String, "down checksum of version %d", version)
		got = append(got, version)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, versions, got, "checksum rows must match the applied history")
}

type columnFacts struct {
	dataType string
	nullable string
	def      sql.NullString
}

func readColumn(t *testing.T, conn *sql.DB, table, column string) columnFacts {
	t.Helper()
	var facts columnFacts
	err := conn.QueryRow(`SELECT data_type, is_nullable, column_default FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`, table, column).
		Scan(&facts.dataType, &facts.nullable, &facts.def)
	require.NoError(t, err, "column %s.%s", table, column)
	return facts
}

func constraintType(t *testing.T, conn *sql.DB, table, name string) string {
	t.Helper()
	var kind string
	err := conn.QueryRow(`SELECT constraint_type FROM information_schema.table_constraints
		WHERE table_schema = current_schema() AND table_name = $1 AND constraint_name = $2`, table, name).Scan(&kind)
	require.NoError(t, err, "constraint %s on %s", name, table)
	return kind
}

// seededRows returns every seeded row as text, for before/after comparison.
func seededRows(t *testing.T, conn *sql.DB) []string {
	t.Helper()
	var rows []string
	for _, query := range []string{
		`SELECT 'customer', id, name, email, created_at::text FROM customers ORDER BY id`,
		`SELECT 'order', id, customer_id::text, total::text, COALESCE(note, '<null>') FROM orders ORDER BY id`,
	} {
		result, err := conn.Query(query)
		require.NoError(t, err)
		for result.Next() {
			var kind, a, b, c string
			var id int64
			require.NoError(t, result.Scan(&kind, &id, &a, &b, &c))
			rows = append(rows, fmt.Sprintf("%s %d %s %s %s", kind, id, a, b, c))
		}
		require.NoError(t, result.Err())
		result.Close()
	}
	return rows
}

// TestPostgresSchemaLifecycle drives models through makemigrations, migrate up,
// schema evolution with seeded data, a no-op regeneration, a failed migration
// with documented recovery, and checksum tamper detection, all through the
// forge CLI commands. docs-site/docs/migrations.md documents the same steps.
func TestPostgresSchemaLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	opts := testhelpers.LocalPostgresOpts(t.Name())
	opts.DBName = fmt.Sprintf("lifecycle_%d", time.Now().UnixNano())
	conn, dsn, cleanup, err := testhelpers.StartPostgresContainer(ctx, opts)
	if err != nil {
		testhelpers.SkipOrFailNoDB(t, "Postgres unavailable: %v", err)
	}
	defer func() { require.NoError(t, cleanup()) }()

	u, err := url.Parse(dsn)
	require.NoError(t, err)
	password, _ := u.User.Password()
	for key, value := range map[string]string{
		"DRIVER": "postgres", "HOST": u.Hostname(), "PORT": u.Port(), "USER": u.User.Username(),
		"PASSWORD": password, "NAME": strings.TrimPrefix(u.Path, "/"), "SSLMODE": "disable",
	} {
		t.Setenv("FORGE_DATABASE_"+key, value)
	}

	work := t.TempDir()
	modelsDir := filepath.Join(work, "models")
	migrationsDir := filepath.Join(work, "migrations")
	require.NoError(t, os.MkdirAll(modelsDir, 0o755))
	require.NoError(t, os.MkdirAll(migrationsDir, 0o755))
	writeModels := func(source string) {
		require.NoError(t, os.WriteFile(filepath.Join(modelsDir, "models.go"), []byte(source), 0o644))
	}
	makemigrations := func(name string) string {
		out, err := forgeRun(t, "makemigrations", name, "--auto", "--models", modelsDir, "--path", migrationsDir)
		require.NoError(t, err, out)
		return out
	}
	migrateUp := func() (string, error) {
		return forgeRun(t, "migrate", "up", "--path", migrationsDir)
	}

	// 1. Initial schema and representative rows.
	writeModels(lifecycleModelsV1)
	require.Contains(t, makemigrations("initial"), filepath.Join(migrationsDir, "000001_initial.up.sql"))
	require.Equal(t, []string{"000001_initial.down.sql", "000001_initial.up.sql"}, listMigrationFiles(t, migrationsDir))
	out, err := migrateUp()
	require.NoError(t, err, out)
	testhelpers.AssertForeignKeyExists(ctx, t, conn, "postgres", "orders", "customer_id")
	_, err = conn.Exec(`INSERT INTO customers (name, email) VALUES
		('Ada', 'ada@vip.example'), ('Grace', 'grace@example.com'), ('Linus', 'linus@vip.example')`)
	require.NoError(t, err)
	_, err = conn.Exec(`INSERT INTO orders (customer_id, total, note) VALUES
		(1, 10.5, 'first'), (1, 0, NULL), (2, 99.99, 'gift'), (3, 7, NULL)`)
	require.NoError(t, err)
	seeded := seededRows(t, conn)
	require.Len(t, seeded, 7)

	// 2. Evolve: column with default, new related table, foreign key, CHECK and UNIQUE.
	writeModels(lifecycleModelsV2)
	require.Contains(t, makemigrations("evolve"), filepath.Join(migrationsDir, "000002_evolve.up.sql"))
	require.Len(t, listMigrationFiles(t, migrationsDir), 4)
	evolveUp, err := os.ReadFile(filepath.Join(migrationsDir, "000002_evolve.up.sql"))
	require.NoError(t, err)
	for _, want := range []string{
		`ALTER TABLE customers ADD COLUMN "status" TEXT NOT NULL DEFAULT 'active';`,
		`ALTER TABLE orders ADD COLUMN "coupon_id" BIGINT;`,
		`CREATE TABLE IF NOT EXISTS coupons (`,
		`FOREIGN KEY ("coupon_id") REFERENCES coupons (id) ON DELETE SET NULL`,
		`ALTER TABLE orders ADD CONSTRAINT orders_total_nonnegative CHECK (total >= 0);`,
		`ALTER TABLE customers ADD CONSTRAINT customers_name_email_key UNIQUE ("name", "email");`,
	} {
		require.Contains(t, string(evolveUp), want)
	}
	for _, unwanted := range []string{"DROP", "fk_orders_customer_id", `"total" TYPE`, `"created_at"`} {
		require.NotContains(t, string(evolveUp), unwanted, "evolution must not touch unchanged schema")
	}
	out, err = migrateUp()
	require.NoError(t, err, out)

	status := readColumn(t, conn, "customers", "status")
	require.Equal(t, "text", status.dataType)
	require.Equal(t, "NO", status.nullable)
	require.Contains(t, status.def.String, "'active'")
	coupon := readColumn(t, conn, "orders", "coupon_id")
	require.Equal(t, "bigint", coupon.dataType)
	require.Equal(t, "YES", coupon.nullable)
	testhelpers.AssertForeignKeyExists(ctx, t, conn, "postgres", "orders", "coupon_id")
	testhelpers.AssertForeignKeyExists(ctx, t, conn, "postgres", "orders", "customer_id")
	require.Equal(t, "CHECK", constraintType(t, conn, "orders", "orders_total_nonnegative"))
	require.Equal(t, "UNIQUE", constraintType(t, conn, "customers", "customers_name_email_key"))
	require.Equal(t, seeded, seededRows(t, conn), "evolution must preserve seeded rows")
	var active int
	require.NoError(t, conn.QueryRow(`SELECT count(*) FROM customers WHERE status = 'active'`).Scan(&active))
	require.Equal(t, 3, active, "existing rows take the new column's default")
	_, err = conn.Exec(`INSERT INTO orders (customer_id, total) VALUES (1, -1)`)
	require.ErrorContains(t, err, "orders_total_nonnegative")
	_, err = conn.Exec(`INSERT INTO orders (customer_id, coupon_id) VALUES (1, 424242)`)
	require.ErrorContains(t, err, "fk_orders_coupon_id")

	// 3. Regenerating unchanged models writes nothing; state matches history.
	before := listMigrationFiles(t, migrationsDir)
	require.Contains(t, makemigrations("noop"), "No changes detected")
	require.Equal(t, before, listMigrationFiles(t, migrationsDir), "regenerating unchanged models must not write a migration")
	require.Equal(t, migrationState{version: 2}, readMigrationState(t, conn))
	assertChecksumsMatchFiles(t, conn, migrationsDir, 1, 2)
	out, err = forgeRun(t, "migrate", "recover", "--verify", "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "Summary: 2 verified, 0 mismatched, 0 missing, 0 unverified")
	require.Contains(t, out, "state is clean")

	// 4. A migration that fails after a statement has run leaves the version dirty.
	upFile := filepath.Join(migrationsDir, "000003_backfill_status.up.sql")
	require.NoError(t, os.WriteFile(upFile, []byte(backfillFailingUp), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(migrationsDir, "000003_backfill_status.down.sql"), []byte(backfillDown), 0o644))
	out, err = migrateUp()
	require.Error(t, err, out)
	require.ErrorContains(t, err, "fk_orders_customer_id")
	require.Equal(t, migrationState{version: 3, dirty: true}, readMigrationState(t, conn))
	var vip int
	require.NoError(t, conn.QueryRow(`SELECT count(*) FROM customers WHERE status = 'vip'`).Scan(&vip))
	require.Zero(t, vip, "PostgreSQL runs a migration file in one transaction, so the UPDATE before the failure is rolled back")
	assertChecksumsMatchFiles(t, conn, migrationsDir, 1, 2)
	out, err = migrateUp()
	require.ErrorContains(t, err, "dirty state (version 3)", out)

	out, err = forgeRun(t, "migrate", "recover", "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "dirty migration detected at version 3")
	require.Contains(t, out, "forge migrate force 2", "recover must print the rolled-back recovery path")

	// Documented recovery: the file rolled back, so fix it, force the last
	// good version, and apply again.
	require.NoError(t, os.WriteFile(upFile, []byte(backfillFixedUp), 0o644))
	out, err = forgeRun(t, "migrate", "force", "2", "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Equal(t, migrationState{version: 2}, readMigrationState(t, conn))
	out, err = migrateUp()
	require.NoError(t, err, out)
	require.Equal(t, migrationState{version: 3}, readMigrationState(t, conn))
	require.NoError(t, conn.QueryRow(`SELECT count(*) FROM customers WHERE status = 'vip'`).Scan(&vip))
	require.Equal(t, 2, vip)
	assertChecksumsMatchFiles(t, conn, migrationsDir, 1, 2, 3)
	before = listMigrationFiles(t, migrationsDir)
	require.Contains(t, makemigrations("after_recovery"), "No changes detected")
	require.Equal(t, before, listMigrationFiles(t, migrationsDir), "a data-only migration must not change generated state")

	// 5. Editing an applied migration is detected and blocks further applies.
	initialUp := filepath.Join(migrationsDir, "000001_initial.up.sql")
	original, err := os.ReadFile(initialUp)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(initialUp, append(append([]byte{}, original...), []byte("-- edited after apply\n")...), 0o644))
	out, err = forgeRun(t, "migrate", "recover", "--verify", "--path", migrationsDir)
	require.Error(t, err, out)
	require.Contains(t, out, "1  mismatched")
	out, err = migrateUp()
	require.ErrorContains(t, err, "1 (mismatched)", out)
	require.NoError(t, os.WriteFile(initialUp, original, 0o644))
	out, err = forgeRun(t, "migrate", "recover", "--verify", "--path", migrationsDir)
	require.NoError(t, err, out)
	require.Contains(t, out, "Summary: 3 verified, 0 mismatched, 0 missing, 0 unverified")
	require.Equal(t, seeded[:3], seededRows(t, conn)[:3], "customer rows survive the whole lifecycle")
}
