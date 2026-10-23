package operation

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	_ "github.com/mattn/go-sqlite3"
)

type testActor struct {
	id         string
	allowed    bool
	checkErr   error
	permission string
}

type countingBeginner struct {
	database *db.DB
	calls    int
}

func (b *countingBeginner) Begin(ctx context.Context) (ManagedTransaction, error) {
	b.calls++
	return b.database.BeginTx(ctx, nil)
}

func (a *testActor) ID() string { return a.id }
func (*testActor) Kind() string { return "user" }
func (a *testActor) HasPermission(_ context.Context, permission string) (bool, error) {
	a.permission = permission
	return a.allowed, a.checkErr
}

func testRunner(t *testing.T, sink audit.Sink) (*Runner, *db.DB) {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "operations.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE operation_records (value TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if sink == nil {
		sqlSink := audit.NewSQLSink(database)
		if err := sqlSink.EnsureSchema(context.Background()); err != nil {
			t.Fatal(err)
		}
		sink = sqlSink
	}
	return NewRunner(database, sink), database
}

func authorizedContext() context.Context {
	return WithActor(context.Background(), &testActor{id: "actor-7", allowed: true})
}

func TestRunDeniedDoesNotRunBusinessOperationAndAuditsDenial(t *testing.T) {
	_, database := testRunner(t, nil)
	beginner := &countingBeginner{database: database}
	runner := NewRunner(database, audit.NewSQLSink(database), WithBeginner(beginner))
	actor := &testActor{id: "actor-7"}
	called := false
	op := Operation[struct{}, struct{}]{Name: "records.create", Permission: "records.add", Run: func(context.Context, Transaction, Actor, struct{}) (struct{}, []Effect, error) {
		called = true
		return struct{}{}, nil, nil
	}}
	_, err := Run(WithActor(context.Background(), actor), runner, op, struct{}{})
	if !errors.Is(err, ErrDenied) || called {
		t.Fatalf("Run() error = %v, business called = %v", err, called)
	}
	if actor.permission != op.Permission {
		t.Fatalf("checked permission = %q, want %q", actor.permission, op.Permission)
	}
	var outcomes int
	if err := database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE outcome = 'denied' AND actor_id = 'actor-7'`).Scan(&outcomes); err != nil {
		t.Fatal(err)
	}
	if outcomes != 1 {
		t.Fatalf("denied audit rows = %d, want 1", outcomes)
	}
	if beginner.calls != 1 {
		t.Fatalf("transactions begun = %d, want only the separate denial-audit transaction", beginner.calls)
	}
	var rows int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("business rows = %d, want 0", rows)
	}
}

func TestRunCommitsDataAndAuditBeforeRunningEffects(t *testing.T) {
	runner, database := testRunner(t, nil)
	effects := 0
	op := Operation[string, string]{
		Name: "records.create", Permission: "records.add",
		Audit: AuditDetails{CorrelationID: "request-1", Source: "test"},
		Run: func(ctx context.Context, tx Transaction, _ Actor, input string) (string, []Effect, error) {
			if _, err := tx.ExecContext(ctx, `INSERT INTO operation_records (value) VALUES (?)`, input); err != nil {
				return "", nil, err
			}
			return input, []Effect{func(context.Context) error {
				var rows int
				if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&rows); err != nil {
					return err
				}
				if rows != 1 {
					return fmt.Errorf("effect ran before commit")
				}
				effects++
				return nil
			}}, nil
		},
	}
	got, err := Run(authorizedContext(), runner, op, "created")
	if err != nil || got != "created" {
		t.Fatalf("Run() = %q, %v", got, err)
	}
	if effects != 1 {
		t.Fatalf("effects ran %d times, want 1", effects)
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("business row count = %d, err = %v", count, err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE outcome = 'success' AND operation = 'records.create'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("success audit count = %d, err = %v", count, err)
	}
}

func TestRunErrorRollsBackBusinessAndSkipsEffects(t *testing.T) {
	runner, database := testRunner(t, nil)
	effects := 0
	opErr := errors.New("operation failed")
	op := Operation[struct{}, struct{}]{Name: "records.create", Permission: "records.add", Run: func(ctx context.Context, tx Transaction, _ Actor, _ struct{}) (struct{}, []Effect, error) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO operation_records (value) VALUES ('partial')`); err != nil {
			return struct{}{}, nil, err
		}
		return struct{}{}, []Effect{func(context.Context) error { effects++; return nil }}, opErr
	}}
	_, err := Run(authorizedContext(), runner, op, struct{}{})
	if !errors.Is(err, opErr) {
		t.Fatalf("Run() error = %v, want wrapped operation error", err)
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("business rows = %d, err = %v; want rollback", count, err)
	}
	if effects != 0 {
		t.Fatalf("effects ran %d times, want 0", effects)
	}
	if err := database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE outcome = 'success'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("success audit rows = %d, err = %v", count, err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE outcome = 'failed'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("failed audit rows = %d, err = %v", count, err)
	}
}

type failingSink struct{ err error }

func (s failingSink) Record(context.Context, audit.DBTX, audit.Event) error { return s.err }

func TestRunAuditFailureRollsBack(t *testing.T) {
	runner, database := testRunner(t, failingSink{err: errors.New("audit unavailable")})
	op := Operation[struct{}, struct{}]{Name: "records.create", Permission: "records.add", Run: func(ctx context.Context, tx Transaction, _ Actor, _ struct{}) (struct{}, []Effect, error) {
		_, err := tx.ExecContext(ctx, `INSERT INTO operation_records (value) VALUES ('must rollback')`)
		return struct{}{}, nil, err
	}}
	_, err := Run(authorizedContext(), runner, op, struct{}{})
	if err == nil {
		t.Fatal("Run() error = nil, want fail-closed audit error")
	}
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("business rows = %d, err = %v; want rollback", count, err)
	}
}

func TestRunPanicRollsBack(t *testing.T) {
	runner, database := testRunner(t, nil)
	op := Operation[struct{}, struct{}]{Name: "records.create", Permission: "records.add", Run: func(ctx context.Context, tx Transaction, _ Actor, _ struct{}) (struct{}, []Effect, error) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO operation_records (value) VALUES ('panic')`); err != nil {
			return struct{}{}, nil, err
		}
		panic("boom")
	}}
	func() {
		defer func() {
			if recovered := recover(); recovered != "boom" {
				t.Fatalf("panic = %v, want boom", recovered)
			}
		}()
		_, _ = Run(authorizedContext(), runner, op, struct{}{})
	}()
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM operation_records`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("business rows = %d, err = %v; want rollback", count, err)
	}
	if err := database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE outcome = 'failed'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("failed audit rows = %d, err = %v", count, err)
	}
}
