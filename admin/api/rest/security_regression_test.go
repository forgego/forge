package rest

import (
	"context"
	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/schema"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type objectViewAdmin struct {
	mockAdmin
	historyCalls int
}

func (a *objectViewAdmin) HasViewPermission(_ context.Context, _ any, obj any) bool {
	return obj == nil
}
func (a *objectViewAdmin) GetHistory(context.Context, string) ([]core.LogEntry, error) {
	a.historyCalls++
	return []core.LogEntry{{ChangeStats: `{"private":"history"}`}}, nil
}
func TestDetailAndHistoryEnforceObjectViewPermission(t *testing.T) {
	admin := &objectViewAdmin{mockAdmin: mockAdmin{getObjectResponse: map[string]any{"id": int64(7), "name": "private"}}}
	router := NewRouter(core.NewRegistry())
	for _, route := range []string{"detail", "history"} {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/items/7", nil)
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", "7")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			response := httptest.NewRecorder()
			if route == "detail" {
				router.handleDetail(admin)(response, req)
			} else {
				router.handleHistory(admin)(response, req)
			}
			assert.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
			assert.NotContains(t, response.Body.String(), "private")
		})
	}
	assert.Zero(t, admin.historyCalls, "denied history must not be loaded")
}

type privateResponseModel struct {
	schema.BaseSchema
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"credential,omitempty" db:"private_value"`
}

func (privateResponseModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()), schema.StringField("name"), schema.StringField("private_value", schema.WriteOnly()),
	}
}

type privateResponseAdmin struct{ mockAdmin }

func (*privateResponseAdmin) ModelType() reflect.Type { return reflect.TypeOf(privateResponseModel{}) }
func TestAdminResponsesExcludeWriteOnlyFields(t *testing.T) {
	object := &privateResponseModel{ID: 9007199254740993, Name: "public", Secret: "synthetic-private-value"}
	admin := &privateResponseAdmin{mockAdmin: mockAdmin{
		getObjectResponse: object, updateObjectResponse: object,
		createObjectFn: func(map[string]any) (any, error) { return object, nil },
		listResponse:   &core.PaginatedResponse{Count: 1, Results: []map[string]any{{"id": int64(9007199254740993), "name": "public", "private_value": "synthetic-private-value", "credential": "synthetic-private-value"}}},
		metadata:       &core.Metadata{ListDisplay: []string{"name", "credential", "private_value"}},
	}}
	router := NewRouter(core.NewRegistry())
	for _, test := range []struct {
		name, method, body string
		handler            http.HandlerFunc
		status             int
	}{
		{"detail", http.MethodGet, "", router.handleDetail(admin), 200},
		{"list", http.MethodGet, "", router.handleList(admin), 200},
		{"create", http.MethodPost, `{"name":"public"}`, router.handleCreate(admin), 201},
		{"patch", http.MethodPatch, `{"name":"public"}`, router.handleUpdate(admin), 200},
		{"replace", http.MethodPut, `{"name":"public"}`, router.handleReplace(admin), 200},
		{"bulk-create", http.MethodPost, `[{"name":"public"}]`, router.handleBulkCreate(admin), 201},
		{"bulk-update", http.MethodPatch, `{"ids":[1],"data":{"name":"public"}}`, router.handleBulkUpdate(admin), 200},
		{"export", http.MethodGet, "", router.handleExport(admin), 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.method, "/items/1", strings.NewReader(test.body))
			routeCtx := chi.NewRouteContext()
			routeCtx.URLParams.Add("id", "1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
			response := httptest.NewRecorder()
			test.handler(response, req)
			assert.Equal(t, test.status, response.Code, response.Body.String())
			assert.NotContains(t, response.Body.String(), "synthetic-private-value")
			assert.NotContains(t, response.Body.String(), `"private_value"`)
			assert.NotContains(t, response.Body.String(), `"credential"`)
			if test.name == "detail" {
				assert.Contains(t, response.Body.String(), "9007199254740993", "integer precision must survive filtering")
			}
		})
	}
}

func TestExportPreservesLargeIntegerFromTypedRows(t *testing.T) {
	admin := &privateResponseAdmin{mockAdmin: mockAdmin{
		metadata:     &core.Metadata{ListDisplay: []string{"id", "name"}},
		listResponse: &core.PaginatedResponse{Count: 1, Results: []*privateResponseModel{{ID: 9007199254740993, Name: "large-id"}}},
	}}
	response := httptest.NewRecorder()
	NewRouter(core.NewRegistry()).handleExport(admin)(response, httptest.NewRequest(http.MethodGet, "/items/export", nil))
	assert.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Contains(t, response.Body.String(), "9007199254740993")
}
