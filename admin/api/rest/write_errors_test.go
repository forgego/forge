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

	"github.com/forgego/forge/admin/core"
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
	// "unknown field" carries no keyword the string heuristic recognizes.
	verrs := &validation.ValidationErrors{}
	verrs.Add("bogus", "unknown field")

	rec := httptest.NewRecorder()
	respondWriteError(rec, "create_failed", verrs)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	code, _, details := decodeWriteError(t, rec)
	assert.Equal(t, "validation_error", code)
	assert.Equal(t, []string{"unknown field"}, details["bogus"])
}

func TestDeleteFailure_ReferencedRecordIsConflictWithoutDriverText(t *testing.T) {
	err := fmt.Errorf("delete failed: %w", &pq.Error{
		Code:    "23503",
		Message: `update or delete on table "categories" violates foreign key constraint "categories_parent_id_fkey" on table "categories"`,
		Detail:  `Key (id)=(1) is still referenced from table "categories".`,
	})

	status, code, message := deleteFailure(err)
	assert.Equal(t, http.StatusConflict, status)
	assert.Equal(t, "in_use", code)
	assert.NotContains(t, message, "categories_parent_id_fkey")

	status, code, message = deleteFailure(errors.New("pq: connection reset"))
	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Equal(t, "delete_failed", code)
	assert.NotContains(t, message, "pq:")
}

func TestRespondWriteError_UnexpectedErrorDoesNotLeakDriverText(t *testing.T) {
	rec := httptest.NewRecorder()
	respondWriteError(rec, "update_failed", errors.New(`pq: relation "secret_table" does not exist`))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	code, msg, _ := decodeWriteError(t, rec)
	assert.Equal(t, "update_failed", code)
	assert.False(t, strings.Contains(msg, "secret_table"))
}

func TestRespondWriteError_PostgresDataExceptionIsValidationWithoutDriverText(t *testing.T) {
	cases := []struct {
		name  string
		err   *pq.Error
		field string
	}{
		{"invalid timestamp", &pq.Error{Code: "22007", Message: `invalid input syntax for type timestamp: "secret-input"`}, "non_field_errors"},
		{"invalid integer", &pq.Error{Code: "22P02", Message: `invalid input syntax for type integer: "secret-input"`}, "non_field_errors"},
		{"value too long", &pq.Error{Code: "22001", Message: `value too long for type character varying(2) secret-input`}, "non_field_errors"},
		{"column named", &pq.Error{Code: "22001", Message: `value too long secret-input`, Column: "slug"}, "slug"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			respondWriteError(rec, "create_failed", fmt.Errorf("insert failed: %w", tc.err))

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			code, msg, details := decodeWriteError(t, rec)
			assert.Equal(t, "validation_error", code)
			assert.Equal(t, []string{msg}, details[tc.field])
			assert.NotContains(t, rec.Body.String(), "secret-input")
			assert.NotContains(t, rec.Body.String(), "syntax")
		})
	}
}

func TestRespondWriteError_DriverErrorsSkipKeywordFallback(t *testing.T) {
	// Both messages contain a keyword isValidationError matches
	// ("required"), but a driver error is never echoed to the client.
	pgErr := &pq.Error{Code: "42703", Message: `column "is_required_secret" does not exist`}

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	_, liteErr := db.Exec(`SELECT is_required_secret FROM sqlite_master`)
	require.Error(t, liteErr)

	for name, driverErr := range map[string]error{"postgres": pgErr, "sqlite": liteErr} {
		t.Run(name, func(t *testing.T) {
			require.True(t, isValidationError(driverErr), "precondition: keyword heuristic matches")
			rec := httptest.NewRecorder()
			respondWriteError(rec, "update_failed", fmt.Errorf("update failed: %w", driverErr))

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			code, _, _ := decodeWriteError(t, rec)
			assert.Equal(t, "update_failed", code)
			assert.NotContains(t, rec.Body.String(), "is_required_secret")
		})
	}
}

