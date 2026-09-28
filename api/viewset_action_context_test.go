package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type coreActionRecordingPermission struct {
	action string
	ok     bool
}

func (p *coreActionRecordingPermission) HasPermission(r *http.Request, _ permissions.ViewSet) bool {
	p.action, p.ok = core.ActionFromContext(r.Context())
	return true
}

func (*coreActionRecordingPermission) HasObjectPermission(*http.Request, permissions.ViewSet, interface{}) bool {
	return true
}

func (*coreActionRecordingPermission) GetMessage() string { return "" }
func (*coreActionRecordingPermission) GetCode() string    { return "" }

func TestBaseViewSet_DispatchStoresActionInCoreContext(t *testing.T) {
	permission := &coreActionRecordingPermission{}
	vs := NewBaseViewSet(
		func() Serializer { return NewBaseSerializer(nil) },
		&dummyQueryset{},
		&map[string]interface{}{},
	)
	vs.Permissions = []permissions.Permission{permission}
	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/items/", nil))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.True(t, permission.ok)
	assert.Equal(t, "list", permission.action)
}
