package orm

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/internal/testutils"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeValuesModel pins issue #291: every column below has a database
// default that differs from the Go zero value.
type writeValuesModel struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Active    bool      `db:"active"`
	Rank      int64     `db:"rank"`
	Label     string    `db:"label"`
	InStock   bool      `db:"in_stock"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (writeValuesModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.BoolField("active", schema.Default(true)),
		schema.Int64Field("rank", schema.Default(5)),
		schema.StringField("label", schema.Default("x")),
		schema.BoolField("in_stock", schema.DBDefault("true")),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.TimeField("updated_at", schema.AutoNow()),
	}
}

func (writeValuesModel) Meta() schema.Meta { return schema.Meta{TableName: "write_values"} }

type writeValuesBackend struct {
	database *db.DB
	table    string
	// past is a SQL literal for a timestamp in 2001.
	past string
}

// forEachWriteValuesBackend runs fn against SQLite and, when it is
// reachable, PostgreSQL (FORGE_TEST_DATABASE_URL).
func forEachWriteValuesBackend(t *testing.T, fn func(t *testing.T, backend writeValuesBackend)) {
	t.Run("sqlite", func(t *testing.T) {
		sqliteDB, err := db.NewDB(filepath.Join(t.TempDir(), "write-values.sqlite"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = sqliteDB.Close() })
		_, err = sqliteDB.Exec(`CREATE TABLE write_values (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT 1,
			rank INTEGER NOT NULL DEFAULT 5,
			label TEXT NOT NULL DEFAULT 'x',
			in_stock BOOLEAN NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`)
		require.NoError(t, err)
		fn(t, writeValuesBackend{database: sqliteDB, table: "write_values", past: "'2001-01-01 00:00:00'"})
	})

	t.Run("postgres", func(t *testing.T) {
		sqlDB := testutils.SetupTestDB(t)
		pg := &db.DB{DB: sqlDB, Driver: "postgres"}
		table := fmt.Sprintf("write_values_%d", time.Now().UnixNano())
		_, err := pg.Exec(fmt.Sprintf(`CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			rank BIGINT NOT NULL DEFAULT 5,
			label TEXT NOT NULL DEFAULT 'x',
			in_stock BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ DEFAULT now()
		)`, EscapeIdentifier(table)))
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = pg.Exec(`DROP TABLE IF EXISTS ` + EscapeIdentifier(table))
			_ = sqlDB.Close()
		})
		fn(t, writeValuesBackend{database: pg, table: table, past: "'2001-01-01 00:00:00+00'"})
	})
}

func TestManagerCreate_WritesExplicitZeroValues(t *testing.T) {
	forEachWriteValuesBackend(t, func(t *testing.T, backend writeValuesBackend) {
		ctx := context.Background()
		manager, err := NewManagerWithDB[writeValuesModel](backend.table, backend.database)
		require.NoError(t, err)

		zero := &writeValuesModel{Name: "zero"}
		require.NoError(t, manager.Create(ctx, zero))
		require.NotZero(t, zero.ID)

		var active, inStock bool
		var rank int64
		var label string
		var createdAtSet bool
		require.NoError(t, backend.database.QueryRow(
			fmt.Sprintf(`SELECT active, rank, label, in_stock, created_at IS NOT NULL FROM %s WHERE id = %d`, EscapeIdentifier(backend.table), zero.ID),
		).Scan(&active, &rank, &label, &inStock, &createdAtSet))
		assert.False(t, active, "explicit false must not become DEFAULT TRUE")
		assert.Zero(t, rank, "explicit 0 must not become DEFAULT 5")
		assert.Equal(t, "", label, "explicit empty string must not become DEFAULT 'x'")
		assert.True(t, inStock, "a zero DBDefault field is omitted, so the database default applies")
		assert.True(t, createdAtSet, "a zero AutoNowAdd field is filled by the database")

		// Manager.New applies schema defaults at construction time.
		withDefaults, err := manager.New()
		require.NoError(t, err)
		withDefaults.Name = "defaults"
		require.NoError(t, manager.Create(ctx, withDefaults))
		stored, err := manager.Get(ctx, withDefaults.ID)
		require.NoError(t, err)
		assert.True(t, stored.Active)
		assert.Equal(t, int64(5), stored.Rank)
		assert.Equal(t, "x", stored.Label)
	})
}

func TestManagerUpdate_RefreshesAutoNow(t *testing.T) {
	forEachWriteValuesBackend(t, func(t *testing.T, backend writeValuesBackend) {
		ctx := context.Background()
		manager, err := NewManagerWithDB[writeValuesModel](backend.table, backend.database)
		require.NoError(t, err)
		table := EscapeIdentifier(backend.table)

		instance := &writeValuesModel{Name: "before"}
		require.NoError(t, manager.Create(ctx, instance))
		_, err = backend.database.Exec(fmt.Sprintf(`UPDATE %s SET updated_at = %s, created_at = %s WHERE id = %d`,
			table, backend.past, backend.past, instance.ID))
		require.NoError(t, err)

		loaded, err := manager.Get(ctx, instance.ID)
		require.NoError(t, err)
		createdAt := loaded.CreatedAt
		past := loaded.UpdatedAt
		require.Equal(t, 2001, past.Year())

		// Update (and Save, which delegates to it) sets AutoNow in SQL
		// and on the struct, and leaves AutoNowAdd alone.
		loaded.Name = "after"
		before := time.Now().Add(-time.Minute)
		require.NoError(t, manager.Save(ctx, loaded))
		assert.True(t, loaded.UpdatedAt.After(before), "struct UpdatedAt = %v", loaded.UpdatedAt)

		stored, err := manager.Get(ctx, instance.ID)
		require.NoError(t, err)
		assert.True(t, stored.UpdatedAt.After(before), "stored updated_at = %v", stored.UpdatedAt)
		assert.WithinDuration(t, loaded.UpdatedAt, stored.UpdatedAt, time.Millisecond)
		assert.True(t, createdAt.Equal(stored.CreatedAt), "created_at changed: %v -> %v", createdAt, stored.CreatedAt)

		// A zero AutoNowAdd on an instance built by hand does not
		// overwrite the stored creation time.
		byHand := &writeValuesModel{ID: instance.ID, Name: "by hand"}
		require.NoError(t, manager.Update(ctx, byHand))
		stored, err = manager.Get(ctx, instance.ID)
		require.NoError(t, err)
		assert.True(t, createdAt.Equal(stored.CreatedAt), "created_at changed: %v -> %v", createdAt, stored.CreatedAt)
		assert.False(t, byHand.UpdatedAt.IsZero())

		// UpdateFields (the admin edit path) refreshes AutoNow too,
		// unless the caller sets it.
		_, err = backend.database.Exec(fmt.Sprintf(`UPDATE %s SET updated_at = %s WHERE id = %d`, table, backend.past, instance.ID))
		require.NoError(t, err)
		require.NoError(t, manager.UpdateFields(ctx, instance.ID, UpdateMap{"name": "fields"}))
		stored, err = manager.Get(ctx, instance.ID)
		require.NoError(t, err)
		assert.Equal(t, "fields", stored.Name)
		assert.True(t, stored.UpdatedAt.After(before), "stored updated_at = %v", stored.UpdatedAt)

		explicit := time.Date(2020, 5, 6, 7, 8, 9, 0, time.UTC)
		require.NoError(t, manager.UpdateFields(ctx, instance.ID, UpdateMap{"updated_at": explicit}))
		stored, err = manager.Get(ctx, instance.ID)
		require.NoError(t, err)
		assert.True(t, explicit.Equal(stored.UpdatedAt), "stored updated_at = %v", stored.UpdatedAt)
	})
}

// blankVoteModel is the review of #291: optional fields without a Default
// must still store NULL. ip_address takes part in a two-column unique
// constraint, where "" would collide and NULL does not; ref is a UUID column,
// where "" is not a valid value.
type blankVoteModel struct {
	schema.BaseSchema
	ID       int64  `db:"id"`
	ReviewID int64  `db:"review_id"`
	IP       string `db:"ip_address"`
	Ref      string `db:"ref"`
}

func (blankVoteModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("review_id", schema.Required()),
		schema.StringField("ip_address"),
		schema.UUIDField("ref"),
	}
}

func (blankVoteModel) Meta() schema.Meta { return schema.Meta{TableName: "blank_votes"} }

func TestManagerCreate_OptionalZeroWithoutDefaultStoresNull(t *testing.T) {
	run := func(t *testing.T, database *db.DB, table, ddl string) {
		_, err := database.Exec(fmt.Sprintf(ddl, EscapeIdentifier(table)))
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = database.Exec(`DROP TABLE IF EXISTS ` + EscapeIdentifier(table)) })

		ctx := context.Background()
		manager, err := NewManagerWithDB[blankVoteModel](table, database)
		require.NoError(t, err)
		require.NoError(t, manager.Create(ctx, &blankVoteModel{ReviewID: 1}))
		require.NoError(t, manager.Create(ctx, &blankVoteModel{ReviewID: 1}), "two blank votes must not collide")

		var nulls int
		require.NoError(t, database.QueryRow(fmt.Sprintf(
			`SELECT COUNT(*) FROM %s WHERE ip_address IS NULL AND ref IS NULL`, EscapeIdentifier(table),
		)).Scan(&nulls))
		assert.Equal(t, 2, nulls)
	}

	t.Run("sqlite", func(t *testing.T) {
		sqliteDB, err := db.NewDB(filepath.Join(t.TempDir(), "blank-votes.sqlite"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = sqliteDB.Close() })
		run(t, sqliteDB, "blank_votes", `CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			review_id INTEGER NOT NULL,
			ip_address TEXT,
			ref TEXT,
			UNIQUE (review_id, ip_address)
		)`)
	})

	t.Run("postgres", func(t *testing.T) {
		sqlDB := testutils.SetupTestDB(t)
		t.Cleanup(func() { _ = sqlDB.Close() })
		run(t, &db.DB{DB: sqlDB, Driver: "postgres"}, fmt.Sprintf("blank_votes_%d", time.Now().UnixNano()), `CREATE TABLE %s (
			id SERIAL PRIMARY KEY,
			review_id BIGINT NOT NULL,
			ip_address TEXT,
			ref UUID,
			UNIQUE (review_id, ip_address)
		)`)
	})
}
