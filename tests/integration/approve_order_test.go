package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	admincore "github.com/forgego/forge/admin/core"
	forgeapi "github.com/forgego/forge/api"
	apicore "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
	"github.com/forgego/forge/tests/testdata/reference_app/operations"
	_ "github.com/mattn/go-sqlite3"
)

type approvalSnapshot struct {
	OrderStatus string
	ApprovedBy  sql.NullString
	History     []historyRow
	Audit       []auditRow
}

type historyRow struct {
	OrderID   int64
	Status    string
	ChangedBy string
}

type auditRow struct {
	Operation string
	ActorID   string
	ActorKind string
	Model     sql.NullString
	Outcome   string
	Source    string
}

type entryPoint struct {
	name string
	run  func(context.Context, int64, principal.Principal) error
}

func TestApproveOrderAcrossAdapters(t *testing.T) {
	var baseline *approvalSnapshot
	for _, adapterName := range []string{"admin", "rest", "cli"} {
		t.Run(adapterName, func(t *testing.T) {
			fixture := newApprovalFixture(t)
			entry := approvalEntryPoint(t, adapterName, fixture.runner)
			if err := entry.run(context.Background(), 1, approvalPrincipal(true)); err != nil {
				t.Fatalf("approve order: %v", err)
			}
			got := fixture.snapshot(t)
			if got.OrderStatus != "approved" || got.ApprovedBy.String != "approver" || len(got.History) != 1 || len(got.Audit) != 1 {
				t.Fatalf("unexpected approval state: %+v", got)
			}
			if got.History[0] != (historyRow{OrderID: 1, Status: "approved", ChangedBy: "approver"}) {
				t.Fatalf("history row = %+v", got.History[0])
			}
			if baseline == nil {
				baseline = &got
			} else if !reflect.DeepEqual(got, *baseline) {
				t.Fatalf("adapter result differs from baseline:\n got: %+v\nwant: %+v", got, *baseline)
			}
		})
	}
}

func TestApproveOrderDenialsPreserveBusinessRows(t *testing.T) {
	for _, adapterName := range []string{"admin", "rest", "cli"} {
		t.Run(adapterName, func(t *testing.T) {
			fixture := newApprovalFixture(t)
			entry := approvalEntryPoint(t, adapterName, fixture.runner)
			if err := entry.run(context.Background(), 1, approvalPrincipal(false)); err == nil {
				t.Fatal("denied approval succeeded")
			} else {
				assertDeniedResult(t, adapterName, err)
			}
			got := fixture.snapshot(t)
			if got.OrderStatus != "pending" || got.ApprovedBy.Valid || len(got.History) != 0 || len(got.Audit) != 0 {
				t.Fatalf("denied operation changed business rows or success audit: %+v", got)
			}
			var denied int
			if err := fixture.database.QueryRow(`SELECT count(*) FROM forge_audit_events WHERE operation = ? AND outcome = 'denied'`, operations.ApproveOrder.Name).Scan(&denied); err != nil {
				t.Fatal(err)
			}
			if denied != 1 {
				t.Fatalf("denial audit count = %d, want 1", denied)
			}
		})
	}
}

func TestApproveOrderConcurrentApprovals(t *testing.T) {
	fixture := newApprovalFixture(t)
	entry := approvalEntryPoint(t, "rest", fixture.runner)
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			results <- entry.run(context.Background(), 1, approvalPrincipal(true))
		}()
	}
	wait.Wait()
	close(results)
	var successes, conflicts int
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		if errors.Is(err, operation.ErrConflict) {
			var statusErr *httpStatusError
			if errors.As(err, &statusErr) && statusErr.status != http.StatusConflict {
				t.Fatalf("conflict HTTP status = %d, want 409", statusErr.status)
			}
			conflicts++
			continue
		}
		t.Fatalf("unexpected concurrent result: %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes = %d, conflicts = %d; want one each", successes, conflicts)
	}
	got := fixture.snapshot(t)
	if got.OrderStatus != "approved" || len(got.History) != 1 || len(got.Audit) != 1 {
		t.Fatalf("concurrent final state = %+v", got)
	}
}

type approvalFixture struct {
	database *db.DB
	runner   *operation.Runner
}

