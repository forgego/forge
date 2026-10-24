package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	admincore "github.com/forgego/forge/admin/core"
	forgeapi "github.com/forgego/forge/api"
	apicore "github.com/forgego/forge/api/core"
	apierrors "github.com/forgego/forge/api/errors"
	"github.com/forgego/forge/api/exceptions"
	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/config"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/log"
	"github.com/forgego/forge/operation"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/stores"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type correlationItemInput struct {
	Name string `json:"name"`
}

type correlationItemResult struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type correlationViewSet struct {
	sink audit.Sink
}

func (v *correlationViewSet) List(w http.ResponseWriter, r *http.Request) {}

func (v *correlationViewSet) Create(w http.ResponseWriter, r *http.Request) {
	var payload correlationItemInput
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		apierrors.WriteError(w, r, exceptions.NewValidationError(map[string][]string{"body": {"invalid JSON"}}))
		return
	}
	if payload.Name == "" {
		apierrors.WriteError(w, r, exceptions.NewValidationError(map[string][]string{"name": {"name is required"}}))
		return
	}

	event := audit.Event{
		ActorID:   "tester",
		ActorKind: "user",
		Operation: "item.create",
		Outcome:   audit.OutcomeSuccess,
	}
	if err := audit.Record(r.Context(), v.sink, nil, event, audit.FailClosed, audit.Options{}); err != nil {
		apierrors.WriteError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(correlationItemResult{ID: 1, Name: payload.Name})
}

