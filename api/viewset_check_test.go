package api

import (
	"context"
	"fmt"
	"io"
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

type checkedModel struct {
	schema.BaseSchema
	ID    int64  `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
}

func (checkedModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
	}
}

func newCheckSerializer() Serializer { return NewBaseSerializer(nil) }

func newCheckManager(t *testing.T) *orm.Manager[checkedModel] {
	t.Helper()
	manager, err := orm.NewManager[checkedModel]("check_models")
	require.NoError(t, err)
	return manager
}

// checkSecretDSN stands in for a secret held by a configured queryset; check
// messages must never format it.
const checkSecretDSN = "postgres://app:hunter2@db.internal/app"

// checkReadQueryset has only the read operations, plus a secret-bearing field.
type checkReadQueryset struct {
	DSN   string
	items map[int64]*checkedModel
}

func (q *checkReadQueryset) Count(context.Context) (int64, error) { return int64(len(q.items)), nil }
func (q *checkReadQueryset) All(context.Context) ([]*checkedModel, error) {
	out := make([]*checkedModel, 0, len(q.items))
	for _, item := range q.items {
		out = append(out, item)
	}
	return out, nil
}
func (q *checkReadQueryset) Get(_ context.Context, id int64) (*checkedModel, error) {
	if item, ok := q.items[id]; ok {
		cp := *item
		return &cp, nil
	}
	return nil, fmt.Errorf("not found")
}

// checkWritableQueryset adds the write operations and counts their calls.
type checkWritableQueryset struct {
	checkReadQueryset
	writes int
}

func (q *checkWritableQueryset) Create(context.Context, *checkedModel) error { q.writes++; return nil }
func (q *checkWritableQueryset) Update(context.Context, *checkedModel) error { q.writes++; return nil }
func (q *checkWritableQueryset) Delete(context.Context, *checkedModel) error { q.writes++; return nil }

type checkBadGet struct{ checkWritableQueryset }

func (*checkBadGet) Get(context.Context, string) (*checkedModel, error) { return nil, nil }

type checkValueGet struct{ checkWritableQueryset }

func (*checkValueGet) Get(context.Context, int64) (checkedModel, error) { return checkedModel{}, nil }

type checkBadCreate struct{ checkWritableQueryset }

func (*checkBadCreate) Create(context.Context, *checkedModel) (int64, error) { return 0, nil }

type otherCheckedModel struct{ ID int64 }

type checkWrongUpdateType struct{ checkWritableQueryset }

func (*checkWrongUpdateType) Update(context.Context, *otherCheckedModel) error { return nil }

type checkConcreteContext struct{ checkWritableQueryset }

func (*checkConcreteContext) Delete(*http.Request, *checkedModel) error { return nil }

type checkBadCount struct{ checkWritableQueryset }

func (*checkBadCount) Count(context.Context) int { return 0 }

type checkBadOrderBy struct{ checkWritableQueryset }

func (q *checkBadOrderBy) OrderBy(field string) *checkBadOrderBy { return q }

type checkBadFilter struct{ checkWritableQueryset }

func (q *checkBadFilter) Filter(expr string) *checkBadFilter { return q }

func newCheckViewSet(queryset interface{}) *BaseViewSet {
	return NewBaseViewSet(newCheckSerializer, queryset, &checkedModel{})
}

func registerPanic(resource string, vs ViewSet) (message string) {
	defer func() {
		if r := recover(); r != nil {
			message = fmt.Sprint(r)
		}
	}()
	NewRouter("/api").Register(resource, vs)
	return ""
}

func TestRouterRegister_AcceptsOrmManagerAndGeneratedViewSetShape(t *testing.T) {
	manager := newCheckManager(t)

	writable := newCheckViewSet(manager)
	assert.NoError(t, writable.CheckConfiguration())

	readOnly := newCheckViewSet(manager)
	readOnly.ReadOnly = true
	assert.NoError(t, readOnly.CheckConfiguration())

	// Generated viewsets embed *api.BaseViewSet (codegen/templates/api.tmpl),
	// so the check is promoted to them.
	type generatedViewSet struct{ *BaseViewSet }
	generated := &generatedViewSet{BaseViewSet: newCheckViewSet(manager)}
	generated.ExcludeResponseFields = NonSerializableFields(&checkedModel{})
	generated.ReadOnlyRequestFields = NonEditableFields(&checkedModel{})
	assert.Empty(t, registerPanic("check-models", generated))

	config := &ViewSetConfig{Model: &checkedModel{}, Queryset: manager, Serializer: NewBaseSerializer(nil)}
	assert.Empty(t, registerPanic("check-models", config))
}

func TestRouterRegister_RejectsMisconfiguredViewSets(t *testing.T) {
	var nilManager *orm.Manager[checkedModel]
	var nilModel *checkedModel
	readable := &checkReadQueryset{DSN: checkSecretDSN}

	for _, tc := range []struct {
		name string
		vs   ViewSet
		want []string
	}{
		{"nil viewset", nil, []string{"viewset is nil"}},
		{"typed nil BaseViewSet", (*BaseViewSet)(nil), []string{"viewset is a nil *api.BaseViewSet"}},
		{"nil serializer factory", NewBaseViewSet(nil, newCheckManager(t), &checkedModel{}), []string{"Serializer factory is nil"}},
		{"serializer factory returns nil", NewBaseViewSet(func() Serializer { return nil }, newCheckManager(t), &checkedModel{}), []string{"Serializer factory returned nil"}},
		{"serializer factory returns typed nil", NewBaseViewSet(func() Serializer { return (*BaseSerializer)(nil) }, newCheckManager(t), &checkedModel{}), []string{"Serializer factory returned nil"}},
		{"serializer factory panics", NewBaseViewSet(func() Serializer { panic(checkSecretDSN) }, newCheckManager(t), &checkedModel{}), []string{"Serializer factory panicked"}},
		{"nil model", NewBaseViewSet(newCheckSerializer, newCheckManager(t), nil), []string{"Model is nil"}},
		{"typed nil model", NewBaseViewSet(newCheckSerializer, newCheckManager(t), nilModel), []string{"Model is a nil *api.checkedModel"}},
		{"non-pointer model", NewBaseViewSet(newCheckSerializer, newCheckManager(t), checkedModel{}), []string{"Model must be a pointer to a struct, got api.checkedModel"}},
		{"nil queryset", newCheckViewSet(nil), []string{"Queryset is nil"}},
		{"typed nil manager", newCheckViewSet(nilManager), []string{"Queryset is a nil *orm.Manager[github.com/forgego/forge/api.checkedModel]"}},
		{"read-only queryset on writable viewset", newCheckViewSet(readable), []string{
			"missing Create(context.Context, *Model) error, needed by create",
			"missing Update(context.Context, *Model) error, needed by update and partial_update",
			"missing Delete(context.Context, *Model) error, needed by destroy",
			"set ReadOnly to serve only list and retrieve",
		}},
		{"Get with wrong id type", newCheckViewSet(&checkBadGet{}), []string{"Get has signature Get(context.Context, string) (*api.checkedModel, error)"}},
		{"Get returning a value", newCheckViewSet(&checkValueGet{}), []string{"Get has signature Get(context.Context, int64) (api.checkedModel, error)"}},
		{"Create with extra result", newCheckViewSet(&checkBadCreate{}), []string{"Create has signature Create(context.Context, *api.checkedModel) (int64, error); create needs Create(context.Context, *Model) error"}},
		{"Update for another model", newCheckViewSet(&checkWrongUpdateType{}), []string{"Update takes *api.otherCheckedModel but receives *api.checkedModel"}},
		{"Delete without context", newCheckViewSet(&checkConcreteContext{}), []string{"Delete has signature Delete(*http.Request, *api.checkedModel) error"}},
		{"Count without error", newCheckViewSet(&checkBadCount{}), []string{"Count has signature Count(context.Context) int; list needs Count(context.Context) (int64, error)"}},
		{"non-variadic OrderBy", newCheckViewSet(&checkBadOrderBy{}), []string{"OrderBy has signature OrderBy(string) *api.checkBadOrderBy; list ordering needs OrderBy(...string) T"}},
		{"Filter without expressions", newCheckViewSet(&checkBadFilter{}), []string{"Filter has signature Filter(string) *api.checkBadFilter"}},
		{"ViewSetConfig without serializer", &ViewSetConfig{Model: &checkedModel{}, Queryset: newCheckManager(t)}, []string{"Serializer is nil"}},
		{"ViewSetConfig without queryset", &ViewSetConfig{Model: &checkedModel{}, Serializer: NewBaseSerializer(nil)}, []string{"Queryset is nil"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			message := registerPanic("widgets", tc.vs)
			require.NotEmpty(t, message, "Register must panic")
			assert.Contains(t, message, `api: resource "widgets": `)
			for _, want := range tc.want {
				assert.Contains(t, message, want)
			}
			assert.NotContains(t, message, "hunter2", "messages must not format configured values")
		})
	}
}

func TestRouterRegister_ReportsEveryProblemAtOnce(t *testing.T) {
	message := registerPanic("widgets", NewBaseViewSet(nil, nil, nil))
	assert.Contains(t, message, "Serializer factory is nil")
	assert.Contains(t, message, "Model is nil")
	assert.Contains(t, message, "Queryset is nil")
}

func TestRouterRegister_ReadOnlyNeedsOnlyReadOperations(t *testing.T) {
	queryset := &checkWritableQueryset{checkReadQueryset: checkReadQueryset{items: map[int64]*checkedModel{1: {ID: 1, Title: "one"}}}}
	readOnlyQueryset := &checkReadQueryset{items: queryset.items}

	vs := newCheckViewSet(readOnlyQueryset)
	vs.ReadOnly = true
	require.NoError(t, vs.CheckConfiguration())

	missingGet := newCheckViewSet(&unusedListOperations{})
	missingGet.ReadOnly = true
	assert.Contains(t, registerPanic("widgets", missingGet), "missing Get(context.Context, int64) (*Model, error), needed by retrieve")

	// A writable queryset behind a ReadOnly viewset is never written.
	vs = newCheckViewSet(queryset)
	vs.ReadOnly = true
	router := NewRouter("/api")
	router.Register("widgets", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	for _, path := range []string{"/api/widgets/", "/api/widgets/1"} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusOK, rec.Code, path)
	}
	for _, req := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/api/widgets/", jsonBody(`{"title":"x"}`)),
		httptest.NewRequest(http.MethodPut, "/api/widgets/1", jsonBody(`{"title":"x"}`)),
		httptest.NewRequest(http.MethodPatch, "/api/widgets/1", jsonBody(`{"title":"x"}`)),
		httptest.NewRequest(http.MethodDelete, "/api/widgets/1", nil),
	} {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, req.Method)
		assert.Equal(t, "GET", rec.Header().Get("Allow"), req.Method)
	}
	assert.Zero(t, queryset.writes)
	assert.Equal(t, "one", queryset.items[1].Title)
}

// checkCustomViewSet implements ViewSet without BaseViewSet.
type checkCustomViewSet struct{}

func (checkCustomViewSet) List(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}
func (checkCustomViewSet) Create(http.ResponseWriter, *http.Request)   {}
func (checkCustomViewSet) Retrieve(http.ResponseWriter, *http.Request) {}
func (checkCustomViewSet) Update(http.ResponseWriter, *http.Request)   {}
func (checkCustomViewSet) PartialUpdate(http.ResponseWriter, *http.Request) {
}
func (checkCustomViewSet) Destroy(http.ResponseWriter, *http.Request) {}

// checkOverridingViewSet embeds BaseViewSet for its helpers but serves its
// own data, so it replaces the check.
type checkOverridingViewSet struct {
	*BaseViewSet
	checkCustomViewSet
}

func (*checkOverridingViewSet) CheckConfiguration() error { return nil }
func (v *checkOverridingViewSet) List(w http.ResponseWriter, r *http.Request) {
	v.checkCustomViewSet.List(w, r)
}
func (v *checkOverridingViewSet) Create(w http.ResponseWriter, r *http.Request) {}
func (v *checkOverridingViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
}
func (v *checkOverridingViewSet) Update(w http.ResponseWriter, r *http.Request) {}
func (v *checkOverridingViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
}
func (v *checkOverridingViewSet) Destroy(w http.ResponseWriter, r *http.Request) {}

func TestRouterRegister_CustomViewSetsKeepWorking(t *testing.T) {
	for name, vs := range map[string]ViewSet{
		"value":               checkCustomViewSet{},
		"pointer":             &checkCustomViewSet{},
		"overrides the check": &checkOverridingViewSet{BaseViewSet: &BaseViewSet{}},
	} {
		t.Run(name, func(t *testing.T) {
			router := NewRouter("/api")
			require.NotPanics(t, func() { router.Register("custom", vs) })
			handler := forgehttp.NewRouter()
			require.NotPanics(t, func() { router.RegisterRoutes(handler) })
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/custom/", nil))
			assert.Equal(t, http.StatusTeapot, rec.Code)
		})
	}

	assert.Contains(t, registerPanic("custom", (*checkCustomViewSet)(nil)), "viewset is a nil *api.checkCustomViewSet")
	assert.Contains(t, registerPanic("custom", &struct {
		*BaseViewSet
	}{}), "BaseViewSet is nil")
}

func TestRouterRegisterRoutes_ChecksAgainBeforeMounting(t *testing.T) {
	vs := newCheckViewSet(newCheckManager(t))
	router := NewRouter("/api")
	router.Register("widgets", vs)
	vs.Queryset = nil

	handler := forgehttp.NewRouter()
	assert.PanicsWithValue(t, `api: resource "widgets": Queryset is nil`, func() {
		router.RegisterRoutes(handler)
	})
}

func TestViewSetConfig_CheckDoesNotFreezeConfiguration(t *testing.T) {
	config := &ViewSetConfig{Model: &checkedModel{}, Queryset: newCheckManager(t), Serializer: NewBaseSerializer(nil)}
	router := NewRouter("/api")
	router.Register("widgets", config)
	assert.Nil(t, config.viewSet, "checking must not build and cache the viewset")

	config.ReadOnly = true
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/widgets/1", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "a field set after Register must still apply")
}

func jsonBody(body string) io.Reader { return strings.NewReader(body) }
