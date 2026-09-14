package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type routerTestModel struct {
	ID int64 `json:"id"`
}

type routerTestQueryset struct {
	items []*routerTestModel
}

func (q *routerTestQueryset) All(context.Context) ([]*routerTestModel, error) {
	return q.items, nil
}

func (q *routerTestQueryset) Count(context.Context) (int64, error) {
	return int64(len(q.items)), nil
}

func (q *routerTestQueryset) Offset(int) *routerTestQueryset            { return q }
func (q *routerTestQueryset) Limit(int) *routerTestQueryset             { return q }
func (q *routerTestQueryset) Create(context.Context, interface{}) error { return nil }

func (q *routerTestQueryset) Get(_ context.Context, id int64) (*routerTestModel, error) {
	for _, item := range q.items {
		if item.ID == id {
			return item, nil
		}
	}
	return nil, nil
}

type routerTestSerializer struct {
	*BaseSerializer
}

func newRouterTestSerializer() Serializer {
	return &routerTestSerializer{BaseSerializer: NewBaseSerializer(make(map[string]interface{}))}
}

func (s *routerTestSerializer) Fields() []string         { return []string{"id"} }
func (s *routerTestSerializer) ReadOnlyFields() []string { return []string{"id"} }

func createRouterTestViewSet() *BaseViewSet {
	return NewBaseViewSet(
		newRouterTestSerializer,
		&routerTestQueryset{items: []*routerTestModel{{ID: 42}}},
		&routerTestModel{},
	)
}

func TestRouter_Actions_ReachableAndParams(t *testing.T) {
	router := NewRouter("/api/v1")
	viewset := createRouterTestViewSet()
	router.Register("items", viewset)

	var listActionCalled bool
	listHandler := func(w http.ResponseWriter, r *http.Request) {
		listActionCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"list_published"}`))
	}

	var detailActionCalled bool
	var capturedID string
	detailHandler := func(w http.ResponseWriter, r *http.Request) {
		detailActionCalled = true
		capturedID = forgehttp.URLParam(r, "id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"detail_published"}`))
	}

	router.Action("items", "publish_all", ActionConfig{
		Methods: []string{http.MethodPost},
		Detail:  false,
	}, listHandler)

	router.Action("items", "publish", ActionConfig{
		Methods: []string{http.MethodGet},
		Detail:  true,
	}, detailHandler)

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	// 1. List action with POST is reachable
	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/items/publish_all", nil)
	postRec := httptest.NewRecorder()
	httpRouter.ServeHTTP(postRec, postReq)
	assert.True(t, listActionCalled)
	assert.Equal(t, http.StatusOK, postRec.Code)
	assert.Equal(t, `{"status":"list_published"}`, postRec.Body.String())

	// 2. Detail action with GET is reachable and receives "id" URL param
	detailReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/42/publish", nil)
	detailRec := httptest.NewRecorder()
	httpRouter.ServeHTTP(detailRec, detailReq)
	assert.True(t, detailActionCalled)
	assert.Equal(t, http.StatusOK, detailRec.Code)
	assert.Equal(t, "42", capturedID)
	assert.Equal(t, `{"status":"detail_published"}`, detailRec.Body.String())

	// 3. /resource/{id} still hits Retrieve
	retrieveReq := httptest.NewRequest(http.MethodGet, "/api/v1/items/42", nil)
	retrieveRec := httptest.NewRecorder()
	httpRouter.ServeHTTP(retrieveRec, retrieveReq)
	assert.Equal(t, http.StatusOK, retrieveRec.Code)
	assert.Contains(t, retrieveRec.Body.String(), `"id":42`)

	// 4. OPTIONS returns 200 with metadata JSON
	optionsReq := httptest.NewRequest(http.MethodOptions, "/api/v1/items/", nil)
	optionsRec := httptest.NewRecorder()
	httpRouter.ServeHTTP(optionsRec, optionsReq)
	assert.Equal(t, http.StatusOK, optionsRec.Code)
	assert.Contains(t, optionsRec.Header().Get("Content-Type"), "application/json")
}

func TestRouter_Action_UnknownMethodPanics(t *testing.T) {
	router := NewRouter("/api/v1")
	assert.PanicsWithValue(t, "unknown HTTP method: INVALID", func() {
		router.Action("items", "bad", ActionConfig{
			Methods: []string{"INVALID"},
		}, func(w http.ResponseWriter, r *http.Request) {})
	})
}

func TestRouter_Action_UnregisteredResource(t *testing.T) {
	router := NewRouter("/api/v1")

	var called bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	// Registering for an unregistered resource returns no error at registration
	router.Action("unregistered", "custom", ActionConfig{
		Methods: []string{http.MethodGet},
	}, handler)

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unregistered/custom", nil)
	rec := httptest.NewRecorder()
	httpRouter.ServeHTTP(rec, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_Action_DefaultValues(t *testing.T) {
	router := NewRouter("/api/v1")

	var called bool
	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	// Empty methods defaults to GET, empty URLPath defaults to name
	router.Action("items", "status", ActionConfig{}, handler)

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/status", nil)
	rec := httptest.NewRecorder()
	httpRouter.ServeHTTP(rec, req)

	require.True(t, called)
	assert.Equal(t, http.StatusOK, rec.Code)
}
