package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

type populateInner struct {
	City string `json:"city"`
}

type populateTestModel struct {
	Price    float64                `json:"price"`
	Rating   float32                `json:"rating"`
	Quantity int                    `json:"quantity"`
	Data     []byte                 `json:"data"`
	Meta     map[string]interface{} `json:"meta"`
	Tags     []string               `json:"tags"`
	Addr     populateInner          `json:"address"`
}

func TestPopulateFromMap_ConvertsFloatDecimalJSONBytes(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte("hello"))

	m := &populateTestModel{}
	err := populateFromMap(m, map[string]interface{}{
		"price":   float64(12.5),
		"rating":  float64(4.5),
		"data":    raw,
		"meta":    map[string]interface{}{"a": float64(1)},
		"tags":    []interface{}{"a", "b"},
		"address": map[string]interface{}{"city": "Berlin"},
	})
	require.NoError(t, err)
	assert.Equal(t, 12.5, m.Price)
	assert.Equal(t, float32(4.5), m.Rating)
	assert.Equal(t, []byte("hello"), m.Data)
	assert.Equal(t, map[string]interface{}{"a": float64(1)}, m.Meta)
	assert.Equal(t, []string{"a", "b"}, m.Tags)
	assert.Equal(t, "Berlin", m.Addr.City)

	// Numeric strings decode into float fields as well.
	m2 := &populateTestModel{}
	require.NoError(t, populateFromMap(m2, map[string]interface{}{"price": "12.5"}))
	assert.Equal(t, 12.5, m2.Price)
}

func TestPopulateFromMap_RejectsUnconvertibleValue(t *testing.T) {
	m := &populateTestModel{}
	err := populateFromMap(m, map[string]interface{}{"price": "abc"})
	require.Error(t, err)
	var fe *fieldError
	require.ErrorAs(t, err, &fe)
	assert.Equal(t, "price", fe.Field)
}

func TestPopulateFromMap_RejectsNonIntegralIntegerRepresentations(t *testing.T) {
	type integerModel struct {
		Signed   int64  `json:"signed"`
		Unsigned uint64 `json:"unsigned"`
	}
	tests := []struct {
		name  string
		field string
		value interface{}
	}{
		{name: "signed numeric string", field: "signed", value: "1.9"},
		{name: "signed json number", field: "signed", value: json.Number("1.9")},
		{name: "unsigned float", field: "unsigned", value: 1.9},
		{name: "unsigned numeric string", field: "unsigned", value: "1.9"},
		{name: "unsigned json number", field: "unsigned", value: json.Number("1.9")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := populateFromMap(&integerModel{}, map[string]interface{}{tt.field: tt.value})
			require.Error(t, err)
			var fieldErr *fieldError
			require.ErrorAs(t, err, &fieldErr)
			assert.Equal(t, tt.field, fieldErr.Field)
		})
	}
}

type priceTestItem struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Metadata []byte  `json:"metadata"`
}

type priceTestSerializer struct {
	*BaseSerializer
}

func newPriceTestSerializer() Serializer {
	return &priceTestSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *priceTestSerializer) Fields() []string {
	return []string{"id", "title", "price", "quantity", "metadata"}
}

type priceTestManager struct {
	mu          sync.Mutex
	items       map[int64]*priceTestItem
	createCalls int
	created     *priceTestItem
}

func (m *priceTestManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls++
	item, ok := model.(*priceTestItem)
	if !ok {
		return errors.New("unexpected model type")
	}
	cp := *item
	m.created = &cp
	return nil
}

func (m *priceTestManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[id]
	if !ok {
		return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
	}
	cp := *item
	return &cp, nil
}

func (m *priceTestManager) Update(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := model.(*priceTestItem)
	if !ok {
		return errors.New("unexpected model type")
	}
	m.items[item.ID] = item
	return nil
}

func (m *priceTestManager) Delete(ctx context.Context, model interface{}) error {
	return nil
}

func TestBaseViewSet_Create_InvalidFieldValueReturns400(t *testing.T) {
	mgr := &priceTestManager{items: map[int64]*priceTestItem{}}
	vs := NewBaseViewSet(newPriceTestSerializer, mgr, &priceTestItem{})

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"price":"abc"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, 0, mgr.createCalls)
}

func TestBaseViewSet_Create_RejectsFractionalIntegerWithoutCallingManager(t *testing.T) {
	mgr := &priceTestManager{items: map[int64]*priceTestItem{}}
	handler := newPriceTestHandler(mgr)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"quantity":1.9}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "quantity")
	assert.Equal(t, 0, mgr.createCalls)
}

func TestBaseViewSet_Create_AcceptsWholeNumberFloatForInteger(t *testing.T) {
	mgr := &priceTestManager{items: map[int64]*priceTestItem{}}
	handler := newPriceTestHandler(mgr)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"quantity":2.0}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.NotNil(t, mgr.created)
	assert.Equal(t, 2, mgr.created.Quantity)
}

func TestBaseViewSet_Create_AcceptsJSONAndBase64ForBytes(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []byte
	}{
		{name: "object", body: `{"metadata":{"a":1}}`, want: []byte(`{"a":1}`)},
		{name: "array", body: `{"metadata":["a",1]}`, want: []byte(`["a",1]`)},
		{name: "base64", body: `{"metadata":"aGVsbG8="}`, want: []byte("hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &priceTestManager{items: map[int64]*priceTestItem{}}
			handler := newPriceTestHandler(mgr)
			req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusCreated, rec.Code)
			require.NotNil(t, mgr.created)
			assert.Equal(t, tt.want, mgr.created.Metadata)
		})
	}
}

func TestBaseViewSet_Create_RejectsInvalidBase64ForBytes(t *testing.T) {
	mgr := &priceTestManager{items: map[int64]*priceTestItem{}}
	handler := newPriceTestHandler(mgr)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"metadata":"%%%"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "metadata")
	assert.Equal(t, 0, mgr.createCalls)
}

func newPriceTestHandler(mgr *priceTestManager) *forgehttp.Router {
	vs := NewBaseViewSet(newPriceTestSerializer, mgr, &priceTestItem{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

func TestBaseViewSet_Patch_ConvertsFloat(t *testing.T) {
	mgr := &priceTestManager{
		items: map[int64]*priceTestItem{
			1: {ID: 1, Title: "widget", Price: 1.0},
		},
	}
	vs := NewBaseViewSet(newPriceTestSerializer, mgr, &priceTestItem{})

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodPatch, "/api/items/1", bytes.NewBufferString(`{"price": 12.5}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	require.NotNil(t, mgr.items[1])
	assert.Equal(t, 12.5, mgr.items[1].Price)
}
