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
	validate "github.com/forgego/forge/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type persistenceTestItem struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type persistenceTestSerializer struct {
	*BaseSerializer
}

func newPersistenceTestSerializer() Serializer {
	return &persistenceTestSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *persistenceTestSerializer) Fields() []string {
	return []string{"id", "title"}
}

type persistenceTestManager struct {
	mu          sync.Mutex
	items       map[int64]*persistenceTestItem
	createErr   error
	deleteErr   error
	getErr      error
	createCalls int
}

func (m *persistenceTestManager) Create(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls++
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}

func (m *persistenceTestManager) Get(ctx context.Context, id int64) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getErr != nil {
		return nil, m.getErr
	}
	if item, ok := m.items[id]; ok {
		cp := *item
		return &cp, nil
	}
	return nil, forgeerrors.NewNotFoundErrorWithMessage("not found")
}

func (m *persistenceTestManager) Update(ctx context.Context, model interface{}) error {
	return nil
}

func (m *persistenceTestManager) Delete(ctx context.Context, model interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}

type persistenceErrorQuerySet struct {
	err error
}

func (qs *persistenceErrorQuerySet) Count(ctx context.Context) (int64, error) {
	return 0, qs.err
}

func (qs *persistenceErrorQuerySet) All(ctx context.Context) (interface{}, error) {
	return nil, qs.err
}

func (qs *persistenceErrorQuerySet) Filter(expr interface{}) interface{} { return qs }

func (qs *persistenceErrorQuerySet) OrderBy(fields ...interface{}) interface{} { return qs }

func (qs *persistenceErrorQuerySet) Limit(limit int) interface{} { return qs }

func (qs *persistenceErrorQuerySet) Offset(offset int) interface{} { return qs }

func newPersistenceRouter(mgr *persistenceTestManager) *forgehttp.Router {
	vs := NewBaseViewSet(newPersistenceTestSerializer, mgr, &persistenceTestItem{})
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	return handler
}

func TestBaseViewSet_PersistenceError_DoesNotLeakMessage(t *testing.T) {
	secret := errors.New("pq: duplicate key value violates unique constraint secret_idx")

	// Create failure.
	mgr := &persistenceTestManager{
		items:     map[int64]*persistenceTestItem{},
		createErr: secret,
	}
	handler := newPersistenceRouter(mgr)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret_idx")
	assert.NotContains(t, rec.Body.String(), "pq:")

	// Destroy failure.
	mgr = &persistenceTestManager{
		items:     map[int64]*persistenceTestItem{1: {ID: 1, Title: "x"}},
		deleteErr: secret,
	}
	handler = newPersistenceRouter(mgr)
	req = httptest.NewRequest(http.MethodDelete, "/api/items/1", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret_idx")
	assert.NotContains(t, rec.Body.String(), "pq:")

	// List queryset failure.
	vs := NewBaseViewSet(
		newPersistenceTestSerializer,
		&persistenceErrorQuerySet{err: secret},
		&persistenceTestItem{},
	)
	router := NewRouter("/api")
	router.Register("items", vs)
	listHandler := forgehttp.NewRouter()
	router.RegisterRoutes(listHandler)
	req = httptest.NewRequest(http.MethodGet, "/api/items/", nil)
	rec = httptest.NewRecorder()
	listHandler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "secret_idx")
	assert.NotContains(t, rec.Body.String(), "pq:")
}

func TestBaseViewSet_PersistenceError_UsesErrorWriter(t *testing.T) {
	mgr := &persistenceTestManager{
		items:     map[int64]*persistenceTestItem{},
		createErr: errors.New("boom"),
	}
	vs := NewBaseViewSet(newPersistenceTestSerializer, mgr, &persistenceTestItem{})
	var called bool
	var captured error
	vs.ErrorWriter = func(w http.ResponseWriter, r *http.Request, err error) {
		called = true
		captured = err
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{}`))
	}

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.True(t, called)
	require.NotNil(t, captured)
	assert.Equal(t, http.StatusTeapot, rec.Code)
}

func TestBaseViewSet_InvalidInputError_Returns400(t *testing.T) {
	mgr := &persistenceTestManager{
		items:     map[int64]*persistenceTestItem{},
		createErr: forgeerrors.NewInvalidInputError("title", "is required"),
	}
	handler := newPersistenceRouter(mgr)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "title")
}

func TestBaseViewSet_GeneratedModelValidationError_Returns400(t *testing.T) {
	type generatedModel struct {
		Title string `validate:"required"`
	}
	validationErr := validate.ValidateModel(validate.NewValidator(), &generatedModel{})
	require.Error(t, validationErr)

	mgr := &persistenceTestManager{
		items:     map[int64]*persistenceTestItem{},
		createErr: validationErr,
	}
	handler := newPersistenceRouter(mgr)
	req := httptest.NewRequest(http.MethodPost, "/api/items/", bytes.NewBufferString(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Title")
	assert.Contains(t, rec.Body.String(), "is required")
}

func TestBaseViewSet_LookupOperationalError_Returns500WithoutDetail(t *testing.T) {
	mgr := &persistenceTestManager{
		items:  map[int64]*persistenceTestItem{},
		getErr: errors.New("sql: connection refused at 10.0.0.5"),
	}
	handler := newPersistenceRouter(mgr)
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			var body *bytes.Buffer
			if method == http.MethodPut || method == http.MethodPatch {
				body = bytes.NewBufferString(`{"title":"x"}`)
			} else {
				body = bytes.NewBuffer(nil)
			}
			req := httptest.NewRequest(method, "/api/items/1", body)
			if body.Len() > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusInternalServerError, rec.Code)
			assert.NotContains(t, rec.Body.String(), "10.0.0.5")
		})
	}
}

func TestBaseViewSet_LookupNotFound_Returns404(t *testing.T) {
	mgr := &persistenceTestManager{
		items:  map[int64]*persistenceTestItem{},
		getErr: forgeerrors.NewNotFoundErrorWithMessage("items matching query does not exist"),
	}
	handler := newPersistenceRouter(mgr)
	for _, method := range []string{http.MethodGet, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			var body *bytes.Buffer
			if method == http.MethodPatch {
				body = bytes.NewBufferString(`{"title":"x"}`)
			} else {
				body = bytes.NewBuffer(nil)
			}
			req := httptest.NewRequest(method, "/api/items/1", body)
			if body.Len() > 0 {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			require.Equal(t, http.StatusNotFound, rec.Code)
			assert.Contains(t, rec.Body.String(), "Not found")
		})
	}
}
