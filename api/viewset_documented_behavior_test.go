package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/renderers"
	forgehttp "github.com/forgego/forge/server"
)

func TestRequestHeadersSelectParserAndRenderer(t *testing.T) {
	n := NewContentNegotiator([]renderers.Renderer{renderers.NewJSONRenderer(), renderers.NewXMLRenderer()}, []parsers.Parser{parsers.NewJSONParser(), parsers.NewFormParser()})
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Accept", "application/xml")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if n.SelectRenderer(r).MediaType() != "application/xml" || n.SelectParser(r).MediaType() != "application/x-www-form-urlencoded" {
		t.Fatal("headers did not select the requested formats")
	}
}

func newDocumentedItemsRouter(items []listPaginationTestItem) (*forgehttp.Router, *spyPaginatedManager) {
	mgr := &spyPaginatedManager{items: items}
	vs := NewBaseViewSet(newListPaginationSerializer, mgr, &listPaginationTestItem{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler, mgr
}

func TestListUsesPageNumberWhenOtherPaginationParametersAreSupplied(t *testing.T) {
	handler, mgr := newDocumentedItemsRouter([]listPaginationTestItem{{ID: 1, Name: "one"}, {ID: 2, Name: "two"}, {ID: 3, Name: "three"}})
	r := httptest.NewRequest(http.MethodGet, "/api/items/?page=2&page_size=1&limit=3&offset=0&cursor=opaque", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != http.StatusOK || mgr.receivedOffset != 1 || mgr.receivedLimit != 1 {
		t.Fatalf("status=%d offset=%d limit=%d body=%s", w.Code, mgr.receivedOffset, mgr.receivedLimit, w.Body.String())
	}
}

func TestListPastEndPageReturnsEmptyResults(t *testing.T) {
	handler, _ := newDocumentedItemsRouter([]listPaginationTestItem{{ID: 1, Name: "one"}})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/items/?page=5&page_size=1", nil))
	var body struct {
		Count   int           `json:"count"`
		Results []interface{} `json:"results"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || body.Count != 1 || len(body.Results) != 0 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestDestroyDeletesRecord(t *testing.T) {
	manager, handler := setupAPISQLiteViewSet(t)
	item := &apiSQLiteModel{Email: "delete@example.test"}
	if err := manager.Create(context.Background(), item); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/api/items/1", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	count, err := manager.Count(context.Background())
	if err != nil || count != 0 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

type denyDocumentedAction struct{}

func (denyDocumentedAction) HasPermission(*http.Request, permissions.ViewSet) bool { return false }
func (denyDocumentedAction) HasObjectPermission(*http.Request, permissions.ViewSet, interface{}) bool {
	return false
}
func (denyDocumentedAction) GetMessage() string { return "denied" }
func (denyDocumentedAction) GetCode() string    { return "denied" }

func TestCustomActionBypassesViewSetAccessChecks(t *testing.T) {
	router := NewRouter("/api")
	vs := createRouterTestViewSet()
	vs.Permissions = []permissions.Permission{denyDocumentedAction{}}
	router.Register("items", vs)
	router.Action("items", "ping", ActionConfig{}, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusAccepted) })
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	custom := httptest.NewRecorder()
	handler.ServeHTTP(custom, httptest.NewRequest(http.MethodGet, "/api/items/ping", nil))
	standard := httptest.NewRecorder()
	handler.ServeHTTP(standard, httptest.NewRequest(http.MethodGet, "/api/items/", nil))
	if custom.Code != http.StatusAccepted || standard.Code != http.StatusForbidden {
		t.Fatalf("custom=%d standard=%d", custom.Code, standard.Code)
	}
}

func TestConfigurationErrorNamesOperationsWithoutConfiguredValues(t *testing.T) {
	qs := &checkReadQueryset{DSN: checkSecretDSN}
	err := newCheckViewSet(qs).CheckConfiguration()
	if err == nil || !strings.Contains(err.Error(), "Create") || strings.Contains(err.Error(), checkSecretDSN) {
		t.Fatalf("error=%v", err)
	}
}

type computedTestSerializer struct{ *BaseSerializer }

func (s *computedTestSerializer) ToJSON() ([]byte, error) {
	return []byte(`{"id":42,"computed":"from serializer"}`), nil
}
func (s *computedTestSerializer) Fields() []string { return []string{"id", "computed"} }

func TestBaseViewSetResponsesContainOnlyModelFields(t *testing.T) {
	vs := NewBaseViewSet(func() Serializer { return &computedTestSerializer{NewBaseSerializer(nil)} }, &routerTestQueryset{items: []*routerTestModel{{ID: 42}}}, &routerTestModel{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/items/42", nil))
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || body["id"] != float64(42) {
		t.Fatalf("status=%d body=%v", w.Code, body)
	}
	if _, ok := body["computed"]; ok {
		t.Fatalf("computed serializer value leaked into response: %v", body)
	}
}
