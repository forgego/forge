package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAdmin struct {
	modelName             string
	moduleAllowed         bool
	metadata              *core.Metadata
	metadataErr           error
	listResponse          *core.PaginatedResponse
	listErr               error
	createObjectFn        func(data map[string]interface{}) (interface{}, error)
	createCalls           []map[string]interface{}
	hasChangePermission   func(obj interface{}) bool
	hasDeletePermission   func(obj interface{}) bool
	getObjectFn           func(id interface{}) (interface{}, error)
	getObjectResponse     interface{}
	getObjectErr          error
	updateObjectFn        func(id interface{}, data map[string]interface{}) (interface{}, error)
	updateObjectResponse  interface{}
	updateObjectErr       error
	lastUpdateID          interface{}
	lastUpdateData        map[string]interface{}
	updateCalls           []mockUpdateCall
	deleteObjectFn        func(id interface{}) error
	deleteCalls           []interface{}
	executeActionFn       func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error)
	lastActionName        string
	lastActionIDs         []interface{}
	lastActionParams      map[string]interface{}
	historyEntries        []core.LogEntry
	historyErr            error
	denyView              bool
	autocompleteItems     []core.AutocompleteItem
	autocompleteErr       error
	lastAutocompleteQuery string
	lastAutocompleteLimit int
}

type mockUpdateCall struct {
	ID   interface{}
	Data map[string]interface{}
}

func (m *mockAdmin) SetDB(database *db.DB) {}
func (m *mockAdmin) ModelName() string     { return m.modelName }
func (m *mockAdmin) ModelType() reflect.Type {
	return reflect.TypeOf(struct{}{})
}
func (m *mockAdmin) GetMetadata(ctx context.Context, user interface{}) (*core.Metadata, error) {
	if m.metadataErr != nil {
		return nil, m.metadataErr
	}
	return m.metadata, nil
}
func (m *mockAdmin) HasAddPermission(ctx context.Context, user interface{}) bool {
	return true
}
func (m *mockAdmin) HasChangePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	if m.hasChangePermission != nil {
		return m.hasChangePermission(obj)
	}
	return true
}
func (m *mockAdmin) HasDeletePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	if m.hasDeletePermission != nil {
		return m.hasDeletePermission(obj)
	}
	return true
}
func (m *mockAdmin) HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return !m.denyView
}
func (m *mockAdmin) HasModulePermission(ctx context.Context, user interface{}) bool {
	return m.moduleAllowed
}
func (m *mockAdmin) ManagerInterface() interface{} { return nil }
func (m *mockAdmin) ConfigInterface() interface{}  { return nil }
func (m *mockAdmin) GetHistory(ctx context.Context, objectID string) ([]core.LogEntry, error) {
	if m.historyErr != nil {
		return nil, m.historyErr
	}
	return m.historyEntries, nil
}
func (m *mockAdmin) LogAction(ctx context.Context, user interface{}, objectID string, repr string, action core.ActionType, changes string) error {
	return nil
}
func (m *mockAdmin) PageType() string { return "list" }
func (m *mockAdmin) ListObjects(ctx context.Context, params core.ListParams) (*core.PaginatedResponse, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if m.listResponse == nil {
		return &core.PaginatedResponse{}, nil
	}
	return m.listResponse, nil
}
func (m *mockAdmin) GetObject(ctx context.Context, id interface{}) (interface{}, error) {
	if m.getObjectFn != nil {
		return m.getObjectFn(id)
	}
	if m.getObjectErr != nil {
		return nil, m.getObjectErr
	}
	return m.getObjectResponse, nil
}
func (m *mockAdmin) CreateObject(ctx context.Context, data map[string]interface{}) (interface{}, error) {
	m.createCalls = append(m.createCalls, data)
	if m.createObjectFn != nil {
		return m.createObjectFn(data)
	}
	return data, nil
}
func (m *mockAdmin) UpdateObject(ctx context.Context, id interface{}, data map[string]interface{}) (interface{}, error) {
	m.updateCalls = append(m.updateCalls, mockUpdateCall{ID: id, Data: data})
	m.lastUpdateID = id
	m.lastUpdateData = data
	if m.updateObjectFn != nil {
		return m.updateObjectFn(id, data)
	}
	if m.updateObjectErr != nil {
		return nil, m.updateObjectErr
	}
	return m.updateObjectResponse, nil
}
func (m *mockAdmin) DeleteObject(ctx context.Context, id interface{}) error {
	m.deleteCalls = append(m.deleteCalls, id)
	if m.deleteObjectFn != nil {
		return m.deleteObjectFn(id)
	}
	return nil
}
func (m *mockAdmin) ExecuteAction(ctx context.Context, actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
	m.lastActionName = actionName
	m.lastActionIDs = append([]interface{}(nil), ids...)
	m.lastActionParams = params
	if m.executeActionFn != nil {
		return m.executeActionFn(actionName, ids, params)
	}
	return nil, nil
}
func (m *mockAdmin) Autocomplete(ctx context.Context, query string, limit int) ([]core.AutocompleteItem, error) {
	m.lastAutocompleteQuery = query
	m.lastAutocompleteLimit = limit
	if m.autocompleteErr != nil {
		return nil, m.autocompleteErr
	}
	return m.autocompleteItems, nil
}

