package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/require"
)

type sampleProduct struct {
	schema.BaseSchema
	ID         int64   `json:"id" db:"id"`
	Name       string  `json:"name" db:"name"`
	Price      float64 `json:"price" db:"price"`
	IsFeatured bool    `json:"is_featured" db:"is_featured"`
	Archived   bool    `json:"archived" db:"archived"`
}

func (sampleProduct) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
		schema.FloatField("price"),
		schema.BoolField("is_featured"),
		schema.BoolField("archived"),
	}
}

type sampleProductSerializer struct{ *BaseSerializer }

func newSampleProductSerializer() Serializer {
	return &sampleProductSerializer{NewBaseSerializer(nil)}
}

func (sampleProductSerializer) Fields() []string {
	return []string{"id", "name", "price", "is_featured", "archived"}
}

func sampleProducts(t *testing.T) (*orm.Manager[sampleProduct], *db.DB) {
	t.Helper()
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "products.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	_, err = database.Exec(`CREATE TABLE products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL, is_featured BOOLEAN NOT NULL DEFAULT 0, archived BOOLEAN NOT NULL DEFAULT 0)`)
	require.NoError(t, err)
	manager, err := orm.NewManagerWithDB[sampleProduct]("products", database)
	require.NoError(t, err)
	return manager, database
}

