package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/require"
)

type managerRegressionItem struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (managerRegressionItem) Fields() []schema.Field {
	return []schema.Field{
		{Name: "id", Type: schema.TypeInt64, PrimaryKey: true},
		{Name: "name", Type: schema.TypeString},
	}
}

type managerRegressionSerializer struct {
	*BaseSerializer
}

func newManagerRegressionSerializer() Serializer {
	return &managerRegressionSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *managerRegressionSerializer) Fields() []string { return []string{"id", "name"} }

type managerRegressionQuerySet struct {
	mgr    *managerRegressionStore
	limit  int
	offset int
}

func (qs *managerRegressionQuerySet) Count(ctx context.Context) (int64, error) {
	return int64(len(qs.mgr.items)), nil
}

func (qs *managerRegressionQuerySet) Offset(n int) interface{} {
	qs.offset = n
	return qs
}

func (qs *managerRegressionQuerySet) Limit(n int) interface{} {
	qs.limit = n
	return qs
}

func (qs *managerRegressionQuerySet) Filter(interface{}) interface{} { return qs }

func (qs *managerRegressionQuerySet) OrderBy(...interface{}) interface{} { return qs }

func (qs *managerRegressionQuerySet) All(ctx context.Context) ([]*managerRegressionItem, error) {
	qs.mgr.mu.RLock()
	defer qs.mgr.mu.RUnlock()
	start := qs.offset
	if start > len(qs.mgr.items) {
		start = len(qs.mgr.items)
	}
	end := start + qs.limit
	if qs.limit <= 0 || end > len(qs.mgr.items) {
		end = len(qs.mgr.items)
	}
	out := make([]*managerRegressionItem, 0, end-start)
	for _, it := range qs.mgr.items[start:end] {
		cp := *it
		out = append(out, &cp)
	}
	return out, nil
}

type managerRegressionStore struct {
	mu     sync.RWMutex
	items  []*managerRegressionItem
	nextID int64
}

func (m *managerRegressionStore) QuerySet() interface{} {
	return &managerRegressionQuerySet{mgr: m}
}

func (m *managerRegressionStore) Count(ctx context.Context) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.items)), nil
}

func (m *managerRegressionStore) All(ctx context.Context) ([]*managerRegressionItem, error) {
	return (&managerRegressionQuerySet{mgr: m}).All(ctx)
}

func (m *managerRegressionStore) Create(ctx context.Context, model interface{}) error {
	item, ok := model.(*managerRegressionItem)
	if !ok {
		return fmt.Errorf("unexpected model type %T", model)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	item.ID = m.nextID
	cp := *item
	m.items = append(m.items, &cp)
	return nil
}

func (m *managerRegressionStore) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, it := range m.items {
		if it.ID == id {
			cp := *it
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *managerRegressionStore) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (m *managerRegressionStore) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func TestViewSet_BuiltWithManager_CreateRetrieveListSucceed(t *testing.T) {
	mgr := &managerRegressionStore{}

	vs := NewBaseViewSet(
		newManagerRegressionSerializer,
		mgr,
		&managerRegressionItem{},
	)

	router := NewRouter("/api")
	router.Register("widgets", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	// Create via the bare manager (exercises getManager needing Create).
	createReq := httptest.NewRequest(http.MethodPost, "/api/widgets/", strings.NewReader(`{"name":"w1"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	handler.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code, "create with bare manager must not 500: %s", createRec.Body.String())

	// Retrieve via the bare manager (exercises getManager needing Get).
	getReq := httptest.NewRequest(http.MethodGet, "/api/widgets/1", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code, "retrieve with bare manager must not 500: %s", getRec.Body.String())

	// Paginated list converts the bare manager to a queryset.
	listReq := httptest.NewRequest(http.MethodGet, "/api/widgets/?page=1&page_size=20", nil)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code, "paginated list must succeed: %s", listRec.Body.String())

	var resp struct {
		Count   int                      `json:"count"`
		Results []map[string]interface{} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &resp))
	require.Equal(t, 1, resp.Count)
	require.Len(t, resp.Results, 1)
}
