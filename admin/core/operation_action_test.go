package core

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
	_ "github.com/mattn/go-sqlite3"
)

func operationActionFixture(t *testing.T) (*operation.Runner, *db.DB) {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "admin-operation.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE operation_records (value TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	sink := audit.NewSQLSink(database)
	if err := sink.EnsureSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return operation.NewRunner(database, sink), database
}

func TestOperationActionAllowedReturnsPerCallResult(t *testing.T) {
	runner, database := operationActionFixture(t)
	action := OperationAction[string, string]{Runner: runner, RequiresConfirmation: true, Operation: operation.Operation[string, string]{Name: "records.create", Permission: "records.add", Run: func(ctx context.Context, tx operation.Transaction, actor operation.Actor, input string) (string, []operation.Effect, error) {
		if actor.ID() != "user-1" {
			t.Fatalf("actor ID = %q", actor.ID())
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO operation_records (value) VALUES (?)`, input); err != nil {
			return "", nil, err
		}
		return "created:" + input, nil, nil
	}}}
	ctx := core.WithPrincipal(context.Background(), principal.Principal{ID: "user-1", Kind: principal.KindUser, PermissionNames: []string{"records.add"}})
	got, err := action.Execute(ctx, "one", true)
	if err != nil || got != "created:one" {
		t.Fatalf("Execute() = %q, %v", got, err)
	}
	got, err = action.Execute(ctx, "two", true)
	if err != nil || got != "created:two" {
		t.Fatalf("second Execute() = %q, %v", got, err)
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("record count = %d, %v", count, err)
	}
}

func TestOperationActionDeniedDoesNotChangeData(t *testing.T) {
	runner, database := operationActionFixture(t)
	called := false
	action := OperationAction[string, string]{Runner: runner, Operation: operation.Operation[string, string]{Name: "records.create", Permission: "records.add", Run: func(context.Context, operation.Transaction, operation.Actor, string) (string, []operation.Effect, error) {
		called = true
		return "", nil, nil
	}}}
	ctx := core.WithPrincipal(context.Background(), principal.Principal{ID: "user-1", Kind: principal.KindUser})
	_, err := action.Execute(ctx, "data", false)
	var httpErr *ActionHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != 403 {
		t.Fatalf("Execute() error = %v, want HTTP 403", err)
	}
	if called {
		t.Fatal("operation callback ran on denial")
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("record count = %d, %v", count, err)
	}
}

func TestOperationActionRequiresConfirmation(t *testing.T) {
	runner, _ := operationActionFixture(t)
	action := OperationAction[string, string]{Runner: runner, RequiresConfirmation: true, Operation: operation.Operation[string, string]{Name: "records.create", Permission: "records.add", Run: func(context.Context, operation.Transaction, operation.Actor, string) (string, []operation.Effect, error) {
		return "", nil, nil
	}}}
	_, err := action.Execute(context.Background(), "data", false)
	var httpErr *ActionHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != 400 {
		t.Fatalf("Execute() error = %v, want HTTP 400", err)
	}
}

func TestOperationActionReturnsOperationError(t *testing.T) {
	runner, _ := operationActionFixture(t)
	wantErr := errors.New("business failure")
	action := OperationAction[string, string]{Runner: runner, Operation: operation.Operation[string, string]{Name: "records.create", Permission: "records.add", Run: func(context.Context, operation.Transaction, operation.Actor, string) (string, []operation.Effect, error) {
		return "", nil, wantErr
	}}}
	ctx := core.WithPrincipal(context.Background(), principal.Principal{ID: "user-1", Kind: principal.KindUser, PermissionNames: []string{"records.add"}})
	_, err := action.Execute(ctx, "data", true)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want wrapped operation error", err)
	}
}

func TestOperationActionConflictReturnsHTTPConflict(t *testing.T) {
	runner, _ := operationActionFixture(t)
	action := OperationAction[string, string]{Runner: runner, Operation: operation.Operation[string, string]{Name: "records.update", Permission: "records.add", Run: func(context.Context, operation.Transaction, operation.Actor, string) (string, []operation.Effect, error) {
		return "", nil, errors.Join(errors.New("stale state"), operation.ErrConflict)
	}}}
	ctx := core.WithPrincipal(context.Background(), principal.Principal{ID: "user-1", Kind: principal.KindUser, PermissionNames: []string{"records.add"}})
	_, err := action.Execute(ctx, "data", true)
	var httpErr *ActionHTTPError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != 409 || !errors.Is(err, operation.ErrConflict) {
		t.Fatalf("Execute() error = %v, want wrapped HTTP 409 conflict", err)
	}
}
