package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listPaginationTestItem struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (listPaginationTestItem) Fields() []schema.Field {
	return []schema.Field{
		{Name: "id", Type: schema.TypeInt64, PrimaryKey: true},
		{Name: "name", Type: schema.TypeString},
	}
}

type listPaginationSerializer struct {
	*BaseSerializer
}

func newListPaginationSerializer() Serializer {
	return &listPaginationSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *listPaginationSerializer) Fields() []string {
	return []string{"id", "name"}
}

type spyPaginatedQuerySet struct {
	mgr    *spyPaginatedManager
	limit  int
	offset int
}

func (qs *spyPaginatedQuerySet) Count(ctx context.Context) (int64, error) {
	return int64(len(qs.mgr.items)), nil
}

func (qs *spyPaginatedQuerySet) Offset(n int) interface{} {
	qs.offset = n
	qs.mgr.receivedOffset = n
	qs.mgr.offsetCalled = true
	return qs
}

func (qs *spyPaginatedQuerySet) Limit(n int) interface{} {
	qs.limit = n
	qs.mgr.receivedLimit = n
	qs.mgr.limitCalled = true
	return qs
}

func (qs *spyPaginatedQuerySet) All(ctx context.Context) ([]*listPaginationTestItem, error) {
	start := qs.offset
	if start > len(qs.mgr.items) {
		start = len(qs.mgr.items)
	}
	end := start + qs.limit
	if qs.limit <= 0 || end > len(qs.mgr.items) {
		end = len(qs.mgr.items)
	}
	slice := qs.mgr.items[start:end]
	res := make([]*listPaginationTestItem, len(slice))
	for i := range slice {
		res[i] = &slice[i]
	}
	return res, nil
}

// spyPaginatedManager mimics an orm.Manager that exposes QuerySet constructor.
// Without QuerySet conversion in List, it only has All/Count without Offset/Limit,
// which causes the unpaginated full table to be loaded.
type spyPaginatedManager struct {
	items          []listPaginationTestItem
	receivedLimit  int
	receivedOffset int
	limitCalled    bool
	offsetCalled   bool
}

func (m *spyPaginatedManager) QuerySet() interface{} {
	return &spyPaginatedQuerySet{mgr: m}
}

func (m *spyPaginatedManager) Count(ctx context.Context) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *spyPaginatedManager) All(ctx context.Context) ([]*listPaginationTestItem, error) {
	// If Offset/Limit were not applied, return all rows (the bug)
	res := make([]*listPaginationTestItem, len(m.items))
	for i := range m.items {
		res[i] = &m.items[i]
	}
	return res, nil
}

func (m *spyPaginatedManager) Create(ctx context.Context, model interface{}) error { return nil }
func (m *spyPaginatedManager) Get(ctx context.Context, id int64) (interface{}, error) {
	return nil, nil
}
func (m *spyPaginatedManager) Update(ctx context.Context, model interface{}) error { return nil }
func (m *spyPaginatedManager) Delete(ctx context.Context, model interface{}) error { return nil }

func TestList_PaginatedRequest_AppliesOffsetLimitToQueryset(t *testing.T) {
	items := []listPaginationTestItem{
		{ID: 1, Name: "Item 1"},
		{ID: 2, Name: "Item 2"},
		{ID: 3, Name: "Item 3"},
		{ID: 4, Name: "Item 4"},
		{ID: 5, Name: "Item 5"},
	}

	mgr := &spyPaginatedManager{items: items}

	vs := NewBaseViewSet(
		newListPaginationSerializer,
		mgr,
		&listPaginationTestItem{},
	)

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/items/?page=2&page_size=2", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Count   int                      `json:"count"`
		Results []map[string]interface{} `json:"results"`
	}
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Over 5 rows, page 2 with page_size 2 must return 2 rows
	assert.Equal(t, 5, resp.Count, "total count should be 5")
	assert.Equal(t, 2, len(resp.Results), "paginated request over 5 rows with page_size=2 must return 2 rows")

	// Queryset must receive limit 2 and offset 2
	assert.True(t, mgr.limitCalled, "limit must be called on queryset")
	assert.True(t, mgr.offsetCalled, "offset must be called on queryset")
	assert.Equal(t, 2, mgr.receivedLimit, "limit should be 2")
	assert.Equal(t, 2, mgr.receivedOffset, "offset should be 2")
}