func newApprovalFixture(t *testing.T) *approvalFixture {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "orders.sqlite"), db.WithMaxOpenConns(1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	for _, query := range []string{
		`CREATE TABLE orders (id INTEGER PRIMARY KEY, status TEXT NOT NULL, approved_by TEXT NULL)`,
		`CREATE TABLE order_history (id INTEGER PRIMARY KEY AUTOINCREMENT, order_id INTEGER NOT NULL, status TEXT NOT NULL, changed_by TEXT NOT NULL)`,
		`INSERT INTO orders (id, status) VALUES (1, 'pending')`,
	} {
		if _, err := database.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	sink := audit.NewSQLSink(database)
	if err := sink.EnsureSchema(context.Background()); err != nil {
		t.Fatal(err)
	}
	return &approvalFixture{database: database, runner: operation.NewRunner(database, sink)}
}

func (f *approvalFixture) snapshot(t *testing.T) approvalSnapshot {
	t.Helper()
	var got approvalSnapshot
	if err := f.database.QueryRow(`SELECT status, approved_by FROM orders WHERE id = 1`).Scan(&got.OrderStatus, &got.ApprovedBy); err != nil {
		t.Fatal(err)
	}
	rows, err := f.database.Query(`SELECT order_id, status, changed_by FROM order_history ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var row historyRow
		if err := rows.Scan(&row.OrderID, &row.Status, &row.ChangedBy); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		got.History = append(got.History, row)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	auditRows, err := f.database.Query(`SELECT operation, actor_id, actor_kind, model, outcome, source FROM forge_audit_events WHERE outcome = 'success' ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	for auditRows.Next() {
		var row auditRow
		if err := auditRows.Scan(&row.Operation, &row.ActorID, &row.ActorKind, &row.Model, &row.Outcome, &row.Source); err != nil {
			auditRows.Close()
			t.Fatal(err)
		}
		got.Audit = append(got.Audit, row)
	}
	if err := auditRows.Close(); err != nil {
		t.Fatal(err)
	}
	if err := auditRows.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

func approvalEntryPoint(t *testing.T, name string, runner *operation.Runner) entryPoint {
	t.Helper()
	switch name {
	case "admin":
		action := admincore.OperationAction[operations.ApproveOrderInput, operations.ApproveOrderResult]{Runner: runner, Operation: operations.ApproveOrder}
		return entryPoint{name: name, run: func(ctx context.Context, id int64, actor principal.Principal) error {
			_, err := action.Execute(apicore.WithPrincipal(ctx, actor), operations.ApproveOrderInput{OrderID: id}, true)
			return err
		}}
	case "rest":
		handler := forgeapi.OperationAction(runner, operations.ApproveOrder, func(r *http.Request) (operation.Actor, error) {
			actor, ok := apicore.PrincipalFromContext(r.Context())
			if !ok {
				return nil, nil
			}
			return principal.Actor(actor), nil
		})
		return entryPoint{name: name, run: func(ctx context.Context, id int64, actor principal.Principal) error {
			body, err := json.Marshal(operations.ApproveOrderInput{OrderID: id})
			if err != nil {
				return err
			}
			req := httptest.NewRequest(http.MethodPost, "/orders/approve", bytes.NewReader(body))
			req = req.WithContext(apicore.WithPrincipal(req.Context(), actor))
			response := httptest.NewRecorder()
			handler(response, req)
			if response.Code != http.StatusOK {
				return &httpStatusError{status: response.Code}
			}
			return nil
		}}
	case "cli":
		return entryPoint{name: name, run: func(ctx context.Context, id int64, actor principal.Principal) error {
			input, err := json.Marshal(operations.ApproveOrderInput{OrderID: id})
			if err != nil {
				return err
			}
			cliContext := core.NewContext()
			cliContext.Principal = &actor
			return (core.OperationCommand[operations.ApproveOrderInput, operations.ApproveOrderResult]{Runner: runner, Operation: operations.ApproveOrder, Input: bytes.NewReader(input)}).Execute(ctx, cliContext)
		}}
	default:
		t.Fatalf("unknown entry point %q", name)
		return entryPoint{}
	}
}

type httpStatusError struct{ status int }

func (e *httpStatusError) Error() string { return "unexpected HTTP status" }
func (e *httpStatusError) Is(target error) bool {
	return e.status == http.StatusConflict && target == operation.ErrConflict
}

func approvalPrincipal(allowed bool) principal.Principal {
	p := principal.Principal{ID: "approver", Kind: principal.KindUser}
	if allowed {
		p.PermissionNames = []string{operations.ApproveOrderPermission}
	}
	return p
}

func assertDeniedResult(t *testing.T, entryPoint string, err error) {
	t.Helper()
	switch entryPoint {
	case "admin":
		var actionErr *admincore.ActionHTTPError
		if !errors.As(err, &actionErr) || actionErr.StatusCode != http.StatusForbidden {
			t.Fatalf("admin error = %v, want HTTP 403", err)
		}
	case "rest":
		var statusErr *httpStatusError
		if !errors.As(err, &statusErr) || statusErr.status != http.StatusForbidden {
			t.Fatalf("REST error = %v, want HTTP 403", err)
		}
	case "cli":
		var commandErr *core.OperationError
		if !errors.As(err, &commandErr) || commandErr.ExitCode != core.OperationExitDenied {
			t.Fatalf("CLI error = %v, want exit code 2", err)
		}
	}
}