func TestHandleMetaList_UsesModelCountFromListObjects(t *testing.T) {
	registry := core.NewRegistry()
	err := registry.Register(&mockAdmin{
		modelName:     "products",
		moduleAllowed: true,
		metadata: &core.Metadata{
			VerboseName:       "Product",
			VerboseNamePlural: "Products",
			Permissions:       core.PermissionMetadata{View: true},
		},
		listResponse: &core.PaginatedResponse{Count: 42},
	})
	require.NoError(t, err)

	router := NewRouter(registry)
	req := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	rec := httptest.NewRecorder()

	router.handleMetaList(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload struct {
		Models []core.ModelListMetadata `json:"models"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Models, 1)
	assert.Equal(t, int64(42), payload.Models[0].Count)
	assert.Equal(t, "products", payload.Models[0].Name)
}

func TestHandleMetaList_FallsBackToZeroWhenListFails(t *testing.T) {
	registry := core.NewRegistry()
	err := registry.Register(&mockAdmin{
		modelName:     "orders",
		moduleAllowed: true,
		metadata: &core.Metadata{
			VerboseName:       "Order",
			VerboseNamePlural: "Orders",
		},
		listErr: assert.AnError,
	})
	require.NoError(t, err)

	router := NewRouter(registry)
	req := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	rec := httptest.NewRecorder()

	router.handleMetaList(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload struct {
		Models []core.ModelListMetadata `json:"models"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Models, 1)
	assert.Equal(t, int64(0), payload.Models[0].Count)
}

func TestHandleConfig_UsesEnvironmentVariable(t *testing.T) {
	t.Setenv("FORGE_ENV", "staging")
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "")
	t.Setenv("ENV", "")

	router := NewRouter(core.NewRegistry())
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	router.handleConfig(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, "staging", payload["environment"])
}

func TestRegisterRoutes_RequiresAuthenticationForProtectedEndpoints(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	registry := core.NewRegistry()
	err := registry.Register(&mockAdmin{
		modelName:     "products",
		moduleAllowed: true,
		metadata: &core.Metadata{
			VerboseName:       "Product",
			VerboseNamePlural: "Products",
			Permissions:       core.PermissionMetadata{View: true},
		},
	})
	require.NoError(t, err)

	apiRouter := NewRouter(registry)
	root := chi.NewRouter()
	apiRouter.RegisterRoutes(root)

	unauthorizedReq := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	unauthorizedRec := httptest.NewRecorder()
	root.ServeHTTP(unauthorizedRec, unauthorizedReq)
	require.Equal(t, http.StatusUnauthorized, unauthorizedRec.Code)

	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginRec := httptest.NewRecorder()
	root.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginPayload))
	token, ok := loginPayload["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	authorizedReq := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	authorizedReq.Header.Set("Authorization", "Bearer "+token)
	authorizedRec := httptest.NewRecorder()
	root.ServeHTTP(authorizedRec, authorizedReq)
	require.Equal(t, http.StatusOK, authorizedRec.Code)
}

