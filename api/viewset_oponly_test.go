package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type opOnlyItem struct {
	schema.BaseSchema
	ID     int64  `db:"id" json:"id"`
	Title  string `db:"title" json:"title"`
	Status string `db:"status" json:"status"`
}

func (opOnlyItem) Meta() schema.Meta {
	return schema.Meta{
		TableName:           "op_only_items",
		OperationOnlyFields: []string{"status"},
	}
}

func (opOnlyItem) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
		schema.StringField("status", schema.OperationOnly()),
	}
}

type fakeOpOnlyManager struct {
	unusedListOperations
	unusedDeleteOperation
	mu    sync.Mutex
	items map[int64]*opOnlyItem
}

func (m *fakeOpOnlyManager) Create(ctx context.Context, item interface{}) error {
	return nil
}

func (m *fakeOpOnlyManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := m.items[id]; ok {
		copy := *item
		return &copy, nil
	}
	return nil, nil
}

func (m *fakeOpOnlyManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if item, ok := model.(*opOnlyItem); ok {
		copy := *item
		m.items[item.ID] = &copy
	}
	return nil
}

func TestBaseViewSet_BlockOperationOnlyFields(t *testing.T) {
	mgr := &fakeOpOnlyManager{
		items: map[int64]*opOnlyItem{
			1: {ID: 1, Title: "Item 1", Status: "pending"},
		},
	}

	vs := NewBaseViewSet(
		func() Serializer { return NewBaseSerializer(nil) },
		mgr,
		&opOnlyItem{},
	)
	vs.OperationOnlyFields = []string{"status"}
	vs.Authentication = nil
	vs.Permissions = nil

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	t.Run("PartialUpdate blocks operation-only field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(`{"status": "approved"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "operation_only_field", body["code"])
		assert.Equal(t, "pending", mgr.items[1].Status)
	})

	t.Run("Update blocks operation-only field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/items/1", bytes.NewBufferString(`{"status": "approved", "title": "Item 1"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "operation_only_field", body["code"])
		assert.Equal(t, "pending", mgr.items[1].Status)
	})

	t.Run("PartialUpdate of allowed field succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(`{"title": "Updated Title"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Updated Title", mgr.items[1].Title)
		assert.Equal(t, "pending", mgr.items[1].Status)
	})
}