func TestRespondWriteError_ORMValidationErrorStillBadRequest(t *testing.T) {
	// orm.Manager wraps Clean/Validate failures as "validation failed: ...";
	// they are not driver errors, so the keyword fallback still applies.
	err := fmt.Errorf("validation failed: %w", errors.New("name is required"))

	rec := httptest.NewRecorder()
	respondWriteError(rec, "create_failed", err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	code, msg, _ := decodeWriteError(t, rec)
	assert.Equal(t, "validation_error", code)
	assert.Equal(t, "validation failed: name is required", msg)
}

type bulkErrorBody struct {
	Errors []bulkItemError `json:"errors"`
	Error  struct {
		Code    string `json:"code"`
		Details struct {
			Errors []bulkItemError `json:"errors"`
		} `json:"details"`
	} `json:"error"`
}

func decodeBulkErrors(t *testing.T, rec *httptest.ResponseRecorder) bulkErrorBody {
	t.Helper()
	var body bulkErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

var slugConflict = &pq.Error{
	Code:       "23505",
	Message:    `duplicate key value violates unique constraint "products_slug_key"`,
	Detail:     "Key (slug)=(dup) already exists.",
	Constraint: "products_slug_key",
}

func TestHandleBulkCreate_ClassifiesItemErrorsWithoutDriverText(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			switch data["name"] {
			case "dup":
				return nil, fmt.Errorf("insert failed: %w", slugConflict)
			case "broken":
				return nil, fmt.Errorf("insert failed: %w", &pq.Error{Code: "42P01", Message: `relation "secret_table" does not exist`})
			}
			return map[string]interface{}{"name": data["name"]}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", strings.NewReader(`[{"name":"ok"},{"name":"dup"},{"name":"broken"}]`))
	rec := httptest.NewRecorder()
	router.handleBulkCreate(admin)(rec, req)

	require.Equal(t, http.StatusMultiStatus, rec.Code)
	assert.NotContains(t, rec.Body.String(), "products_slug_key")
	assert.NotContains(t, rec.Body.String(), "secret_table")
	body := decodeBulkErrors(t, rec)
	require.Len(t, body.Errors, 2)
	assert.Equal(t, bulkItemError{Index: 1, Code: "conflict", Message: "A record with this slug already exists."}, body.Errors[0])
	assert.Equal(t, "create_failed", body.Errors[1].Code)
}

func TestHandleBulkCreate_AllConflictsReturn409(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("insert failed: %w", slugConflict)
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", strings.NewReader(`[{"name":"a"},{"name":"b"}]`))
	rec := httptest.NewRecorder()
	router.handleBulkCreate(admin)(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.NotContains(t, rec.Body.String(), "products_slug_key")
	body := decodeBulkErrors(t, rec)
	assert.Equal(t, "create_failed", body.Error.Code)
	require.Len(t, body.Error.Details.Errors, 2)
	assert.Equal(t, "conflict", body.Error.Details.Errors[0].Code)
}

func TestHandleBulkUpdate_AllInvalidReferencesReturn400WithoutDriverText(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("update query failed: %w", &pq.Error{
				Code:    "23503",
				Message: `insert or update on table "products" violates foreign key constraint "products_brand_id_fkey"`,
				Detail:  `Key (brand_id)=(99) is not present in table "brands".`,
			})
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", strings.NewReader(`{"ids":[1,2],"data":{"brand_id":99}}`))
	rec := httptest.NewRecorder()
	router.handleBulkUpdate(admin)(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.NotContains(t, rec.Body.String(), "products_brand_id_fkey")
	body := decodeBulkErrors(t, rec)
	require.Len(t, body.Error.Details.Errors, 2)
	assert.Equal(t, bulkItemError{Index: 0, Code: "invalid_reference", Message: "The selected brand_id does not exist."}, body.Error.Details.Errors[0])
}

func TestHandleBulkUpdate_UnclassifiedErrorIsGenericServerError(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			return nil, errors.New(`pq: relation "secret_table" does not exist`)
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", strings.NewReader(`{"ids":[1],"data":{"active":true}}`))
	rec := httptest.NewRecorder()
	router.handleBulkUpdate(admin)(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret_table")
	body := decodeBulkErrors(t, rec)
	require.Len(t, body.Error.Details.Errors, 1)
	assert.Equal(t, "update_failed", body.Error.Details.Errors[0].Code)
}

func TestHandleBulkDelete_AllInUseReturn409(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		deleteObjectFn: func(id interface{}) error {
			return fmt.Errorf("delete failed: %w", &pq.Error{
				Code:   "23503",
				Detail: `Key (id)=(1) is still referenced from table "products".`,
			})
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", strings.NewReader(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()
	router.handleBulkDelete(admin)(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeBulkErrors(t, rec)
	require.Len(t, body.Error.Details.Errors, 2)
	assert.Equal(t, "in_use", body.Error.Details.Errors[0].Code)
}

func TestBulkFailure_MixedClientErrorsReturn400AndAnyServerErrorReturns500(t *testing.T) {
	rec := httptest.NewRecorder()
	bulkFailure(rec, "delete_failed", "x", []bulkItemError{{Code: "in_use"}, {Code: "permission_denied"}})
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = httptest.NewRecorder()
	bulkFailure(rec, "delete_failed", "x", []bulkItemError{{Code: "delete_failed"}, {Code: "in_use"}, {Code: "not_found"}})
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
