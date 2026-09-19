package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typeAssertPermission struct {
	called bool
}

func (p *typeAssertPermission) HasPermission(r *http.Request, view permissions.ViewSet) bool {
	baseVS, ok := view.(*BaseViewSet)
	if !ok || baseVS == nil {
		return false
	}
	p.called = true
	return true
}

func (p *typeAssertPermission) HasObjectPermission(r *http.Request, view permissions.ViewSet, obj interface{}) bool {
	baseVS, ok := view.(*BaseViewSet)
	if !ok || baseVS == nil {
		return false
	}
	return true
}

func (p *typeAssertPermission) GetMessage() string { return "custom permission denied" }
func (p *typeAssertPermission) GetCode() string    { return "permission_denied" }

type typeAssertThrottle struct {
	called bool
}

func (t *typeAssertThrottle) AllowRequest(r *http.Request, view interface{}) (bool, time.Duration, error) {
	baseVS, ok := view.(*BaseViewSet)
	if !ok || baseVS == nil {
		return false, time.Second, nil
	}
	t.called = true
	return true, 0, nil
}

func (t *typeAssertThrottle) GetScope(r *http.Request, view interface{}) string {
	return "test_scope"
}

func TestCustomThrottleAndPermissionAssertBaseViewSet(t *testing.T) {
	qs := &dummyQueryset{}
	vs := NewBaseViewSet(
		func() Serializer { return NewBaseSerializer(nil) },
		qs,
		map[string]interface{}{},
	)

	perm := &typeAssertPermission{}
	throttle := &typeAssertThrottle{}

	vs.Permissions = []permissions.Permission{perm}
	vs.Throttles = []throttling.Throttle{throttle}

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, perm.called, "custom permission asserting *BaseViewSet must be called")
	assert.True(t, throttle.called, "custom throttle asserting *BaseViewSet must be called")
}