func sampleResponse(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestViewSetServesCRUDRoutes(t *testing.T) {
	manager, _ := sampleProducts(t)
	// The documented constructor, registration and route mounting sequence.
	viewset := NewBaseViewSet(newSampleProductSerializer, manager, &sampleProduct{})
	apiRouter := NewRouter("/api/v1")
	apiRouter.Register("products", viewset)
	router := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(router)

	created := sampleResponse(t, router, http.MethodPost, "/api/v1/products/", `{"name":"Desk","price":40}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	var product sampleProduct
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &product))
	require.Positive(t, product.ID)
	path := "/api/v1/products/" + strconv.FormatInt(product.ID, 10)
	require.Equal(t, http.StatusOK, sampleResponse(t, router, http.MethodGet, path, "").Code)
	listed := sampleResponse(t, router, http.MethodGet, "/api/v1/products/?page=1&page_size=1", "")
	require.Equal(t, http.StatusOK, listed.Code, listed.Body.String())
	var page struct {
		Count   int             `json:"count"`
		Results []sampleProduct `json:"results"`
	}
	require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &page))
	require.Equal(t, 1, page.Count)
	require.Len(t, page.Results, 1)
	require.Equal(t, http.StatusOK, sampleResponse(t, router, http.MethodPut, path, `{"name":"Chair","price":20,"is_featured":false,"archived":false}`).Code)
	require.Equal(t, http.StatusOK, sampleResponse(t, router, http.MethodPatch, path, `{"price":25}`).Code)
	stored, err := manager.Get(context.Background(), product.ID)
	require.NoError(t, err)
	require.Equal(t, "Chair", stored.Name)
	require.Equal(t, 25.0, stored.Price)
	require.Equal(t, http.StatusNoContent, sampleResponse(t, router, http.MethodDelete, path, "").Code)
	count, err := manager.Count(context.Background())
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestCustomPaginationSlicesDatabaseRows(t *testing.T) {
	manager, _ := sampleProducts(t)
	for _, name := range []string{"one", "two", "three"} {
		require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: name}))
	}
	router := forgehttp.NewRouter()
	router.Get("/products", func(w http.ResponseWriter, r *http.Request) {
		page, pageSize, offset := ParsePaginationParams(r, 20)
		rows, err := manager.QuerySet().OrderBy("id").Offset(offset).Limit(pageSize).All(r.Context())
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		count, err := manager.Count(r.Context())
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := SendPaginatedResponse(w, r, rows, int(count), page, pageSize); err != nil {
			t.Error(err)
		}
	})
	rec := sampleResponse(t, router, http.MethodGet, "/products?page=2&page_size=1&search=foo", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var response struct {
		Count    int             `json:"count"`
		Next     string          `json:"next"`
		Previous string          `json:"previous"`
		Results  []sampleProduct `json:"results"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, 3, response.Count)
	require.Len(t, response.Results, 1)
	require.Equal(t, "two", response.Results[0].Name)
	require.Contains(t, response.Next, "page=3")
	require.Contains(t, response.Previous, "page=1")
	require.Contains(t, response.Next, "search=foo")
}

func TestPageSizeSettingsCapRequests(t *testing.T) {
	before := GetSettings()
	t.Cleanup(func() { SetSettings(before) })
	settings := DefaultSettings()
	settings.PageSize = 50
	settings.MaxPageSize = 200
	SetSettings(settings)
	manager, _ := sampleProducts(t)
	for i := range 3 {
		require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: strconv.Itoa(i)}))
	}
	viewset := NewBaseViewSet(newSampleProductSerializer, manager, &sampleProduct{})
	apiRouter := NewRouter("/api/v1")
	apiRouter.Register("products", viewset)
	router := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(router)
	for _, path := range []string{"/api/v1/products/", "/api/v1/products/?page_size=500"} {
		rec := sampleResponse(t, router, http.MethodGet, path, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var page struct {
			Results []sampleProduct `json:"results"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
		require.Len(t, page.Results, 3)
	}
	_, size, _ := ParsePaginationParams(httptest.NewRequest(http.MethodGet, "/?page_size=500", nil), settings.PageSize)
	require.Equal(t, 200, size)
}

func TestProductActionsReadAndUpdateDatabase(t *testing.T) {
	manager, _ := sampleProducts(t)
	require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: "Desk", Price: 100, IsFeatured: true}))
	require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: "Chair", Price: 50}))
	apiRouter := NewRouter("/api/v1")
	apiRouter.Register("products", NewBaseViewSet(newSampleProductSerializer, manager, &sampleProduct{}))
	// Adapt the documented ProductObjects name to the SQLite manager above.
	apiRouter.Action("products", "featured", ActionConfig{}, func(w http.ResponseWriter, r *http.Request) {
		featured, err := manager.QuerySet().Filter(orm.F("is_featured").Eq(true)).Limit(6).All(r.Context())
		if err != nil {
			_ = forgehttp.SendError(w, http.StatusInternalServerError, "could not load products")
			return
		}
		_ = forgehttp.SendJSON(w, http.StatusOK, SerializeMany(featured))
	})
	apiRouter.Action("products", "apply_discount", ActionConfig{Methods: []string{http.MethodPost}, Detail: true, URLPath: "apply-discount"}, func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(forgehttp.URLParam(r, "id"), 10, 64)
		if err != nil {
			_ = forgehttp.SendError(w, http.StatusBadRequest, "invalid id")
			return
		}
		var body struct {
			Percent float64 `json:"percent"`
		}
		if err := forgehttp.GetJSON(r, &body); err != nil || body.Percent <= 0 || body.Percent >= 100 {
			_ = forgehttp.SendError(w, http.StatusBadRequest, "percent must be between 0 and 100")
			return
		}
		product, err := manager.Get(r.Context(), id)
		if err != nil {
			_ = forgehttp.SendError(w, http.StatusNotFound, "not found")
			return
		}
		product.Price = product.Price * (1 - body.Percent/100)
		if err := manager.Update(r.Context(), product); err != nil {
			_ = forgehttp.SendError(w, http.StatusInternalServerError, "could not save product")
			return
		}
		_ = forgehttp.SendJSON(w, http.StatusOK, SerializeModel(product))
	})
	router := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(router)
	featured := sampleResponse(t, router, http.MethodGet, "/api/v1/products/featured", "")
	require.Equal(t, http.StatusOK, featured.Code, featured.Body.String())
	var products []sampleProduct
	require.NoError(t, json.Unmarshal(featured.Body.Bytes(), &products))
	require.Len(t, products, 1)
	require.Equal(t, "Desk", products[0].Name)
	discounted := sampleResponse(t, router, http.MethodPost, "/api/v1/products/1/apply-discount", `{"percent":25}`)
	require.Equal(t, http.StatusOK, discounted.Code, discounted.Body.String())
	stored, err := manager.Get(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 75.0, stored.Price)
	require.Equal(t, http.StatusBadRequest, sampleResponse(t, router, http.MethodPost, "/api/v1/products/1/apply-discount", `{"percent":100}`).Code)
}

type activeProducts struct{ *orm.Manager[sampleProduct] }

func (m activeProducts) QuerySet() orm.QuerySet[sampleProduct] {
	return m.Manager.QuerySet().Filter(orm.F("archived").Eq(false))
}

func (m activeProducts) Get(ctx context.Context, id int64) (*sampleProduct, error) {
	return m.QuerySet().Filter(orm.F("id").Eq(id)).Get(ctx)
}

func TestScopedViewSetHidesArchivedProducts(t *testing.T) {
	manager, _ := sampleProducts(t)
	require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: "Visible"}))
	require.NoError(t, manager.Create(context.Background(), &sampleProduct{Name: "Hidden", Archived: true}))
	viewset := NewBaseViewSet(newSampleProductSerializer, activeProducts{manager}, &sampleProduct{})
	apiRouter := NewRouter("/api/v1")
	apiRouter.Register("products", viewset)
	router := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(router)
	list := sampleResponse(t, router, http.MethodGet, "/api/v1/products/", "")
	require.Equal(t, http.StatusOK, list.Code, list.Body.String())
	var page struct {
		Count   int             `json:"count"`
		Results []sampleProduct `json:"results"`
	}
	require.NoError(t, json.Unmarshal(list.Body.Bytes(), &page))
	require.Equal(t, 1, page.Count)
	require.Len(t, page.Results, 1)
	require.Equal(t, "Visible", page.Results[0].Name)
	require.Equal(t, http.StatusNotFound, sampleResponse(t, router, http.MethodGet, "/api/v1/products/2", "").Code)
}

type documentedPriceSerializer struct{ *BaseSerializer }

func newDocumentedPriceSerializer() Serializer {
	return &documentedPriceSerializer{NewBaseSerializer(nil)}
}
func (s *documentedPriceSerializer) New() Serializer { return newDocumentedPriceSerializer() }
func (documentedPriceSerializer) Fields() []string {
	return []string{"id", "sku", "name", "price", "tax_rate", "price_with_tax", "in_stock", "created_at"}
}
func (documentedPriceSerializer) ReadonlyFields() []string {
	return []string{"id", "price_with_tax", "created_at"}
}
func (s *documentedPriceSerializer) Validate() error {
	price := s.GetFloat("price")
	if price < 0.01 {
		s.AddError("price", "price must be at least $0.01")
		return errors.New("validation failed")
	}
	return s.BaseSerializer.Validate()
}

func TestSerializerValidatesPriceAndLimitsResponse(t *testing.T) {
	manager, _ := sampleProducts(t)
	viewset := NewBaseViewSet(newDocumentedPriceSerializer, manager, &sampleProduct{})
	apiRouter := NewRouter("/api/v1")
	apiRouter.Register("products", viewset)
	router := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(router)
	bad := sampleResponse(t, router, http.MethodPost, "/api/v1/products/", `{"name":"Desk","price":0}`)
	require.Equal(t, http.StatusBadRequest, bad.Code, bad.Body.String())
	require.Contains(t, bad.Body.String(), "price must be at least $0.01")
	good := sampleResponse(t, router, http.MethodPost, "/api/v1/products/", `{"name":"Desk","price":20,"id":999}`)
	require.Equal(t, http.StatusCreated, good.Code, good.Body.String())
	var response map[string]any
	require.NoError(t, json.Unmarshal(good.Body.Bytes(), &response))
	require.Equal(t, float64(1), response["id"])
	require.Equal(t, float64(20), response["price"])
	require.NotContains(t, response, "archived")
	count, err := manager.Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

type sampleCustomer struct {
	schema.BaseSchema
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func (sampleCustomer) Meta() schema.Meta { return schema.Meta{TableName: "sample_customers"} }

func (sampleCustomer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type sampleOrder struct {
	schema.BaseSchema
	ID         int64           `json:"id" db:"id"`
	CustomerID int64           `json:"customer_id" db:"customer_id"`
	Customer   *sampleCustomer `json:"customer" db:"customer"`
}

func (sampleOrder) Meta() schema.Meta { return schema.Meta{TableName: "sample_orders"} }

func (sampleOrder) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id"),
	}
}

func (sampleOrder) Relations() []schema.Relation {
	return []schema.Relation{schema.ForeignKeyField("Customer", "sampleCustomer", schema.OnDelete(schema.CascadeCASCADE))}
}

type sampleOrderSerializer struct{ *BaseSerializer }

func TestPopulatedRelationSerializesAsNestedObject(t *testing.T) {
	database, err := db.NewDBWithDriver("sqlite3", filepath.Join(t.TempDir(), "orders.sqlite"), db.WithMaxOpenConns(1))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, database.Close()) })
	_, err = database.Exec(`CREATE TABLE sample_customers (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE sample_orders (id INTEGER PRIMARY KEY, customer_id INTEGER REFERENCES sample_customers(id));
		INSERT INTO sample_customers (id, name) VALUES (1, 'Ada');
		INSERT INTO sample_orders (id, customer_id) VALUES (7, 1);`)
	require.NoError(t, err)
	_, err = orm.GetModelSchema[sampleCustomer]()
	require.NoError(t, err)
	manager, err := orm.NewManagerWithDB[sampleOrder]("sample_orders", database)
	require.NoError(t, err)
	router := forgehttp.NewRouter()
	router.Get("/orders/7", func(w http.ResponseWriter, r *http.Request) {
		orders, err := manager.QuerySet().SelectRelated("Customer").All(r.Context())
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(orders) != 1 || orders[0].Customer == nil {
			t.Error("SelectRelated did not populate Customer")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		serializer := &sampleOrderSerializer{BaseSerializer: NewBaseSerializer(nil)}
		_ = forgehttp.SendJSON(w, http.StatusOK, filterOutputMap(serializer, SerializeModel(orders[0])))
	})
	rec := sampleResponse(t, router, http.MethodGet, "/orders/7", "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body struct {
		Customer map[string]any `json:"customer"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "Ada", body.Customer["name"])
	require.Equal(t, float64(1), body.Customer["id"])
}
