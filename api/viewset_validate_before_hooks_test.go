package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hookRecordingModel fails validation when title is empty and records when
// its business hooks run, mirroring models whose BeforeCreate/BeforeUpdate
// hooks have side effects (notifications, external writes).
type hookRecordingModel struct {
	ID              int64  `json:"id" db:"id"`
	Title           string `json:"title" db:"title"`
	beforeCreateRan bool
	beforeUpdateRan bool
}

func (m *hookRecordingModel) Validate() error {
	if m.Title == "" {
		return errors.New("title is required")
	}
	return nil
}

func (m *hookRecordingModel) BeforeCreate(ctx context.Context) error {
	m.beforeCreateRan = true
	return nil
}

func (m *hookRecordingModel) BeforeUpdate(ctx context.Context) error {
	m.beforeUpdateRan = true
	return nil
}

type hookRecordingManager struct {
	mu       sync.Mutex
	items    map[int64]*hookRecordingModel
	nextID   int64
	lastSeen *hookRecordingModel
}

func (m *hookRecordingManager) runBeforeHook(ctx context.Context, model *hookRecordingModel, update bool) {
	// Mirror orm.Manager: business hooks run inside Create/Update.
	if update {
		_ = model.BeforeUpdate(ctx)
	} else {
		_ = model.BeforeCreate(ctx)
	}
}

func (m *hookRecordingManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*hookRecordingModel)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.runBeforeHook(ctx, item, false)
	m.nextID++
	item.ID = m.nextID
	cp := *item
	m.lastSeen = &cp
	if m.items == nil {
		m.items = make(map[int64]*hookRecordingModel)
	}
	m.items[item.ID] = &cp
	return nil
}

func (m *hookRecordingManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *item
	return &cp, nil
}

func (m *hookRecordingManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*hookRecordingModel)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.runBeforeHook(ctx, item, true)
	cp := *item
	m.lastSeen = &cp
	m.items[item.ID] = &cp
	return nil
}

func (m *hookRecordingManager) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func newHookRecordingHandler(mgr *hookRecordingManager) *forgehttp.Router {
	vs := NewBaseViewSet(newHookRecordingSerializer, mgr, &hookRecordingModel{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

type hookRecordingSerializer struct {
	*BaseSerializer
}

func newHookRecordingSerializer() Serializer {
	return &hookRecordingSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func TestBaseViewSet_InvalidPayloadFailsBeforeBusinessHooks(t *testing.T) {
	mgr := &hookRecordingManager{}
	handler := newHookRecordingHandler(mgr)

	// Create with an invalid payload: 400 and BeforeCreate must not run.
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Nil(t, mgr.lastSeen, "manager must not be called for an invalid payload")

	// Seed a valid item directly, then PATCH it invalid: 400, no BeforeUpdate.
	mgr.mu.Lock()
	mgr.nextID = 1
	mgr.items = map[int64]*hookRecordingModel{1: {ID: 1, Title: "original"}}
	mgr.mu.Unlock()

	req = httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	require.NotNil(t, mgr.items[1])
	assert.False(t, mgr.items[1].beforeUpdateRan, "BeforeUpdate must not run for an invalid payload")
	assert.Equal(t, "original", mgr.items[1].Title)
}

func TestBaseViewSet_ValidPayloadStillRunsBusinessHooks(t *testing.T) {
	mgr := &hookRecordingManager{}
	handler := newHookRecordingHandler(mgr)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, mgr.lastSeen)
	assert.True(t, mgr.lastSeen.beforeCreateRan, "BeforeCreate must still run for valid payloads")
}
