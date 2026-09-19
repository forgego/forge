package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type readOnlyTestItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	CreatedBy string `json:"created_by"`
}

type readOnlyTestSerializer struct {
	*BaseSerializer
}

func newReadOnlyTestSerializer() Serializer {
	return &readOnlyTestSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *readOnlyTestSerializer) Fields() []string {
	return []string{"id", "title", "created_by"}
}

type readOnlyTestManager struct {
	mu      sync.Mutex
	items   map[int64]*readOnlyTestItem
	created *readOnlyTestItem
	updated *readOnlyTestItem
	nextID  int64
}

func (m *readOnlyTestManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*readOnlyTestItem)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.nextID++
	item.ID = m.nextID
	cp := *item
	m.created = &cp
	if m.items == nil {
		m.items = make(map[int64]*readOnlyTestItem)
	}
	m.items[item.ID] = &cp
	return nil
}

func (m *readOnlyTestManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
	}
	cp := *item
	return &cp, nil
}

func (m *readOnlyTestManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*readOnlyTestItem)
	if !ok {
		return errors.New("unexpected model type")
	}
	cp := *item
	m.updated = &cp
	m.items[item.ID] = &cp
	return nil
}

func (m *readOnlyTestManager) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func TestBaseViewSet_ReadOnlyRequestFields_IgnoredOnCreateAndUpdate(t *testing.T) {
	mgr := &readOnlyTestManager{
		items: map[int64]*readOnlyTestItem{
			1: {ID: 1, Title: "original", CreatedBy: "original-author"},
		},
		nextID: 1,
	}
	vs := NewBaseViewSet(
		newReadOnlyTestSerializer,
		mgr,
		&readOnlyTestItem{},
	)
	vs.ReadOnlyRequestFields = []string{"created_by"}

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	// POST: created_by must be ignored, title must be applied.
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":"hello","created_by":"hacker"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, mgr.created)
	assert.Equal(t, "hello", mgr.created.Title)
	assert.Equal(t, "", mgr.created.CreatedBy)

	// PATCH: created_by must be ignored, title must be applied.
	req = httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(`{"title":"updated","created_by":"hacker"}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, mgr.updated)
	assert.Equal(t, "updated", mgr.updated.Title)
	assert.Equal(t, "original-author", mgr.updated.CreatedBy)
}
