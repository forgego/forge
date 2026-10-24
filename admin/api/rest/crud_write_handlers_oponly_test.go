package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opOnlyMockAdmin struct {
	mockAdmin
	opFields []string
}

func (m *opOnlyMockAdmin) OperationOnlyFields() []string {
	return m.opFields
}

func (m *opOnlyMockAdmin) ModelType() reflect.Type {
	return reflect.TypeOf(struct{}{})
}

func TestCrudWriteHandlers_BlockOperationOnlyFields(t *testing.T) {
	admin := &opOnlyMockAdmin{
		mockAdmin: mockAdmin{
			modelName: "orders",
			getObjectFn: func(id interface{}) (interface{}, error) {
				return map[string]interface{}{"id": id, "status": "pending"}, nil
			},
			hasChangePermission: func(obj interface{}) bool {
				return true
			},
		},
		opFields: []string{"status"},
	}

	router := NewRouter(core.NewRegistry())

	t.Run("HandleUpdate blocks operation-only field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/orders/1", bytes.NewBufferString(`{"status": "approved"}`))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()

		router.handleUpdate(admin)(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		errObj := body["error"].(map[string]interface{})
		assert.Equal(t, "operation_only_field", errObj["code"])
	})

	t.Run("HandleReplace blocks operation-only field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/orders/1", bytes.NewBufferString(`{"status": "approved", "note": "all good"}`))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "1")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()

		router.handleReplace(admin)(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		errObj := body["error"].(map[string]interface{})
		assert.Equal(t, "operation_only_field", errObj["code"])
	})

	t.Run("handleBulkUpdate blocks operation-only field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/orders/bulk-update", bytes.NewBufferString(`{"ids": [1], "data": {"status": "approved"}}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.handleBulkUpdate(admin)(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		errObj := body["error"].(map[string]interface{})
		assert.Equal(t, "operation_only_field", errObj["code"])
	})
}
