package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type collBoundsItem struct {
	schema.BaseSchema
	ID       int64  `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Category string `json:"category" db:"category"`
}

func (collBoundsItem) Fields() []schema.Field {
	return []schema.Field{
		{Name: "id", Type: schema.TypeInt64, PrimaryKey: true, DBColumn: "id", Serialize: true},
		{Name: "name", Type: schema.TypeString, DBColumn: "name", Serialize: true},
		{Name: "category", Type: schema.TypeString, DBColumn: "category", Serialize: true},
	}
}

func (collBoundsItem) Meta() schema.Meta {
	return schema.Meta{TableName: "coll_bounds_items"}
}

type collBoundsSerializer struct {
	*BaseSerializer
}

func newCollBoundsSerializer() Serializer {
	return &collBoundsSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *collBoundsSerializer) Fields() []string {
	return []string{"id", "name", "category"}
}

type collBoundsQuerySet struct {
	items         []*collBoundsItem
	limit         int
	offset        int
	orderedFields []interface{}
	filterCalls   int
}

func (qs *collBoundsQuerySet) Count(ctx context.Context) (int64, error) {
	return int64(len(qs.items)), nil
}

func (qs *collBoundsQuerySet) Limit(n int) interface{} {
	qs.limit = n
	return qs
}

func (qs *collBoundsQuerySet) Offset(n int) interface{} {
	qs.offset = n
	return qs
}

func (qs *collBoundsQuerySet) OrderBy(fields ...interface{}) interface{} {
	qs.orderedFields = append([]interface{}(nil), fields...)
	return qs
}

func (qs *collBoundsQuerySet) Filter(expr interface{}) interface{} {
	qs.filterCalls++
	if e, ok := expr.(orm.Expression); ok {
		sql, args, err := e.ToSQL(orm.NewSQLBuilder())
		if err == nil && len(args) > 0 {
			val, _ := args[0].(string)
			if strings.Contains(sql, "category") {
				filtered := make([]*collBoundsItem, 0)
				for _, it := range qs.items {
					if it.Category == val {
						filtered = append(filtered, it)
					}
				}
				qs.items = filtered
			}
		}
	}
	return qs
}

func (qs *collBoundsQuerySet) All(ctx context.Context) ([]*collBoundsItem, error) {
	start := qs.offset
	if start > len(qs.items) {
		start = len(qs.items)
	}
	end := start + qs.limit
	if qs.limit <= 0 || end > len(qs.items) {
		end = len(qs.items)
	}
	return qs.items[start:end], nil
}

func (qs *collBoundsQuerySet) Get(ctx context.Context, id int64) (interface{}, error) {
	for _, it := range qs.items {
		if it.ID == id {
			copy := *it
			return &copy, nil
		}
	}
	return nil, nil
}

func (qs *collBoundsQuerySet) Create(ctx context.Context, model interface{}) error {
	return nil
}

func (qs *collBoundsQuerySet) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (qs *collBoundsQuerySet) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func setupCollBoundsHandler(qs *collBoundsQuerySet) http.Handler {
	vs := NewBaseViewSet(newCollBoundsSerializer, qs, &collBoundsItem{})
	router := NewRouter("/api")
	router.Register("items", vs)
	h := forgehttp.NewRouter()
	router.RegisterRoutes(h)
	return h
}

func TestCollectionBounds(t *testing.T) {
	// Item 1: Unknown filter field or unknown operator
	t.Run("1_UnknownFilterOrOperator", func(t *testing.T) {
		qs := &collBoundsQuerySet{}
		handler := setupCollBoundsHandler(qs)

		// 1a. Unknown filter field -> 400 with "invalid filter"
		req := httptest.NewRequest(http.MethodGet, "/api/items/?nosuch=1", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid filter")

		// 1b. Unknown filter operator -> 400 with "invalid filter"
		req2 := httptest.NewRequest(http.MethodGet, "/api/items/?name__bogus=x", nil)
		rec2 := httptest.NewRecorder()
		handler.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusBadRequest, rec2.Code)
		assert.Contains(t, rec2.Body.String(), "invalid filter")
	})

	// Item 2: Filter complexity limit
	t.Run("2_FilterComplexityLimit", func(t *testing.T) {
		assert.Equal(t, 20, MaxFilterConditions, "MaxFilterConditions named constant must be 20")
		assert.Equal(t, "filter_too_complex", CodeFilterTooComplex)

		qs := &collBoundsQuerySet{}
		handler := setupCollBoundsHandler(qs)

		// More than 20 filter conditions (21 conditions) -> 400 "filter_too_complex"
		params := make([]string, 21)
		for i := 0; i < 21; i++ {
			params[i] = fmt.Sprintf("cond_%d=%d", i, i)
		}
		url := "/api/items/?" + strings.Join(params, "&")
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var body map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, CodeFilterTooComplex, body["code"])
	})

	// Item 3: page_size above the configured maximum clamped to maximum
	t.Run("3_PageSizeClampedToMaximum", func(t *testing.T) {
		// Test ParsePaginationParams directly
		r := httptest.NewRequest(http.MethodGet, "/api/items/?page_size=500", nil)
		_, pageSize, _ := ParsePaginationParams(r, 20)
		assert.Equal(t, 100, pageSize, "page_size above maximum (100) must be clamped to 100")

		// Test through ViewSet list endpoint
		qs := &collBoundsQuerySet{}
		handler := setupCollBoundsHandler(qs)
		req := httptest.NewRequest(http.MethodGet, "/api/items/?page_size=500", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 100, qs.limit, "queryset limit must be clamped to maxPageSize (100)")
	})

	// Item 4: Deterministic ordering tiebreak
	t.Run("4_DeterministicOrderingTiebreak", func(t *testing.T) {
		// When requested ordering does not include primary key ("id"), append "id" as tiebreak
		qs := &collBoundsQuerySet{}
		handler := setupCollBoundsHandler(qs)

		req := httptest.NewRequest(http.MethodGet, "/api/items/?ordering=name", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []interface{}{"name", "id"}, qs.orderedFields, "primary key must be appended as tiebreak")

		// When requested ordering already includes primary key, do not duplicate
		qs2 := &collBoundsQuerySet{}
		handler2 := setupCollBoundsHandler(qs2)

		req2 := httptest.NewRequest(http.MethodGet, "/api/items/?ordering=-name,id", nil)
		rec2 := httptest.NewRecorder()
		handler2.ServeHTTP(rec2, req2)

		assert.Equal(t, http.StatusOK, rec2.Code)
		assert.Equal(t, []interface{}{"-name", "id"}, qs2.orderedFields, "primary key should not be appended again")

		// When requested ordering is descending primary key only
		qs3 := &collBoundsQuerySet{}
		handler3 := setupCollBoundsHandler(qs3)

		req3 := httptest.NewRequest(http.MethodGet, "/api/items/?ordering=-id", nil)
		rec3 := httptest.NewRecorder()
		handler3.ServeHTTP(rec3, req3)

		assert.Equal(t, http.StatusOK, rec3.Code)
		assert.Equal(t, []interface{}{"-id"}, qs3.orderedFields, "primary key should not be appended again")
	})

	// Item 5: Count is computed on the scoped/filtered queryset
	t.Run("5_CountComputedOnFilteredQueryset", func(t *testing.T) {
		items := []*collBoundsItem{
			{ID: 1, Name: "Item 1", Category: "books"},
			{ID: 2, Name: "Item 2", Category: "books"},
			{ID: 3, Name: "Item 3", Category: "electronics"},
			{ID: 4, Name: "Item 4", Category: "books"},
			{ID: 5, Name: "Item 5", Category: "electronics"},
		}
		qs := &collBoundsQuerySet{items: items}
		handler := setupCollBoundsHandler(qs)

		// Filter for category=books: 3 items match out of 5 total
		req := httptest.NewRequest(http.MethodGet, "/api/items/?category=books", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp struct {
			Count   int                      `json:"count"`
			Results []map[string]interface{} `json:"results"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, 3, resp.Count, "count must describe scoped/filtered queryset, not total table rows")
		assert.Equal(t, 3, len(resp.Results), "results must match filtered queryset")
	})

	// Item 6: Request body size limit for create/update
	t.Run("6_RequestBodySizeLimit", func(t *testing.T) {
		assert.Equal(t, int64(1<<20), int64(MaxRequestBodySize), "MaxRequestBodySize named constant must be 1 MiB")

		qs := &collBoundsQuerySet{
			items: []*collBoundsItem{
				{ID: 1, Name: "Existing Item", Category: "books"},
			},
		}
		handler := setupCollBoundsHandler(qs)

		// Over limit: 1 MiB + 10 bytes
		largePayload := fmt.Sprintf(`{"name": "%s"}`, strings.Repeat("a", (1<<20)+10))

		// 6a. POST (Create) with body > 1 MiB -> 413
		reqPost := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(largePayload))
		reqPost.Header.Set("Content-Type", "application/json")
		recPost := httptest.NewRecorder()
		handler.ServeHTTP(recPost, reqPost)
		assert.Equal(t, http.StatusRequestEntityTooLarge, recPost.Code, "POST body over limit must return 413")

		// 6b. PUT (Update) with body > 1 MiB -> 413
		reqPut := httptest.NewRequest(http.MethodPut, "/api/items/1", bytes.NewBufferString(largePayload))
		reqPut.Header.Set("Content-Type", "application/json")
		recPut := httptest.NewRecorder()
		handler.ServeHTTP(recPut, reqPut)
		assert.Equal(t, http.StatusRequestEntityTooLarge, recPut.Code, "PUT body over limit must return 413")

		// 6c. PATCH (PartialUpdate) with body > 1 MiB -> 413
		reqPatch := httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(largePayload))
		reqPatch.Header.Set("Content-Type", "application/json")
		recPatch := httptest.NewRecorder()
		handler.ServeHTTP(recPatch, reqPatch)
		assert.Equal(t, http.StatusRequestEntityTooLarge, recPatch.Code, "PATCH body over limit must return 413")

		// 6d. POST with body <= 1 MiB -> does not return 413
		normalPayload := `{"name": "Valid Item", "category": "books"}`
		reqValid := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(normalPayload))
		reqValid.Header.Set("Content-Type", "application/json")
		recValid := httptest.NewRecorder()
		handler.ServeHTTP(recValid, reqValid)
		assert.NotEqual(t, http.StatusRequestEntityTooLarge, recValid.Code)
		assert.Equal(t, http.StatusCreated, recValid.Code)
	})
}
