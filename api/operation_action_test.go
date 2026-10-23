package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/operation"
	_ "github.com/mattn/go-sqlite3"
)

type operationActionPrincipal struct{ allowed bool }

func (operationActionPrincipal) ID() string   { return "test-user" }
func (operationActionPrincipal) Kind() string { return "user" }
func (p operationActionPrincipal) HasPermission(context.Context, string) (bool, error) {
	return p.allowed, nil
}

func operationActionRunner(t *testing.T) (*operation.Runner, *db.DB) {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "action.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE action_records (value TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	sink := audit.NewSQLSink(database)
	if err := sink.EnsureSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return operation.NewRunner(database, sink), database
}

func TestOperationAction(t *testing.T) {
	for _, tc := range []struct {
		name       string
		allowed    bool
		runErr     error
		wantStatus int
		wantResult bool
		wantRows   int
	}{
		{name: "allowed", allowed: true, wantStatus: 200, wantResult: true, wantRows: 1},
		{name: "denied", wantStatus: 403},
		{name: "operation error", allowed: true, runErr: errors.New("failed"), wantStatus: 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runner, database := operationActionRunner(t)
			op := operation.Operation[string, map[string]string]{Name: "records.create", Permission: "records.add", Run: func(ctx context.Context, tx operation.Transaction, _ operation.Actor, input string) (map[string]string, []operation.Effect, error) {
				if tc.runErr != nil {
					return nil, nil, tc.runErr
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO action_records VALUES (?)`, input); err != nil {
					return nil, nil, err
				}
				return map[string]string{"value": input}, nil, nil
			}}
			handler := OperationAction(runner, op, func(r *http.Request) (operation.Actor, error) {
				principal, _ := core.UserFromContext(r.Context())
				actor, ok := principal.(operationActionPrincipal)
				if !ok {
					return nil, nil
				}
				return actor, nil
			})
			req := httptest.NewRequest("POST", "/items", strings.NewReader(`"created"`))
			req = req.WithContext(core.WithUser(req.Context(), operationActionPrincipal{allowed: tc.allowed}))
			response := httptest.NewRecorder()
			handler(response, req)
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, body %s", response.Code, response.Body.String())
			}
			var rows int
			if err := database.QueryRow(`SELECT count(*) FROM action_records`).Scan(&rows); err != nil {
				t.Fatal(err)
			}
			if rows != tc.wantRows {
				t.Fatalf("rows = %d, want %d", rows, tc.wantRows)
			}
			if tc.wantResult {
				var result map[string]string
				if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result["value"] != "created" {
					t.Fatalf("result = %#v", result)
				}
			}
		})
	}
}
