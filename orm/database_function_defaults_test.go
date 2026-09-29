package orm_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/stores/storestest"
)

type stampedRow struct {
	schema.BaseSchema
	ID     int64     `db:"id" json:"id"`
	UID    string    `db:"uid" json:"uid"`
	SeenAt time.Time `db:"seen_at" json:"seen_at"`
	Label  string    `db:"label" json:"label"`
}

func (stampedRow) Meta() schema.Meta { return schema.Meta{TableName: "stamped_rows"} }

func (stampedRow) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.UUIDField("uid", schema.Required(), schema.Default("gen_random_uuid()")),
		schema.TimeField("seen_at", schema.Default("CURRENT_TIMESTAMP")),
		schema.StringField("label", schema.Default("new")),
	}
}

// TestDatabaseFunctionDefaultsAreLeftToTheDatabase: a Default naming a
// database function (now(), gen_random_uuid(), CURRENT_TIMESTAMP, ...) is
// what the migration writes as the column DEFAULT, so Manager.New and
// ApplyDefaults leave the field alone and Create leaves the column out: the
// database evaluates the function. Before, the ORM assigned the function name
// as text, or failed to convert it to time.Time.
func TestDatabaseFunctionDefaultsAreLeftToTheDatabase(t *testing.T) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			ctx := context.Background()
			database := storestest.Open(t, backend.Driver, backend.NewDSN(t))
			ddl := `CREATE TABLE stamped_rows (id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
				uid UUID DEFAULT gen_random_uuid(), seen_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP, label TEXT DEFAULT 'new')`
			if backend.Driver == "sqlite3" {
				ddl = `CREATE TABLE stamped_rows (id INTEGER PRIMARY KEY AUTOINCREMENT,
					uid TEXT DEFAULT (lower(hex(randomblob(16)))), seen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, label TEXT DEFAULT 'new')`
			}
			_, err := database.ExecContext(ctx, ddl)
			require.NoError(t, err)
			manager, err := orm.NewManagerWithDB[stampedRow]("stamped_rows", database)
			require.NoError(t, err)

			row, err := manager.New()
			require.NoError(t, err, "a database function default must not fail New")
			assert.Empty(t, row.UID, "gen_random_uuid() is not assigned in Go")
			assert.True(t, row.SeenAt.IsZero(), "CURRENT_TIMESTAMP is not assigned in Go")
			assert.Equal(t, "new", row.Label, "a literal default still is")
			require.NoError(t, manager.Create(ctx, row))

			stored, err := manager.Get(ctx, row.ID)
			require.NoError(t, err)
			assert.NotEqual(t, "gen_random_uuid()", stored.UID)
			assert.NotEmpty(t, stored.UID, "the database filled the column")
			if backend.Driver == "postgres" {
				assert.Regexp(t, uuidPattern, stored.UID)
			}
			assert.WithinDuration(t, time.Now(), stored.SeenAt, time.Minute)
		})
	}
}

type onlyGeneratedRow struct {
	schema.BaseSchema
	ID  int64  `db:"id" json:"id"`
	UID string `db:"uid" json:"uid"`
}

func (onlyGeneratedRow) Meta() schema.Meta { return schema.Meta{TableName: "only_generated_rows"} }

func (onlyGeneratedRow) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.UUIDField("uid", schema.Required(), schema.Default("gen_random_uuid()")),
	}
}

// TestDatabaseFunctionDefaultsInsertADefaultsRow: a model whose only columns are
// left to the database is inserted with DEFAULT VALUES, and a Required field
// with a database-function default is exempt from the required check on
// insert only: an update that writes it empty is rejected.
func TestDatabaseFunctionDefaultsInsertADefaultsRow(t *testing.T) {
	for _, backend := range storestest.Backends() {
		t.Run(backend.Name, func(t *testing.T) {
			ctx := context.Background()
			database := storestest.Open(t, backend.Driver, backend.NewDSN(t))
			ddl := `CREATE TABLE only_generated_rows (id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY, uid UUID DEFAULT gen_random_uuid())`
			if backend.Driver == "sqlite3" {
				ddl = `CREATE TABLE only_generated_rows (id INTEGER PRIMARY KEY AUTOINCREMENT, uid TEXT DEFAULT (lower(hex(randomblob(16)))))`
			}
			_, err := database.ExecContext(ctx, ddl)
			require.NoError(t, err)
			manager, err := orm.NewManagerWithDB[onlyGeneratedRow]("only_generated_rows", database)
			require.NoError(t, err)

			row := &onlyGeneratedRow{}
			require.NoError(t, manager.Create(ctx, row))
			assert.NotZero(t, row.ID)

			bulk := []*onlyGeneratedRow{{}, {}, {}}
			require.NoError(t, manager.BulkCreate(ctx, bulk))
			ids := map[int64]bool{}
			for _, created := range bulk {
				assert.NotZero(t, created.ID)
				ids[created.ID] = true
			}
			assert.Len(t, ids, 3, "each default-only row gets its own id")

			row.UID = ""
			err = manager.Update(ctx, row)
			require.Error(t, err, "an update must not write an empty required field")
			assert.Contains(t, err.Error(), "required")
		})
	}
}
