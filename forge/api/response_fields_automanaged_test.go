package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// autoManagedTestModel mirrors the codegen shape: database-owned timestamps
// and a generated column keep Editable: true.
type autoManagedTestModel struct {
	schema.BaseSchema
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Slug      string    `json:"slug" db:"slug"`
}

func (autoManagedTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id"),
		schema.StringField("title"),
		{Name: "created_at", Type: schema.TypeDateTime, Editable: true, Serialize: true, AutoNowAdd: true},
		{Name: "updated_at", Type: schema.TypeDateTime, Editable: true, Serialize: true, AutoNow: true},
		{Name: "slug", Type: schema.TypeString, Editable: true, Serialize: true, Generated: true},
	}
}

type autoManagedTestSerializer struct {
	*BaseSerializer
}

func newAutoManagedTestSerializer() Serializer {
	return &autoManagedTestSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

type autoManagedTestManager struct {
	mu      sync.Mutex
	created *autoManagedTestModel
}

func (m *autoManagedTestManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*autoManagedTestModel)
	if !ok {
		return errors.New("unexpected model type")
	}
	cp := *item
	m.created = &cp
	return nil
}

func (m *autoManagedTestManager) Get(ctx context.Context, id int64) (interface{}, error) {
	return nil, errors.New("not found")
}

func (m *autoManagedTestManager) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (m *autoManagedTestManager) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func TestNonEditableFields_IncludesAutoManagedFields(t *testing.T) {
	got := NonEditableFields(&autoManagedTestModel{})
	assert.Contains(t, got, "created_at")
	assert.Contains(t, got, "updated_at")
	assert.Contains(t, got, "slug")
	assert.NotContains(t, got, "title")
}

func TestBaseViewSet_Create_IgnoresAutoManagedRequestValues(t *testing.T) {
	mgr := &autoManagedTestManager{}
	vs := NewBaseViewSet(
		newAutoManagedTestSerializer,
		mgr,
		&autoManagedTestModel{},
	)
	vs.ReadOnlyRequestFields = NonEditableFields(&autoManagedTestModel{})

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(
		`{"title":"hello","created_at":"2020-01-01T00:00:00Z","updated_at":"2020-01-02T00:00:00Z","slug":"hacker-slug"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, mgr.created)
	assert.Equal(t, "hello", mgr.created.Title)
	assert.True(t, mgr.created.CreatedAt.IsZero(), "AutoNowAdd value must be ignored")
	assert.True(t, mgr.created.UpdatedAt.IsZero(), "AutoNow value must be ignored")
	assert.Equal(t, "", mgr.created.Slug, "generated column value must be ignored")
}
