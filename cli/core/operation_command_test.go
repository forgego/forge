package core

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
	"github.com/stretchr/testify/require"
)

func TestOperationCommand(t *testing.T) {
	failure := errors.New("callback failure")
	for _, tc := range []struct {
		name, input, category               string
		wantCode                            int
		noPrincipal, denied, fail, conflict bool
	}{
		{name: "success", input: `{"value":"ok"}`},
		{name: "deny", input: `{"value":"ok"}`, category: "denied", wantCode: OperationExitDenied, denied: true},
		{name: "missing principal", input: `{"value":"ok"}`, category: "denied", wantCode: OperationExitDenied, noPrincipal: true},
		{name: "invalid input", input: `{`, category: "invalid_input", wantCode: OperationExitInvalidInput},
		{name: "callback failure", input: `{"value":"ok"}`, category: "failed", wantCode: OperationExitFailed, fail: true},
		{name: "conflict", input: `{"value":"ok"}`, category: OperationCategoryConflict, wantCode: OperationExitConflict, conflict: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "operation.sqlite"))
			require.NoError(t, err)
			defer database.Close()
			require.NoError(t, audit.NewSQLSink(database).EnsureSchema(context.Background()))
			runner := operation.NewRunner(database, audit.NewSQLSink(database))
			var output bytes.Buffer
			op := operation.Operation[struct {
				Value string `json:"value"`
			}, struct {
				Value string `json:"value"`
			}]{Name: "cli.test", Permission: "test.run", Run: func(_ context.Context, _ operation.Transaction, _ operation.Actor, in struct {
				Value string `json:"value"`
			}) (struct {
				Value string `json:"value"`
			}, []operation.Effect, error) {
				if tc.conflict {
					return struct {
						Value string `json:"value"`
					}{}, nil, errors.Join(errors.New("stale state"), operation.ErrConflict)
				}
				if tc.fail {
					return struct {
						Value string `json:"value"`
					}{}, nil, failure
				}
				return in, nil, nil
			}}
			cli := NewContext()
			cli.Principal.PermissionNames = []string{"test.run"}
			if tc.noPrincipal {
				cli.Principal = nil
			}
			if tc.denied {
				p := principal.System("other")
				cli.Principal = &p
			}
			err = (OperationCommand[struct {
				Value string `json:"value"`
			}, struct {
				Value string `json:"value"`
			}]{Runner: runner, Operation: op, Input: bytes.NewBufferString(tc.input), Output: &output}).Execute(context.Background(), cli)
			if tc.category == "" {
				require.NoError(t, err)
				require.JSONEq(t, tc.input, output.String())
				return
			}
			var commandErr *OperationError
			require.ErrorAs(t, err, &commandErr)
			require.Equal(t, tc.category, commandErr.Category)
			require.Equal(t, tc.wantCode, commandErr.ExitCode)
			if tc.fail {
				require.ErrorIs(t, err, failure)
			}
		})
	}
}
