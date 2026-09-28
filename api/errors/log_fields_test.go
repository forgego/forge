package errors

import (
	"database/sql"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// logOutput renders every logged entry and its fields as one string.
func logOutput(t *testing.T, logs *observer.ObservedLogs) string {
	t.Helper()
	var b strings.Builder
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	for _, entry := range logs.AllUntimed() {
		buf, err := enc.EncodeEntry(entry.Entry, entry.Context)
		require.NoError(t, err)
		b.Write(buf.Bytes())
		buf.Free()
	}
	return b.String()
}

func observedHandler() (*Handler, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.DebugLevel)
	cfg := DefaultHandlerConfig()
	cfg.Logger = zap.New(core)
	return NewHandler(cfg), logs
}

// Regression for #290: the error handler logs a driver error's type,
// SQLSTATE and constraint, not its message or detail, which carry row values.
func TestHandleError_LogsDriverErrorWithoutValues(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		leaks    []string
		expected map[string]string
	}{
		{
			name: "postgres unique violation",
			err: fmt.Errorf("create user: %w", &pq.Error{
				Code:       "23505",
				Message:    `duplicate key value violates unique constraint "users_email_key"`,
				Detail:     "Key (email)=(alice@example.com) already exists.",
				Table:      "users",
				Constraint: "users_email_key",
			}),
			leaks: []string{"alice@example.com", "duplicate key value"},
			expected: map[string]string{
				"sqlstate":      "23505",
				"sqlstate_name": "unique_violation",
				"db_constraint": "users_email_key",
				"db_table":      "users",
				"db_driver":     "postgres",
			},
		},
		{
			name: "postgres invalid input carries the value in its message",
			err: &pq.Error{
				Code:    "22P02",
				Message: `invalid input syntax for type integer: "4111-1111-1111-1111"`,
			},
			leaks: []string{"4111-1111-1111-1111"},
			expected: map[string]string{
				"sqlstate":      "22P02",
				"sqlstate_name": "invalid_text_representation",
			},
		},
		{
			name: "sqlite unique violation",
			err: fmt.Errorf("create user: %w", sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintUnique,
			}),
			expected: map[string]string{"db_driver": "sqlite3"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, logs := observedHandler()
			r := httptest.NewRequest(http.MethodPost, "/users", nil)
			h.HandleError(httptest.NewRecorder(), r, tc.err)

			require.Equal(t, 1, logs.Len())
			fields := logs.All()[0].ContextMap()
			for key, want := range tc.expected {
				assert.Equal(t, want, fields[key], "field %s", key)
			}
			assert.NotContains(t, fields, "error", "raw driver error text must not be logged")
			assert.NotEmpty(t, fields["error_class"])

			out := logOutput(t, logs)
			for _, leak := range tc.leaks {
				assert.NotContains(t, out, leak)
			}
		})
	}
}

func TestHandleError_LogsSQLiteConstraintTarget(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE users (email TEXT UNIQUE)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (email) VALUES ('alice@example.com')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO users (email) VALUES ('alice@example.com')`)
	require.Error(t, err)

	h, logs := observedHandler()
	h.HandleError(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/users", nil), err)
	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.Equal(t, "users.email", fields["db_constraint"])
	assert.EqualValues(t, sqlite3.ErrConstraint, fields["sqlite_code"])
	assert.EqualValues(t, sqlite3.ErrConstraintUnique, fields["sqlite_extended_code"])
	assert.NotContains(t, logOutput(t, logs), "alice@example.com")
}

func TestHandleError_LogsOtherErrorsVerbatim(t *testing.T) {
	h, logs := observedHandler()
	h.HandleError(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil),
		stderrors.New("cache backend timeout\nsecond line"))
	require.Equal(t, 1, logs.Len())
	fields := logs.All()[0].ContextMap()
	assert.Equal(t, "cache backend timeoutsecond line", fields["error"])
	assert.Equal(t, "*errors.errorString", fields["error_class"])
}

func TestHandlePanic_DriverErrorWithoutValues(t *testing.T) {
	h, logs := observedHandler()
	h.HandlePanic(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), &pq.Error{
		Code:       "23505",
		Message:    `duplicate key value violates unique constraint "users_email_key"`,
		Detail:     "Key (email)=(alice@example.com) already exists.",
		Constraint: "users_email_key",
	})
	require.Equal(t, 1, logs.Len())
	out := logOutput(t, logs)
	assert.NotContains(t, out, "alice@example.com")
	assert.NotContains(t, out, "duplicate key value")
	assert.Contains(t, out, "users_email_key")
	assert.Contains(t, out, "23505")
}