func TestRegisterRoutes_RejectsInvalidOrMalformedTokens(t *testing.T) {
	registry := core.NewRegistry()
	err := registry.Register(&mockAdmin{
		modelName:     "products",
		moduleAllowed: true,
		metadata: &core.Metadata{
			VerboseName:       "Product",
			VerboseNamePlural: "Products",
			Permissions:       core.PermissionMetadata{View: true},
		},
	})
	require.NoError(t, err)

	apiRouter := NewRouter(registry)
	root := chi.NewRouter()
	apiRouter.RegisterRoutes(root)

	malformedReq := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	malformedReq.Header.Set("Authorization", "Basic abc123")
	malformedRec := httptest.NewRecorder()
	root.ServeHTTP(malformedRec, malformedReq)
	require.Equal(t, http.StatusUnauthorized, malformedRec.Code)

	invalidReq := httptest.NewRequest(http.MethodGet, "/api/meta", nil)
	invalidReq.Header.Set("Authorization", "Bearer invalid-token")
	invalidRec := httptest.NewRecorder()
	root.ServeHTTP(invalidRec, invalidReq)
	require.Equal(t, http.StatusUnauthorized, invalidRec.Code)
}

func TestHandleLogin_DisabledWhenUnset(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "")
	t.Setenv("FORGE_ADMIN_PASSWORD", "")

	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	errPayload, ok := payload["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "admin_login_disabled", errPayload["code"])
}

