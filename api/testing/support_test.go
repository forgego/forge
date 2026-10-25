package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/server"
	_ "github.com/mattn/go-sqlite3"
)

type fakeTB struct {
	testing.TB
	failed   bool
	messages []string
}

func newFakeTB(t *testing.T) *fakeTB {
	return &fakeTB{TB: t}
}

func (f *fakeTB) Helper() {}

func (f *fakeTB) Errorf(format string, args ...any) {
	f.failed = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Fatalf(format string, args ...any) {
	f.failed = true
	f.messages = append(f.messages, fmt.Sprintf(format, args...))
}

func (f *fakeTB) Fail() {
	f.failed = true
}

func (f *fakeTB) FailNow() {
	f.failed = true
}

func (f *fakeTB) Failed() bool {
	return f.failed
}

func TestNewApp(t *testing.T) {
	t.Run("routes and client get 200", func(t *testing.T) {
		app := NewApp(t, WithRoutes(func(r *server.Router) {
			r.Get("/ping", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("pong"))
			})
		}))

		client := app.Client()
		resp := client.Get("/ping")
		AssertAllowed(t, resp)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
		if string(resp.Body) != "pong" {
			t.Fatalf("expected body pong, got %s", string(resp.Body))
		}
	})

	t.Run("with db option preserves database after subtest", func(t *testing.T) {
		database, err := db.NewDBWithDriver("sqlite3", ":memory:", db.WithMaxOpenConns(1))
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}

		t.Run("creates app", func(subT *testing.T) {
			app := NewApp(subT, WithDB(database))
			if app.DB != database {
				subT.Fatalf("expected app.DB to match configured database")
			}
		})

		if err := database.Ping(context.Background()); err != nil {
			t.Fatalf("expected database ping to succeed after subtest finished, got error: %v", err)
		}
	})

	t.Run("with explicit handler", func(t *testing.T) {
		customHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("custom"))
		})

		app := NewApp(t, WithHandler(customHandler))
		resp := app.Client().Get("/custom")
		AssertAllowed(t, resp)
		if string(resp.Body) != "custom" {
			t.Fatalf("expected body custom, got %s", string(resp.Body))
		}
	})
}

func TestAsPrincipal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := core.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("anonymous"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(p.ID))
	})

	client := NewAPIClient(handler)

	// Unauthenticated request: core.PrincipalFromContext reports ok = false
	unauthResp := client.Get("/whoami")
	AssertUnauthorized(t, unauthResp)
	if string(unauthResp.Body) != "anonymous" {
		t.Fatalf("expected anonymous response body, got %s", string(unauthResp.Body))
	}

	// Authenticated request with AsPrincipal: echoes principal ID
	userPrincipal := principal.Principal{
		ID:   "user-987",
		Kind: principal.KindUser,
	}
	authClient := client.AsPrincipal(userPrincipal)
	authResp := authClient.Get("/whoami")
	AssertAllowed(t, authResp)
	if string(authResp.Body) != "user-987" {
		t.Fatalf("expected user-987 response body, got %s", string(authResp.Body))
	}

	// Verify original client remains unauthenticated
	againResp := client.Get("/whoami")
	AssertUnauthorized(t, againResp)
	if string(againResp.Body) != "anonymous" {
		t.Fatalf("expected original client to stay unauthenticated, got %s", string(againResp.Body))
	}

	// Corner case: AsPrincipal on nil client
	var nilClient *APIClient
	if nilClient.AsPrincipal(userPrincipal) != nil {
		t.Fatalf("expected nil client AsPrincipal to return nil")
	}
}

func TestAssertHelpers(t *testing.T) {
	t.Run("AssertAllowed", func(t *testing.T) {
		// Valid 2xx statuses should pass
		for _, status := range []int{200, 201, 202, 204, 299} {
			fake := newFakeTB(t)
			AssertAllowed(fake, &Response{StatusCode: status})
			if fake.failed {
				t.Fatalf("expected status %d to pass AssertAllowed, but it failed", status)
			}
		}

		// Non-2xx statuses should fail
		for _, status := range []int{100, 199, 301, 400, 401, 403, 404, 500} {
			fake := newFakeTB(t)
			AssertAllowed(fake, &Response{StatusCode: status})
			if !fake.failed {
				t.Fatalf("expected status %d to fail AssertAllowed, but it passed", status)
			}
		}

		// Nil response should fail
		fake := newFakeTB(t)
		AssertAllowed(fake, nil)
		if !fake.failed {
			t.Fatalf("expected nil response to fail AssertAllowed, but it passed")
		}
	})

	t.Run("AssertForbidden", func(t *testing.T) {
		// 403 should pass
		fake := newFakeTB(t)
		AssertForbidden(fake, &Response{StatusCode: http.StatusForbidden})
		if fake.failed {
			t.Fatalf("expected 403 to pass AssertForbidden, but it failed")
		}

		// Other statuses should fail
		for _, status := range []int{200, 201, 400, 401, 404, 500} {
			fake := newFakeTB(t)
			AssertForbidden(fake, &Response{StatusCode: status})
			if !fake.failed {
				t.Fatalf("expected status %d to fail AssertForbidden, but it passed", status)
			}
		}

		// Nil response should fail
		fake = newFakeTB(t)
		AssertForbidden(fake, nil)
		if !fake.failed {
			t.Fatalf("expected nil response to fail AssertForbidden, but it passed")
		}
	})

	t.Run("AssertUnauthorized", func(t *testing.T) {
		// 401 should pass
		fake := newFakeTB(t)
		AssertUnauthorized(fake, &Response{StatusCode: http.StatusUnauthorized})
		if fake.failed {
			t.Fatalf("expected 401 to pass AssertUnauthorized, but it failed")
		}

		// Other statuses should fail
		for _, status := range []int{200, 201, 400, 403, 404, 500} {
			fake := newFakeTB(t)
			AssertUnauthorized(fake, &Response{StatusCode: status})
			if !fake.failed {
				t.Fatalf("expected status %d to fail AssertUnauthorized, but it passed", status)
			}
		}

		// Nil response should fail
		fake = newFakeTB(t)
		AssertUnauthorized(fake, nil)
		if !fake.failed {
			t.Fatalf("expected nil response to fail AssertUnauthorized, but it passed")
		}
	})
}

func TestWithRollback(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	database := &db.DB{
		DB:     sqlDB,
		Driver: "sqlite3",
	}

	_, err = database.Exec("CREATE TABLE test_rollback (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	t.Run("insert inside rollback tx", func(subT *testing.T) {
		tx := WithRollback(subT, database)
		if tx == nil {
			subT.Fatal("expected non-nil tx")
		}

		_, err := tx.Exec("INSERT INTO test_rollback (id, name) VALUES (1, 'item-1')")
		if err != nil {
			subT.Fatalf("insert failed: %v", err)
		}

		var count int
		if err := tx.QueryRow("SELECT COUNT(*) FROM test_rollback WHERE id = 1").Scan(&count); err != nil {
			subT.Fatalf("query inside tx failed: %v", err)
		}
		if count != 1 {
			subT.Fatalf("expected count 1 inside tx, got %d", count)
		}
	})

	// After subtest cleanup has run, tx has rolled back: the row must be gone
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM test_rollback WHERE id = 1").Scan(&count); err != nil {
		t.Fatalf("parent query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after subtest cleanup rollback, got %d", count)
	}
}