func (v *correlationViewSet) Retrieve(w http.ResponseWriter, r *http.Request)      {}
func (v *correlationViewSet) Update(w http.ResponseWriter, r *http.Request)        {}
func (v *correlationViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {}
func (v *correlationViewSet) Destroy(w http.ResponseWriter, r *http.Request)       {}

func TestCorrelationIDReachesErrorResponsesAndAuditEvents(t *testing.T) {
	ctx := context.Background()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "correlation.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	require.NoError(t, stores.Migrate(ctx, database))

	sink := audit.NewSQLSink(database)
	require.NoError(t, sink.EnsureSchema(ctx))

	runner := operation.NewRunner(database, sink)

	const permName = "item.write"
	testPrincipal := principal.Principal{
		ID:              "tester",
		Kind:            principal.KindUser,
		PermissionNames: []string{permName},
	}

	op := operation.Operation[correlationItemInput, correlationItemResult]{
		Name:       "api.write",
		Permission: permName,
		Run: func(ctx context.Context, tx operation.Transaction, _ operation.Actor, input correlationItemInput) (correlationItemResult, []operation.Effect, error) {
			return correlationItemResult{ID: 42, Name: input.Name}, nil, nil
		},
	}

	adminOp := operation.Operation[correlationItemInput, correlationItemResult]{
		Name:       "admin.write",
		Permission: permName,
		Run: func(ctx context.Context, tx operation.Transaction, _ operation.Actor, input correlationItemInput) (correlationItemResult, []operation.Effect, error) {
			return correlationItemResult{ID: 99, Name: input.Name}, nil, nil
		},
	}

	opAction := forgeapi.OperationAction(runner, op, func(r *http.Request) (operation.Actor, error) {
		return principal.Actor(testPrincipal), nil
	})

	adminAction := admincore.OperationAction[correlationItemInput, correlationItemResult]{
		Runner:    runner,
		Operation: adminOp,
	}

	settings := &config.Settings{
		App:      config.AppSettings{Name: "correlation-test", Env: "test"},
		Server:   config.ServerSettings{Host: "127.0.0.1", Port: "0"},
		Security: config.SecuritySettings{SessionSecret: "correlation-test-secret-0123456789"},
		Admin:    config.AdminSettings{Enabled: true, Path: "/admin"},
	}

	srv, err := server.NewServer(config.NewConfig(), settings, log.NewNopLogger(), server.WithDatabase(database))
	require.NoError(t, err)

	apiRouter := forgeapi.NewRouter("/api")
	apiRouter.Register("items", &correlationViewSet{sink: sink})

	srv.RegisterRoutes(func(r *server.Router) {
		apiRouter.RegisterRoutes(r)
		r.Post("/api/operations/create", opAction)
		r.Post("/admin/operations/create", func(w http.ResponseWriter, req *http.Request) {
			var input correlationItemInput
			_ = json.NewDecoder(req.Body).Decode(&input)
			res, err := adminAction.Execute(apicore.WithPrincipal(req.Context(), testPrincipal), input, true)
			if err != nil {
				server.SendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			_ = server.SendJSON(w, http.StatusOK, res)
		})
	})

	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(ts.Close)

	doRequest := func(t *testing.T, method, url, requestID string, body any) (*http.Response, []byte) {
		t.Helper()
		var reader io.Reader
		if body != nil {
			raw, err := json.Marshal(body)
			require.NoError(t, err)
			reader = bytes.NewReader(raw)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if requestID != "" {
			req.Header.Set("X-Request-ID", requestID)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		_ = resp.Body.Close()
		return resp, respBody
	}

	const testCorrID = "test-corr-123"

	t.Run("with explicit correlation id", func(t *testing.T) {
		// (1) Validation error on an API viewset
		resp, respBody := doRequest(t, http.MethodPost, ts.URL+"/api/items/", testCorrID, map[string]string{"name": ""})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, testCorrID, resp.Header.Get("X-Request-ID"), "header must match X-Request-ID")

		var errBody map[string]interface{}
		require.NoError(t, json.Unmarshal(respBody, &errBody))
		assert.Equal(t, testCorrID, errBody["request_id"], "error body field request_id must equal X-Request-ID")

		// (2) Audited API write (via ViewSet Create)
		respWrite, _ := doRequest(t, http.MethodPost, ts.URL+"/api/items/", testCorrID, map[string]string{"name": "valid-item"})
		assert.Equal(t, http.StatusCreated, respWrite.StatusCode)
		assert.Equal(t, testCorrID, respWrite.Header.Get("X-Request-ID"), "header must match X-Request-ID")

		var corrID string
		err := database.QueryRowContext(ctx, "SELECT correlation_id FROM forge_audit_events WHERE operation = ? ORDER BY id DESC LIMIT 1", "item.create").Scan(&corrID)
		require.NoError(t, err)
		assert.Equal(t, testCorrID, corrID, "audit event correlation_id must equal X-Request-ID")

		// Also verify audited write via operation action
		respOp, _ := doRequest(t, http.MethodPost, ts.URL+"/api/operations/create", testCorrID, map[string]string{"name": "op-item"})
		assert.Equal(t, http.StatusOK, respOp.StatusCode)
		assert.Equal(t, testCorrID, respOp.Header.Get("X-Request-ID"))

		err = database.QueryRowContext(ctx, "SELECT correlation_id FROM forge_audit_events WHERE operation = ? ORDER BY id DESC LIMIT 1", "api.write").Scan(&corrID)
		require.NoError(t, err)
		assert.Equal(t, testCorrID, corrID, "operation audit event correlation_id must equal X-Request-ID")

		// Also verify audited admin write
		respAdmin, _ := doRequest(t, http.MethodPost, ts.URL+"/admin/operations/create", testCorrID, map[string]string{"name": "admin-item"})
		assert.Equal(t, http.StatusOK, respAdmin.StatusCode)
		assert.Equal(t, testCorrID, respAdmin.Header.Get("X-Request-ID"))

		err = database.QueryRowContext(ctx, "SELECT correlation_id FROM forge_audit_events WHERE operation = ? ORDER BY id DESC LIMIT 1", "admin.write").Scan(&corrID)
		require.NoError(t, err)
		assert.Equal(t, testCorrID, corrID, "admin audit event correlation_id must equal X-Request-ID")
	})

	t.Run("with generated correlation id when none is sent", func(t *testing.T) {
		// (1) Validation error on an API viewset
		resp, respBody := doRequest(t, http.MethodPost, ts.URL+"/api/items/", "", map[string]string{"name": ""})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		genID := resp.Header.Get("X-Request-ID")
		assert.NotEmpty(t, genID, "generated request id must be set on response header")

		var errBody map[string]interface{}
		require.NoError(t, json.Unmarshal(respBody, &errBody))
		assert.Equal(t, genID, errBody["request_id"], "error body field request_id must equal generated request ID")

		// (2) Audited API write
		respWrite, _ := doRequest(t, http.MethodPost, ts.URL+"/api/items/", "", map[string]string{"name": "auto-item"})
		assert.Equal(t, http.StatusCreated, respWrite.StatusCode)
		genWriteID := respWrite.Header.Get("X-Request-ID")
		assert.NotEmpty(t, genWriteID, "generated request id must be set on response header")

		var corrID string
		err := database.QueryRowContext(ctx, "SELECT correlation_id FROM forge_audit_events WHERE operation = ? ORDER BY id DESC LIMIT 1", "item.create").Scan(&corrID)
		require.NoError(t, err)
		assert.Equal(t, genWriteID, corrID, "audit event correlation_id must equal generated request ID")

		// Audited admin write
		respAdmin, _ := doRequest(t, http.MethodPost, ts.URL+"/admin/operations/create", "", map[string]string{"name": "auto-admin-item"})
		assert.Equal(t, http.StatusOK, respAdmin.StatusCode)
		genAdminID := respAdmin.Header.Get("X-Request-ID")
		assert.NotEmpty(t, genAdminID)

		err = database.QueryRowContext(ctx, "SELECT correlation_id FROM forge_audit_events WHERE operation = ? ORDER BY id DESC LIMIT 1", "admin.write").Scan(&corrID)
		require.NoError(t, err)
		assert.Equal(t, genAdminID, corrID)
	})
}