func TestHandleLogin_RejectsInvalidCredentials(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	errPayload, ok := payload["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "invalid_credentials", errPayload["code"])
}

func TestHandleLogin_UsesConfiguredCredentials(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "ops-admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "ops-secret")

	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"ops-admin","password":"ops-secret"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	token, ok := payload["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)
}

func TestHandleReplace_ReplacesObject(t *testing.T) {
	admin := &mockAdmin{
		getObjectResponse:    map[string]interface{}{"id": int64(9), "name": "old"},
		updateObjectResponse: map[string]interface{}{"id": float64(9), "name": "new"},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPut, "/api/products/9/", bytes.NewBufferString(`{"name":"new","active":true}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleReplace(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, int64(9), admin.lastUpdateID)
	require.Equal(t, "new", admin.lastUpdateData["name"])
	require.Equal(t, true, admin.lastUpdateData["active"])

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, "new", payload["name"])
}

func TestHandleReplace_ReturnsNotFoundWhenObjectMissing(t *testing.T) {
	admin := &mockAdmin{
		getObjectErr: assert.AnError,
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPut, "/api/products/9/", bytes.NewBufferString(`{"name":"new"}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleReplace(admin)(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleReplace_ChecksObjectPermission(t *testing.T) {
	admin := &mockAdmin{
		getObjectResponse: map[string]interface{}{"id": int64(9)},
		hasChangePermission: func(obj interface{}) bool {
			return obj == nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPut, "/api/products/9/", bytes.NewBufferString(`{"name":"new"}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleReplace(admin)(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestHandleReplace_RejectsEmptyBody(t *testing.T) {
	admin := &mockAdmin{
		getObjectResponse: map[string]interface{}{"id": int64(9)},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPut, "/api/products/9/", bytes.NewBufferString(`{}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleReplace(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleDetail_AllowsStringID(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			require.Equal(t, "sku-123", id)
			return map[string]interface{}{"id": id}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/products/sku-123/", nil)
	req = withURLParam(req, "id", "sku-123")
	rec := httptest.NewRecorder()

	router.handleDetail(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestHandleUpdate_AllowsStringID(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			require.Equal(t, "sku-123", id)
			return map[string]interface{}{"id": id, "name": data["name"]}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPatch, "/api/products/sku-123/", bytes.NewBufferString(`{"name":"updated"}`))
	req = withURLParam(req, "id", "sku-123")
	rec := httptest.NewRecorder()

	router.handleUpdate(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "sku-123", admin.lastUpdateID)
}

func TestHandleReplace_AllowsStringID(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			require.Equal(t, "sku-123", id)
			return map[string]interface{}{"id": id, "name": data["name"]}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPut, "/api/products/sku-123/", bytes.NewBufferString(`{"name":"replacement"}`))
	req = withURLParam(req, "id", "sku-123")
	rec := httptest.NewRecorder()

	router.handleReplace(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "sku-123", admin.lastUpdateID)
}

func TestHandleUpdate_ChecksObjectPermission(t *testing.T) {
	admin := &mockAdmin{
		getObjectResponse: map[string]interface{}{"id": int64(9)},
		hasChangePermission: func(obj interface{}) bool {
			return obj == nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPatch, "/api/products/9/", bytes.NewBufferString(`{"name":"new"}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleUpdate(admin)(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Len(t, admin.updateCalls, 0)
}

func TestHandleUpdate_ReturnsNotFoundWhenObjectMissing(t *testing.T) {
	admin := &mockAdmin{
		getObjectErr: assert.AnError,
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPatch, "/api/products/9/", bytes.NewBufferString(`{"name":"new"}`))
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleUpdate(admin)(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Len(t, admin.updateCalls, 0)
}

func TestHandleDelete_ChecksObjectPermission(t *testing.T) {
	admin := &mockAdmin{
		getObjectResponse: map[string]interface{}{"id": int64(9)},
		hasDeletePermission: func(obj interface{}) bool {
			return obj == nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/9/", nil)
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleDelete(admin)(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Len(t, admin.deleteCalls, 0)
}

func TestHandleDelete_ReturnsNotFoundWhenObjectMissing(t *testing.T) {
	admin := &mockAdmin{
		getObjectErr: assert.AnError,
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/9/", nil)
	req = withURLParam(req, "id", "9")
	rec := httptest.NewRecorder()

	router.handleDelete(admin)(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Len(t, admin.deleteCalls, 0)
}

func TestHandleDelete_AllowsStringID(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/sku-123/", nil)
	req = withURLParam(req, "id", "sku-123")
	rec := httptest.NewRecorder()

	router.handleDelete(admin)(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, admin.deleteCalls, 1)
	require.Equal(t, "sku-123", admin.deleteCalls[0])
}

func TestHandleBulkCreate_CreatesObjects(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"id":   float64(len(data)),
				"name": data["name"],
			}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`[{"name":"alpha"},{"name":"beta"}]`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.Len(t, admin.createCalls, 2)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, float64(2), payload["created"])
	objects, ok := payload["objects"].([]interface{})
	require.True(t, ok)
	require.Len(t, objects, 2)
}

func TestHandleBulkCreate_ReturnsMultiStatusOnPartialFailures(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			if data["name"] == "bad" {
				return nil, assert.AnError
			}
			return map[string]interface{}{"name": data["name"]}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`[{"name":"good"},{"name":"bad"}]`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusMultiStatus, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, float64(1), payload["created"])
	errors, ok := payload["errors"].([]interface{})
	require.True(t, ok)
	require.Len(t, errors, 1)
}

func TestHandleBulkCreate_RejectsInvalidPayloadShape(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`{"name":"single"}`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkCreate_RejectsEmptyArray(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`[]`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkCreate_ReturnsServerErrorWhenAllCreatesFail(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			return nil, assert.AnError
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`[{"name":"a"}]`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleBulkUpdate_UpdatesObjects(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id, "active": data["active"]}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", bytes.NewBufferString(`{"ids":[1,2],"data":{"active":true}}`))
	rec := httptest.NewRecorder()

	router.handleBulkUpdate(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, admin.updateCalls, 2)
	assert.Equal(t, int64(1), admin.updateCalls[0].ID)
	assert.Equal(t, int64(2), admin.updateCalls[1].ID)
	assert.Equal(t, true, admin.updateCalls[0].Data["active"])
}

func TestHandleBulkUpdate_ReturnsMultiStatusOnPartialFailures(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			if id == int64(2) {
				return nil, assert.AnError
			}
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", bytes.NewBufferString(`{"ids":[1,2],"data":{"active":true}}`))
	rec := httptest.NewRecorder()

	router.handleBulkUpdate(admin)(rec, req)
	require.Equal(t, http.StatusMultiStatus, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, float64(1), payload["updated"])
	errors, ok := payload["errors"].([]interface{})
	require.True(t, ok)
	require.Len(t, errors, 1)
}

func TestHandleBulkUpdate_RejectsMissingData(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", bytes.NewBufferString(`{"ids":[1]}`))
	rec := httptest.NewRecorder()

	router.handleBulkUpdate(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkUpdate_ReturnsServerErrorWhenAllUpdatesFail(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		updateObjectFn: func(id interface{}, data map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("boom")
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-update", bytes.NewBufferString(`{"ids":[1],"data":{"active":true}}`))
	rec := httptest.NewRecorder()

	router.handleBulkUpdate(admin)(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleBulkDelete_DeletesObjects(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, admin.deleteCalls, 2)
	assert.Equal(t, int64(1), admin.deleteCalls[0])
	assert.Equal(t, int64(2), admin.deleteCalls[1])
}

func TestHandleBulkDelete_ReturnsMultiStatusOnPartialFailures(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		deleteObjectFn: func(id interface{}) error {
			if id == int64(2) {
				return assert.AnError
			}
			return nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusMultiStatus, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Equal(t, float64(1), payload["deleted"])
	errors, ok := payload["errors"].([]interface{})
	require.True(t, ok)
	require.Len(t, errors, 1)
}

func TestHandleBulkDelete_RejectsEmptyIDs(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkDelete_ReturnsServerErrorWhenAllDeletesFail(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		deleteObjectFn: func(id interface{}) error {
			return assert.AnError
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[1]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleAction_RejectsEmptyIDs(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/mark_shipped", bytes.NewBufferString(`{"ids":[]}`))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "mark_shipped"))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleAction_ReturnsMultiStatusOnPartialFailures(t *testing.T) {
	admin := &mockAdmin{
		executeActionFn: func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
			return &core.BulkActionResponse{
				Success:  false,
				Affected: 1,
				Message:  "partial",
				Errors: []core.BulkActionError{
					{ID: 2, Code: "permission_denied", Message: "permission denied"},
				},
			}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/mark_shipped", bytes.NewBufferString(`{"ids":[1,2],"params":{"dry_run":false}}`))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "mark_shipped"))
	require.Equal(t, http.StatusMultiStatus, rec.Code)
	require.Equal(t, "mark_shipped", admin.lastActionName)
	require.Equal(t, []interface{}{float64(1), float64(2)}, admin.lastActionIDs)
	require.Equal(t, false, admin.lastActionParams["dry_run"])

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	errorsRaw, ok := payload["errors"].([]interface{})
	require.True(t, ok)
	require.Len(t, errorsRaw, 1)
	errorEntry, ok := errorsRaw[0].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "permission_denied", errorEntry["code"])
}

func TestHandleAction_ReturnsBadRequestWhenNoObjectsAffected(t *testing.T) {
	admin := &mockAdmin{
		executeActionFn: func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
			return &core.BulkActionResponse{
				Success:  false,
				Affected: 0,
				Message:  "none",
				Errors: []core.BulkActionError{
					{ID: 1, Code: "permission_denied", Message: "permission denied"},
				},
			}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/mark_shipped", bytes.NewBufferString(`{"ids":[1]}`))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "mark_shipped"))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func withURLParam(req *http.Request, key string, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	return req.WithContext(ctx)
}

func TestAdminSessionStore_TokenHashingAndRevocation(t *testing.T) {
	store := newAdminSessionStore()

	token, err := store.Issue("alice", 1*time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Raw token must NOT exist directly in store.sessions map (must be hashed)
	store.mu.RLock()
	_, rawFound := store.sessions[token]
	store.mu.RUnlock()
	assert.False(t, rawFound, "raw token must not be stored as key in sessions map")

	// Validate with token succeeds
	session, ok := store.Validate(token)
	require.True(t, ok)
	assert.Equal(t, "alice", session.Username)
	assert.True(t, session.Active)

	// Revoke token
	revoked := store.Revoke(token)
	assert.True(t, revoked)

	// Validate after revoke fails
	_, ok = store.Validate(token)
	assert.False(t, ok, "revoked session must fail validation")

	// Second revoke returns false
	assert.False(t, store.Revoke(token))
}

func TestHandleLogout_RevokesSession(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	registry := core.NewRegistry()
	apiRouter := NewRouter(registry)
	root := chi.NewRouter()
	apiRouter.RegisterRoutes(root)

	// Login
	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	loginRec := httptest.NewRecorder()
	root.ServeHTTP(loginRec, loginReq)
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginPayload map[string]interface{}
	require.NoError(t, json.Unmarshal(loginRec.Body.Bytes(), &loginPayload))
	token, ok := loginPayload["token"].(string)
	require.True(t, ok)

	// Config endpoint succeeds with token
	configReq := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	configReq.Header.Set("Authorization", "Bearer "+token)
	configRec := httptest.NewRecorder()
	root.ServeHTTP(configRec, configReq)
	require.Equal(t, http.StatusOK, configRec.Code)

	// Logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRec := httptest.NewRecorder()
	root.ServeHTTP(logoutRec, logoutReq)
	require.Equal(t, http.StatusOK, logoutRec.Code)

	// Config endpoint now fails with 401 Unauthorized
	configReq2 := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	configReq2.Header.Set("Authorization", "Bearer "+token)
	configRec2 := httptest.NewRecorder()
	root.ServeHTTP(configRec2, configReq2)
	require.Equal(t, http.StatusUnauthorized, configRec2.Code)
}

func TestHandleCreate_ReturnsBadRequestOnValidationError(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("validation failed: name is required")
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleCreate(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	errPayload, ok := payload["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "validation_error", errPayload["code"])
}

func TestHandleHistory_ReturnsAuditLog(t *testing.T) {
	admin := &mockAdmin{
		historyEntries: []core.LogEntry{
			{
				ID:         1,
				Timestamp:  time.Now(),
				ModelName:  "products",
				ObjectID:   "10",
				ObjectRepr: "Product #10",
				Action:     core.ActionAdd,
				UserID:     "admin",
			},
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodGet, "/api/products/10/history", nil)
	req = withURLParam(req, "id", "10")
	rec := httptest.NewRecorder()

	router.handleHistory(admin)(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	entries, ok := payload["entries"].([]interface{})
	require.True(t, ok)
	require.Len(t, entries, 1)
	entryMap := entries[0].(map[string]interface{})
	assert.Equal(t, "products", entryMap["model_name"])
	assert.Equal(t, "10", entryMap["object_id"])
	assert.Equal(t, "add", entryMap["action"])
}

func withURLParams(req *http.Request, params map[string]string) *http.Request {
	routeCtx := chi.NewRouteContext()
	for key, value := range params {
		routeCtx.URLParams.Add(key, value)
	}
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	return req.WithContext(ctx)
}

func registerViewableModel(t *testing.T, registry *core.Registry, admin *mockAdmin) {
	t.Helper()
	require.NoError(t, registry.Register(admin))
}

func TestHandleSavedViewSave_CreatesViewWith201(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{modelName: "products"})
	router := NewRouter(registry)

	req := httptest.NewRequest(http.MethodPost, "/api/saved-views/products", bytes.NewBufferString(`{"name":"Active","filters":{"active":"true"}}`))
	req = withURLParams(req, map[string]string{"model": "products"})
	rec := httptest.NewRecorder()

	router.handleSavedViewSave(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var view savedView
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &view))
	assert.Equal(t, "Active", view.Name)
	require.NotEmpty(t, view.ID)
}

func TestHandleSavedViewSave_UpsertsSameNameWith200(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{modelName: "products"})
	router := NewRouter(registry)

	save := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/saved-views/products", bytes.NewBufferString(body))
		req = withURLParams(req, map[string]string{"model": "products"})
		rec := httptest.NewRecorder()
		router.handleSavedViewSave(rec, req)
		return rec
	}

	first := save(`{"name":"Active","filters":{"a":"1"}}`)
	require.Equal(t, http.StatusCreated, first.Code)
	var firstView savedView
	require.NoError(t, json.Unmarshal(first.Body.Bytes(), &firstView))

	second := save(`{"name":"active","filters":{"a":"2"}}`)
	require.Equal(t, http.StatusOK, second.Code)
	var secondView savedView
	require.NoError(t, json.Unmarshal(second.Body.Bytes(), &secondView))
	assert.Equal(t, firstView.ID, secondView.ID)
}

func TestHandleSavedViewSave_RejectsMissingName(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{modelName: "products"})
	router := NewRouter(registry)

	req := httptest.NewRequest(http.MethodPost, "/api/saved-views/products", bytes.NewBufferString(`{"filters":{}}`))
	req = withURLParams(req, map[string]string{"model": "products"})
	rec := httptest.NewRecorder()

	router.handleSavedViewSave(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleSavedViewDelete_DeletesView(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{modelName: "products"})
	router := NewRouter(registry)

	req := httptest.NewRequest(http.MethodPost, "/api/saved-views/products", bytes.NewBufferString(`{"name":"Temp"}`))
	req = withURLParams(req, map[string]string{"model": "products"})
	rec := httptest.NewRecorder()
	router.handleSavedViewSave(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	var view savedView
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &view))

	delReq := httptest.NewRequest(http.MethodDelete, "/api/saved-views/products/"+view.ID, nil)
	delReq = withURLParams(delReq, map[string]string{"model": "products", "id": view.ID})
	delRec := httptest.NewRecorder()
	router.handleSavedViewDelete(delRec, delReq)
	require.Equal(t, http.StatusNoContent, delRec.Code)

	againRec := httptest.NewRecorder()
	router.handleSavedViewDelete(againRec, delReq)
	require.Equal(t, http.StatusNotFound, againRec.Code)
}

func TestUserKey_UsesMapUsername(t *testing.T) {
	assert.Equal(t, "alice", userKey(map[string]interface{}{"username": "alice", "role": "superuser"}))
	assert.Equal(t, "anonymous", userKey(nil))
}

func TestHandleAutocomplete_ClampsLimit(t *testing.T) {
	admin := &mockAdmin{}
	router := NewRouter(core.NewRegistry())

	for _, tc := range []struct {
		limit    string
		expected int
	}{
		{"9999", 50},
		{"-5", 10},
		{"abc", 10},
		{"", 10},
		{"25", 25},
	} {
		url := "/api/products/autocomplete?q=x"
		if tc.limit != "" {
			url += "&limit=" + tc.limit
		}
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		router.handleAutocomplete(admin)(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, tc.expected, admin.lastAutocompleteLimit, "limit=%q", tc.limit)
		assert.Equal(t, "x", admin.lastAutocompleteQuery)
	}
}

func TestHandleGlobalSearch_RespectsModelsParam(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{
		modelName: "products",
		autocompleteItems: []core.AutocompleteItem{
			{Value: 7, Label: "Laptop Pro"},
		},
	})
	registerViewableModel(t, registry, &mockAdmin{
		modelName: "orders",
		autocompleteItems: []core.AutocompleteItem{
			{Value: 9, Label: "Order 9"},
		},
	})
	router := NewRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=lap&models=products", nil)
	rec := httptest.NewRecorder()
	router.handleGlobalSearch(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload core.SearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Results, 1)
	assert.Equal(t, "products", payload.Results[0].Model)
	require.Len(t, payload.Results[0].Items, 1)
	assert.Equal(t, "/admin/products/7/view", payload.Results[0].Items[0].URL)
}

func TestHandleGlobalSearch_SkipsDeniedModels(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{
		modelName: "products",
		denyView:  true,
		autocompleteItems: []core.AutocompleteItem{
			{Value: 7, Label: "Laptop Pro"},
		},
	})
	router := NewRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=lap", nil)
	rec := httptest.NewRecorder()
	router.handleGlobalSearch(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload core.SearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	assert.Empty(t, payload.Results)
}

func TestHandleGlobalSearch_UsesAdminPrefix(t *testing.T) {
	registry := core.NewRegistry()
	registerViewableModel(t, registry, &mockAdmin{
		modelName: "products",
		autocompleteItems: []core.AutocompleteItem{
			{Value: 7, Label: "Laptop Pro"},
		},
	})
	router := NewRouter(registry).WithAdminPrefix("/custom-admin")

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=lap", nil)
	rec := httptest.NewRecorder()
	router.handleGlobalSearch(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var payload core.SearchResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
	require.Len(t, payload.Results, 1)
	assert.Equal(t, "/custom-admin/products/7/view", payload.Results[0].Items[0].URL)
}

func TestWithAdminPrefix_Normalizes(t *testing.T) {
	router := NewRouter(core.NewRegistry())
	assert.Equal(t, "/admin", router.adminPrefix)
	router.WithAdminPrefix("custom/")
	assert.Equal(t, "/custom", router.adminPrefix)
	router.WithAdminPrefix("")
	assert.Equal(t, "/admin", router.adminPrefix)
}

func TestHandleBulkCreate_AllValidationFailuresReturn400(t *testing.T) {
	admin := &mockAdmin{
		createObjectFn: func(data map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("validation failed: name is required")
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/bulk-create", bytes.NewBufferString(`[{"name":""}]`))
	rec := httptest.NewRecorder()

	router.handleBulkCreate(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkDelete_AllNotFoundReturn400(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return nil, assert.AnError
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleBulkDelete_AllPermissionDeniedReturn403(t *testing.T) {
	admin := &mockAdmin{
		getObjectFn: func(id interface{}) (interface{}, error) {
			return map[string]interface{}{"id": id}, nil
		},
		hasDeletePermission: func(obj interface{}) bool {
			return obj == nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodDelete, "/api/products/bulk-delete", bytes.NewBufferString(`{"ids":[1]}`))
	rec := httptest.NewRecorder()

	router.handleBulkDelete(admin)(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}
