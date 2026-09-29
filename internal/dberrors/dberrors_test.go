package dberrors

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyPostgres(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  *pq.Error
		want Violation
		ok   bool
	}{
		{"unique with key detail", &pq.Error{Code: "23505", Detail: "Key (slug)=(shoes) already exists."}, Violation{Unique, "slug"}, true},
		{"composite key", &pq.Error{Code: "23505", Detail: "Key (a, b)=(1, 2) already exists."}, Violation{Unique, "a, b"}, true},
		{"foreign key", &pq.Error{Code: "23503", Detail: "Key (category_id)=(9) is not present in table \"categories\"."}, Violation{ForeignKey, "category_id"}, true},
		{"not null uses column", &pq.Error{Code: "23502", Column: "name"}, Violation{NotNull, "name"}, true},
		{"check", &pq.Error{Code: "23514"}, Violation{Check, ""}, true},
		{"data exception", &pq.Error{Code: "22001", Column: "slug"}, Violation{DataException, "slug"}, true},
		{"undefined column is not a violation", &pq.Error{Code: "42703"}, Violation{}, false},
		{"exclusion is not classified", &pq.Error{Code: "23P01"}, Violation{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Classify(fmt.Errorf("wrapped: %w", tc.err))
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
			assert.True(t, IsDriverError(tc.err))
		})
	}
}

func TestClassifySQLite(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE parents (id INTEGER PRIMARY KEY);
		CREATE TABLE items (
			id INTEGER PRIMARY KEY,
			slug TEXT UNIQUE,
			name TEXT NOT NULL,
			qty INTEGER CHECK (qty >= 0),
			parent_id INTEGER REFERENCES parents(id)
		);
		INSERT INTO items (id, slug, name, qty) VALUES (1, 'a', 'x', 1);`)
	require.NoError(t, err)

	for _, tc := range []struct {
		name, sql string
		want      Violation
	}{
		{"unique", `INSERT INTO items (slug, name) VALUES ('a', 'y')`, Violation{Unique, "slug"}},
		{"primary key", `INSERT INTO items (id, name) VALUES (1, 'y')`, Violation{Unique, "id"}},
		{"not null", `INSERT INTO items (slug) VALUES ('b')`, Violation{NotNull, "name"}},
		{"check", `INSERT INTO items (name, qty) VALUES ('y', -1)`, Violation{Check, ""}},
		{"foreign key", `INSERT INTO items (name, parent_id) VALUES ('y', 42)`, Violation{ForeignKey, ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, execErr := db.Exec(tc.sql)
			require.Error(t, execErr)
			got, ok := Classify(fmt.Errorf("wrapped: %w", execErr))
			require.True(t, ok, "%v", execErr)
			assert.Equal(t, tc.want.Kind, got.Kind, "%v", execErr)
			assert.Equal(t, tc.want.Field, got.Field, "%v", execErr)
			assert.True(t, IsDriverError(execErr))
		})
	}
}

func TestClassifyOtherErrors(t *testing.T) {
	_, ok := Classify(errors.New("UNIQUE constraint failed: items.slug"))
	assert.False(t, ok, "only driver errors are classified, not look-alike text")
	assert.False(t, IsDriverError(errors.New("boom")))
	_, ok = Classify(nil)
	assert.False(t, ok)
}
