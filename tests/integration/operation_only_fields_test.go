package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/admin/api/rest"
	admincore "github.com/forgego/forge/admin/core"
	forgeapi "github.com/forgego/forge/api"
	apicore "github.com/forgego/forge/api/core"
	"github.com/forgego/forge/audit"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/principal"
	"github.com/forgego/forge/operation"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/forgego/forge/server"
	"github.com/forgego/forge/tests/testdata/reference_app/operations"
	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

type TestOrderModel struct {
	schema.BaseSchema
	ID          int64  `db:"id" json:"id"`
	OrderNumber string `db:"order_number" json:"order_number"`
	Status      string `db:"status" json:"status"`
	ApprovedBy  string `db:"approved_by" json:"approved_by"`
}

func (TestOrderModel) Meta() schema.Meta {
	return schema.Meta{
		TableName:           "orders",
		OperationOnlyFields: []string{"status"},
	}
}

func (TestOrderModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("order_number"),
		schema.StringField("status", schema.OperationOnly()),
		schema.StringField("approved_by"),
	}
}

func TestOperationOnlyFieldsBlocked_GenericPatchRejected(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "orders.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	for _, query := range []string{
		`CREATE TABLE orders (id INTEGER PRIMARY KEY, order_number TEXT NOT NULL, status TEXT NOT NULL, approved_by TEXT NULL)`,
		`CREATE TABLE order_history (id INTEGER PRIMARY KEY AUTOINCREMENT, order_id INTEGER NOT NULL, status TEXT NOT NULL, changed_by TEXT NOT NULL)`,
		`INSERT INTO orders (id, order_number, status) VALUES (1, 'ORD-001', 'pending')`,
	} {
		_, err = database.Exec(query)
		require.NoError(t, err)
	}

	sink := audit.NewSQLSink(database)
	require.NoError(t, sink.EnsureSchema(context.Background()))
	runner := operation.NewRunner(database, sink)

	manager, err := orm.NewManagerWithDB[TestOrderModel]("orders", database)
	require.NoError(t, err)

	adminConfig := &admincore.Config[TestOrderModel]{
		OperationOnlyFields: []string{"status"},
		HasChangePermission: func(ctx context.Context, a *admincore.Admin[TestOrderModel], user interface{}, obj *TestOrderModel) bool {
			return true
		},
	}
	adminInstance, err := admincore.NewAdmin[TestOrderModel](TestOrderModel{}, manager, adminConfig)
	require.NoError(t, err)

	registry := admincore.NewRegistry()
	require.NoError(t, registry.Register(adminInstance))

	adminRouter := rest.NewRouter(registry)
	chiAdmin := chi.NewRouter()
	adminRouter.RegisterRoutes(chiAdmin)

	t.Run("admin REST generic PATCH rejected and row unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/orders/1", strings.NewReader(`{"status": "approved"}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(apicore.WithUser(req.Context(), map[string]interface{}{"role": "superuser"}))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		adminRouter.UpdateHandler(adminInstance).ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code, "PATCH of status must return HTTP 403 Forbidden: %s", rec.Body.String())
		var errResp struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
			Code string `json:"code"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		code := errResp.Error.Code
		if code == "" {
			code = errResp.Code
		}
		require.Equal(t, "operation_only_field", code)

		var currentStatus string
		err := database.QueryRow(`SELECT status FROM orders WHERE id = 1`).Scan(&currentStatus)
		require.NoError(t, err)
		require.Equal(t, "pending", currentStatus, "database row must be unchanged")
	})

	t.Run("admin REST generic PUT rejected and row unchanged", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/orders/1", strings.NewReader(`{"status": "approved", "order_number": "ORD-001"}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(apicore.WithUser(req.Context(), map[string]interface{}{"role": "superuser"}))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		adminRouter.ReplaceHandler(adminInstance).ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		var currentStatus string
		err := database.QueryRow(`SELECT status FROM orders WHERE id = 1`).Scan(&currentStatus)
		require.NoError(t, err)
		require.Equal(t, "pending", currentStatus)
	})

	t.Run("admin REST generic update of allowed field succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/orders/1", strings.NewReader(`{"order_number": "ORD-ADMIN"}`))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(apicore.WithUser(req.Context(), map[string]interface{}{"role": "superuser"}))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		adminRouter.UpdateHandler(adminInstance).ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "allowed field update must succeed: %s", rec.Body.String())
		var num string
		err := database.QueryRow(`SELECT order_number FROM orders WHERE id = 1`).Scan(&num)
		require.NoError(t, err)
		require.Equal(t, "ORD-ADMIN", num)
	})

	t.Run("viewset generic PATCH rejected and row unchanged", func(t *testing.T) {
		vs := forgeapi.NewBaseViewSet(
			func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
			manager,
			&TestOrderModel{},
		)
		vs.OperationOnlyFields = []string{"status"}
		vs.Authentication = nil
		vs.Permissions = nil

		apiRouter := forgeapi.NewRouter("/api/v1")
		apiRouter.Register("orders", vs)
		mux := server.NewRouter()
		apiRouter.RegisterRoutes(mux)

		// Test case-insensitive key "Status"
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/1", bytes.NewBufferString(`{"Status": "approved"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code, "PATCH of status must return HTTP 403 Forbidden: %s", rec.Body.String())

		var errResp struct {
			Code string `json:"code"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
		require.Equal(t, "operation_only_field", errResp.Code)

		var currentStatus string
		err := database.QueryRow(`SELECT status FROM orders WHERE id = 1`).Scan(&currentStatus)
		require.NoError(t, err)
		require.Equal(t, "pending", currentStatus, "database row must be unchanged")
	})

	t.Run("viewset generic PUT rejected and row unchanged", func(t *testing.T) {
		vs := forgeapi.NewBaseViewSet(
			func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
			manager,
			&TestOrderModel{},
		)
		vs.OperationOnlyFields = []string{"status"}
		vs.Authentication = nil
		vs.Permissions = nil

		apiRouter := forgeapi.NewRouter("/api/v1")
		apiRouter.Register("orders", vs)
		mux := server.NewRouter()
		apiRouter.RegisterRoutes(mux)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/orders/1", bytes.NewBufferString(`{"status": "approved", "order_number": "ORD-001"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		var currentStatus string
		err := database.QueryRow(`SELECT status FROM orders WHERE id = 1`).Scan(&currentStatus)
		require.NoError(t, err)
		require.Equal(t, "pending", currentStatus)
	})

	t.Run("viewset generic update of allowed field succeeds", func(t *testing.T) {
		vs := forgeapi.NewBaseViewSet(
			func() forgeapi.Serializer { return forgeapi.NewBaseSerializer(nil) },
			manager,
			&TestOrderModel{},
		)
		vs.OperationOnlyFields = []string{"status"}
		vs.Authentication = nil
		vs.Permissions = nil

		apiRouter := forgeapi.NewRouter("/api/v1")
		apiRouter.Register("orders", vs)
		mux := server.NewRouter()
		apiRouter.RegisterRoutes(mux)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/1", bytes.NewBufferString(`{"order_number": "ORD-VIEWSET"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, "allowed field update must succeed: %s", rec.Body.String())
		var num string
		err := database.QueryRow(`SELECT order_number FROM orders WHERE id = 1`).Scan(&num)
		require.NoError(t, err)
		require.Equal(t, "ORD-VIEWSET", num)
	})

	t.Run("authorized operation changes status", func(t *testing.T) {
		p := principal.Principal{ID: "approver", Kind: principal.KindUser, PermissionNames: []string{operations.ApproveOrderPermission}}
		action := admincore.OperationAction[operations.ApproveOrderInput, operations.ApproveOrderResult]{
			Runner:    runner,
			Operation: operations.ApproveOrder,
		}
		res, err := action.Execute(apicore.WithPrincipal(context.Background(), p), operations.ApproveOrderInput{OrderID: 1}, true)
		require.NoError(t, err)
		require.Equal(t, "approved", res.Status)

		var currentStatus string
		err = database.QueryRow(`SELECT status FROM orders WHERE id = 1`).Scan(&currentStatus)
		require.NoError(t, err)
		require.Equal(t, "approved", currentStatus, "operation must be able to change status")
	})
}
