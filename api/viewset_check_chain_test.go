package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/orm"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// checkFilteredNoCount is a Filter result List cannot count.
type checkFilteredNoCount struct{}

func (*checkFilteredNoCount) All(context.Context) ([]*checkedModel, error) { return nil, nil }

// checkFilteredBadCount has a Count that List would call with a context.
type checkFilteredBadCount struct{}

func (*checkFilteredBadCount) Count() int                                   { return 0 }
func (*checkFilteredBadCount) All(context.Context) ([]*checkedModel, error) { return nil, nil }

// checkStaticFilter's Filter returns a different, statically known type.
type checkStaticFilter struct{ checkWritableQueryset }

func (*checkStaticFilter) Filter(orm.Expression) *checkFilteredNoCount {
	return &checkFilteredNoCount{}
}

// checkPagedNoAll is an Offset result List cannot load.
type checkPagedNoAll struct{}

func (p *checkPagedNoAll) Limit(int) *checkPagedNoAll         { return p }
func (*checkPagedNoAll) Count(context.Context) (int64, error) { return 0, nil }

type checkStaticOffset struct{ checkWritableQueryset }

func (*checkStaticOffset) Offset(int) *checkPagedNoAll    { return &checkPagedNoAll{} }
func (q *checkStaticOffset) Limit(int) *checkStaticOffset { return q }

// checkDynamicFilter's Filter result type is only known per request.
type checkDynamicFilter struct {
	checkWritableQueryset
	result interface{}
}

func (q *checkDynamicFilter) Filter(orm.Expression) interface{} { return q.result }

// Registration follows the list chain: a Filter or Offset result with a
// static type must support the calls List makes on it (#294).
func TestRouterRegister_ChecksStaticListChainResults(t *testing.T) {
	message := registerPanic("widgets", newCheckViewSet(&checkStaticFilter{}))
	require.NotEmpty(t, message, "Register must panic")
	assert.Contains(t, message, "Filter() result *api.checkFilteredNoCount: missing Count(context.Context) (int64, error), needed by list")

	message = registerPanic("widgets", newCheckViewSet(&checkStaticOffset{}))
	require.NotEmpty(t, message, "Register must panic")
	assert.Contains(t, message, "result *api.checkPagedNoAll: missing All(context.Context) ([]T, error), needed by list")
}

// A Filter declared to return interface{} is resolved per request. A result
// List cannot use is a logged configuration error answered with 500, not a
// panic (#294).
func TestBaseViewSetList_DynamicChainResultIsConfigurationError(t *testing.T) {
	for _, tc := range []struct {
		name    string
		result  interface{}
		problem string
	}{
		{"missing Count", &checkFilteredNoCount{}, "missing Count(context.Context) (int64, error)"},
		{"Count with wrong signature", &checkFilteredBadCount{}, "Count has signature Count() int"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.ErrorLevel)
			restore := zap.ReplaceGlobals(zap.New(core))
			t.Cleanup(restore)

			queryset := &checkDynamicFilter{result: tc.result}
			queryset.items = map[int64]*checkedModel{1: {ID: 1, Title: "a"}}
			vs := newCheckViewSet(queryset)
			router := NewRouter("/api")
			router.Register("widgets", vs) // statically unknowable: must not panic
			handler := forgehttp.NewRouter()
			router.RegisterRoutes(handler)

			unfiltered := httptest.NewRecorder()
			handler.ServeHTTP(unfiltered, httptest.NewRequest(http.MethodGet, "/api/widgets/", nil))
			assert.Equal(t, http.StatusOK, unfiltered.Code, unfiltered.Body.String())

			response := httptest.NewRecorder()
			require.NotPanics(t, func() {
				handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/widgets/?title=a", nil))
			})
			assert.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
			assert.NotContains(t, response.Body.String(), "Count", "details are logged, not sent")

			entries := logs.FilterMessage("api: viewset configuration error").All()
			require.Len(t, entries, 1)
			fields := entries[0].ContextMap()
			assert.Equal(t, "widgets", fields["resource"])
			assert.Contains(t, fields["problem"], tc.problem)
		})
	}
}
