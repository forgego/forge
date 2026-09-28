package rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	validation "github.com/forgego/forge/validate"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decodeWriteError(t *testing.T, rec *httptest.ResponseRecorder) (string, string, map[string][]string) {
	t.Helper()
	var body struct {
		Error struct {
			Code    string              `json:"code"`
			Message string              `json:"message"`
			Details map[string][]string `json:"details"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body.Error.Code, body.Error.Message, body.Error.Details
}

func TestRespondWriteError_PostgresUniqueViolationIsFieldConflict(t *testing.T) {
	err := fmt.Errorf("insert failed: %w", &pq.Error{
		Code:       "23505",
		Message:    `duplicate key value violates unique constraint "categories_slug_key"`,
		Detail:     "Key (slug)=(shoes) already exists.",
		Constraint: "categories_slug_key",
	})

	rec := httptest.NewRecorder()
	respondWriteError(rec, "create_failed", err)

	assert.Equal(t, http.StatusConflict, rec.Code)
	code, msg, details := decodeWriteError(t, rec)
	assert.Equal(t, "conflict", code)
	assert.Equal(t, "A record with this slug already exists.", msg)
	assert.Equal(t, []string{msg}, details["slug"])
	assert.NotContains(t, rec.Body.String(), "categories_slug_key")
}

func TestRespondWriteError_PostgresForeignKeyViolationIsBadRequest(t *testing.T) {
	err := fmt.Errorf("update query failed: %w", &pq.Error{
		Code:    "23503",
		Message: `insert or update on table "categories" violates foreign key constraint "categories_parent_id_fkey"`,
		Detail:  `Key (parent_id)=(0) is not present in table "categories".`,
	})

	rec := httptest.NewRecorder()
	respondWriteError(rec, "update_failed", err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	code, msg, details := decodeWriteError(t, rec)
	assert.Equal(t, "invalid_reference", code)
	assert.Equal(t, "The selected parent_id does not exist.", msg)
	assert.Equal(t, []string{msg}, details["parent_id"])
	assert.NotContains(t, rec.Body.String(), "categories_parent_id_fkey")
}

func TestRespondWriteError_SQLiteUniqueViolationIsFieldConflict(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE brands (id INTEGER PRIMARY KEY, slug TEXT UNIQUE)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO brands (slug) VALUES ('acme')`)
	require.NoError(t, err)
	_, dupErr := db.Exec(`INSERT INTO brands (slug) VALUES ('acme')`)
	require.Error(t, dupErr)

	rec := httptest.NewRecorder()
	respondWriteError(rec, "create_failed", fmt.Errorf("insert failed: %w", dupErr))

	assert.Equal(t, http.StatusConflict, rec.Code)
	code, _, details := decodeWriteError(t, rec)
	assert.Equal(t, "conflict", code)
	assert.Contains(t, details, "slug")
}

func TestRespondWriteError_TypedValidationErrorIsBadRequest(t *testing.T) {
	// "unknown field" carries no keyword the string heuristic recognises.
	verrs := &validation.ValidationErrors{}
	verrs.Add("bogus", "unknown field")

	rec := httptest.NewRecorder()
	respondWriteError(rec, "create_failed", verrs)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	code, _, details := decodeWriteError(t, rec)
	assert.Equal(t, "validation_error", code)
	assert.Equal(t, []string{"unknown field"}, details["bogus"])
}

func TestRespondWriteError_UnexpectedErrorDoesNotLeakDriverText(t *testing.T) {
	rec := httptest.NewRecorder()
	respondWriteError(rec, "update_failed", errors.New(`pq: relation "secret_table" does not exist`))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	code, msg, _ := decodeWriteError(t, rec)
	assert.Equal(t, "update_failed", code)
	assert.False(t, strings.Contains(msg, "secret_table"))
}
