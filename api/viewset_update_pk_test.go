package api

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type updateTestItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type updateTestItemSerializer struct {
	*BaseSerializer
}

func newUpdateTestItemSerializer() Serializer {
	return &updateTestItemSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *updateTestItemSerializer) Fields() []string {
	return []string{"id", "title"}
}

type fakeUpdateManager struct {
	mu    sync.Mutex
	items map[int64]*updateTestItem
}

func (m *fakeUpdateManager) Create(ctx context.Context, item interface{}) error {
	return nil
}

func (m *fakeUpdateManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
	}
	// Return a copy so in-memory pointer mutation doesn't alias map entry directly
	copy := *item
	return &copy, nil
}

func (m *fakeUpdateManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*updateTestItem)
	m.items[item.ID] = item
	return nil
}

func TestUpdate_IgnoresBodyIDAndDoesNotModifyOtherRow(t *testing.T) {
	mgr := &fakeUpdateManager{
		items: map[int64]*updateTestItem{
			1: {ID: 1, Title: "row 1 original"},
			2: {ID: 2, Title: "row 2 original"},
		},
	}

	vs := NewBaseViewSet(
		newUpdateTestItemSerializer,
		mgr,
		&updateTestItem{},
	)

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	body := `{"id": 2, "title": "row 1 updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/items/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Row 2 must remain untouched
	assert.Equal(t, "row 2 original", mgr.items[2].Title, "row 2 must be unchanged")
	assert.Equal(t, int64(2), mgr.items[2].ID, "row 2 ID must be 2")

	// Row 1 must have been updated
	assert.Equal(t, "row 1 updated", mgr.items[1].Title, "row 1 title should be updated")
	assert.Equal(t, int64(1), mgr.items[1].ID, "row 1 ID must remain 1")
}

type codeTestModel struct {
	schema.BaseSchema
	Code  string `json:"code" db:"code"`
	Title string `json:"title" db:"title"`
}

func (codeTestModel) Fields() []schema.Field {
	return []schema.Field{
		{Name: "code", Type: schema.TypeString, PrimaryKey: true},
		{Name: "title", Type: schema.TypeString},
	}
}

type codeTestModelSerializer struct {
	*BaseSerializer
}

func newCodeTestModelSerializer() Serializer {
	return &codeTestModelSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *codeTestModelSerializer) Fields() []string {
	return []string{"code", "title"}
}

type fakeCodeUpdateManager struct {
	mu    sync.Mutex
	items map[string]*codeTestModel
}

func (m *fakeCodeUpdateManager) Create(ctx context.Context, item interface{}) error {
	return nil
}

func (m *fakeCodeUpdateManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("row%d", id)
	item, ok := m.items[key]
	if !ok {
		return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
	}
	copy := *item
	return &copy, nil
}

func (m *fakeCodeUpdateManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item := model.(*codeTestModel)
	m.items[item.Code] = item
	return nil
}

func TestUpdate_SchemaPrimaryKey_IgnoresBodyCodeAndDoesNotModifyOtherRow(t *testing.T) {
	mgr := &fakeCodeUpdateManager{
		items: map[string]*codeTestModel{
			"row1":  {Code: "row1", Title: "row 1 original"},
			"other": {Code: "other", Title: "other original"},
		},
	}

	vs := NewBaseViewSet(
		newCodeTestModelSerializer,
		mgr,
		&codeTestModel{},
	)

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	body := `{"code": "other", "title": "row 1 updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/items/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Row "other" must remain untouched
	assert.Equal(t, "other original", mgr.items["other"].Title, "row 'other' must be unchanged")
	assert.Equal(t, "other", mgr.items["other"].Code, "row 'other' code must be 'other'")

	// Row "row1" must have been updated
	require.NotNil(t, mgr.items["row1"], "row 'row1' must exist")
	assert.Equal(t, "row 1 updated", mgr.items["row1"].Title, "row 1 title should be updated")
	assert.Equal(t, "row1", mgr.items["row1"].Code, "row 1 code must remain 'row1'")
}
