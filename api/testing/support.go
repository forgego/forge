package testing

import (
	"context"
	"net/http"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/server"
)

// App is an isolated application test harness wrapping a server.Router,
// an optional database connection, and supporting test client creation.
type App struct {
	Router  *server.Router
	DB      *db.DB
	handler http.Handler
}

// AppOption configures an App instance during construction.
type AppOption func(*App)

// WithRoutes configures route registration on the application's router.
func WithRoutes(fn func(*server.Router)) AppOption {
	return func(a *App) {
		if fn != nil && a.Router != nil {
			fn(a.Router)
		}
	}
}

// WithDB configures the database connection used by the test application.
// The caller owns the connection.
func WithDB(database *db.DB) AppOption {
	return func(a *App) {
		a.DB = database
	}
}

// WithHandler sets an explicit http.Handler for the application.
func WithHandler(handler http.Handler) AppOption {
	return func(a *App) {
		a.handler = handler
	}
}

// Handler returns the underlying http.Handler for the application.
func (a *App) Handler() http.Handler {
	if a.handler != nil {
		return a.handler
	}
	return a.Router
}

// ServeHTTP implements http.Handler for App.
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.Handler().ServeHTTP(w, req)
}

// Client returns an *APIClient bound to the application handler.
func (a *App) Client() *APIClient {
	return NewAPIClient(a.Handler())
}

// NewApp creates an isolated application builder for testing.
// It initializes a server.Router and applies the provided options.
// The t parameter is reserved for future cleanup registration.
func NewApp(t testing.TB, opts ...AppOption) *App {
	if t != nil {
		t.Helper()
	}
	app := &App{
		Router: server.NewRouter(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}
	return app
}

// AsPrincipal returns a copy of the APIClient whose requests carry
// the given principal in their context via api/core.WithPrincipal.
// Unauthenticated requests on the original client remain unaffected.
func (c *APIClient) AsPrincipal(p principal.Principal) *APIClient {
	if c == nil {
		return nil
	}
	headers := make(map[string]string, len(c.Headers))
	for k, v := range c.Headers {
		headers[k] = v
	}
	var handler http.Handler
	if c.Handler != nil {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := core.WithPrincipal(r.Context(), p)
			c.Handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	return &APIClient{
		Handler: handler,
		BaseURL: c.BaseURL,
		Headers: headers,
	}
}

// WithRollback begins a database transaction and registers a cleanup
// function on t to roll it back when the test finishes.
func WithRollback(t testing.TB, database *db.DB) *db.Tx {
	t.Helper()
	if database == nil {
		t.Fatalf("WithRollback: database is nil")
		return nil
	}
	tx, err := database.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("WithRollback: failed to begin transaction: %v", err)
		return nil
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
	})
	return tx
}

// AssertAllowed asserts that the response has a 2xx status code.
func AssertAllowed(t testing.TB, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatalf("AssertAllowed: expected response, got nil")
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Fatalf("AssertAllowed: expected 2xx status, got %d (body: %s)", resp.StatusCode, string(resp.Body))
	}
}

// AssertForbidden asserts that the response has a 403 Forbidden status code.
func AssertForbidden(t testing.TB, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatalf("AssertForbidden: expected response, got nil")
		return
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("AssertForbidden: expected 403 Forbidden, got %d (body: %s)", resp.StatusCode, string(resp.Body))
	}
}

// AssertUnauthorized asserts that the response has a 401 Unauthorized status code.
func AssertUnauthorized(t testing.TB, resp *Response) {
	t.Helper()
	if resp == nil {
		t.Fatalf("AssertUnauthorized: expected response, got nil")
		return
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("AssertUnauthorized: expected 401 Unauthorized, got %d (body: %s)", resp.StatusCode, string(resp.Body))
	}
}
