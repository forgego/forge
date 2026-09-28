package orm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type NoteItem struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Note string `db:"note"`
}

func (NoteItem) Meta() schema.Meta {
	return schema.Meta{TableName: "notes"}
}

func (NoteItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("note"),
	}
}

func setupNotesDB(t *testing.T) *db.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "notes_test.sqlite")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)

	_, err = database.Exec(`
		CREATE TABLE "notes" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"note" TEXT NOT NULL
		);
	`)
	require.NoError(t, err)

	_, err = GetModelSchema[NoteItem]()
	require.NoError(t, err)

	return database
}

func TestSQLite_LiteralWithPlaceholdersAndILIKE(t *testing.T) {
	database := setupNotesDB(t)
	ctx := context.Background()

	mgr, err := NewManager[NoteItem]("notes")
	require.NoError(t, err)
	mgr.SetDB(database)

	literalVal := "costs $1 ILIKE"
	item := &NoteItem{Note: literalVal}
	err = mgr.Create(ctx, item)
	require.NoError(t, err)

	// Filter with F(field).Eq and read back unchanged
	qs, err := mgr.Filter(F("note").Eq(literalVal))
	require.NoError(t, err)

	// Also add raw expression
	rawExpr := &RawExpression{
		SQL: "note = 'costs $1 ILIKE'",
	}
	qs = qs.Filter(rawExpr)

	found, err := qs.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, literalVal, found.Note)
}

func TestRawExpression_LiteralNotCorrupted(t *testing.T) {
	builder := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
	raw := &RawExpression{
		SQL:  "note = 'costs $1 ILIKE' AND id = $1",
		Args: []interface{}{123},
	}
	sql, args, err := raw.ToSQL(builder)
	require.NoError(t, err)
	assert.Equal(t, "note = 'costs $1 ILIKE' AND id = ?", sql)
	assert.Equal(t, []interface{}{123}, args)
}

func TestRawExpression_TenPlusArgsMapsCorrectly(t *testing.T) {
	builder := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
	raw := &RawExpression{
		SQL: "c1 = $1 AND c2 = $2 AND c3 = $3 AND c4 = $4 AND c5 = $5 AND c6 = $6 AND c7 = $7 AND c8 = $8 AND c9 = $9 AND c10 = $10",
		Args: []interface{}{
			"v1", "v2", "v3", "v4", "v5",
			"v6", "v7", "v8", "v9", "v10",
		},
	}
	sql, args, err := raw.ToSQL(builder)
	require.NoError(t, err)
	assert.Equal(t, "c1 = $1 AND c2 = $2 AND c3 = $3 AND c4 = $4 AND c5 = $5 AND c6 = $6 AND c7 = $7 AND c8 = $8 AND c9 = $9 AND c10 = $10", sql)
	assert.Len(t, args, 10)
	assert.Equal(t, "v10", args[9])
}

func TestPostgreSQLDialect_BuilderEmitsPositionalPlaceholders(t *testing.T) {
	builder := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
	p1 := builder.AddArg("first")
	p2 := builder.AddArg("second")
	assert.Equal(t, "$1", p1)
	assert.Equal(t, "$2", p2)
}

func TestSQLiteDialect_BuilderEmitsQuestionMarkPlaceholders(t *testing.T) {
	builder := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
	p1 := builder.AddArg("first")
	p2 := builder.AddArg("second")
	assert.Equal(t, "?", p1)
	assert.Equal(t, "?", p2)
}

func TestRawExpression_BackslashDoesNotEscapeClosingQuote(t *testing.T) {
	t.Run("SQLite builder maps placeholder after backslash literal", func(t *testing.T) {
		builder := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		raw := &RawExpression{
			SQL:  `note = 'foo\' AND id = $1`,
			Args: []interface{}{42},
		}
		sql, args, err := raw.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, `note = 'foo\' AND id = ?`, sql)
		assert.Equal(t, []interface{}{42}, args)
	})

	t.Run("PostgreSQL builder maps placeholder after backslash literal", func(t *testing.T) {
		builder := NewSQLBuilderWithDialect(dialect.NewPostgreSQLDialect())
		raw := &RawExpression{
			SQL:  `note = 'foo\' AND id = $1`,
			Args: []interface{}{42},
		}
		sql, args, err := raw.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, `note = 'foo\' AND id = $1`, sql)
		assert.Equal(t, []interface{}{42}, args)
	})

	t.Run("doubled quote literal still works", func(t *testing.T) {
		builder := NewSQLBuilderWithDialect(dialect.NewSQLiteDialect())
		raw := &RawExpression{
			SQL:  `note = 'foo''bar' AND id = $1`,
			Args: []interface{}{42},
		}
		sql, args, err := raw.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, `note = 'foo''bar' AND id = ?`, sql)
		assert.Equal(t, []interface{}{42}, args)
	})
}
