package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordActionPermission struct {
	recorded string
}

func (p *recordActionPermission) HasPermission(r *http.Request, view permissions.ViewSet) bool {
	p.recorded = view.GetAction()
	return true
}

func (p *recordActionPermission) HasObjectPermission(r *http.Request, view permissions.ViewSet, obj interface{}) bool {
	return true
}

func (p *recordActionPermission) GetMessage() string { return "forbidden" }
func (p *recordActionPermission) GetCode() string    { return "permission_denied" }

func TestCustomPermission_ViewGetAction_SeesDispatchedAction(t *testing.T) {
	qs := &dummyQueryset{}
	vs := NewBaseViewSet(
		func() Serializer { return NewBaseSerializer(nil) },
		qs,
		map[string]interface{}{},
	)

	perm := &recordActionPermission{}
	vs.Permissions = []permissions.Permission{perm}

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "list", perm.recorded)
}
