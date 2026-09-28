package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type listOnlyPermission struct{}

func (p listOnlyPermission) HasPermission(r *http.Request, view permissions.ViewSet) bool {
	// Widen the race window to expose concurrent action clobbering
	time.Sleep(100 * time.Microsecond)
	return view.GetAction() == "list"
}

func (p listOnlyPermission) HasObjectPermission(r *http.Request, view permissions.ViewSet, obj interface{}) bool {
	return view.GetAction() == "list"
}

func (p listOnlyPermission) GetMessage() string { return "only list is permitted" }
func (p listOnlyPermission) GetCode() string    { return "permission_denied" }

type dummyQueryset struct{}

func (d *dummyQueryset) All(context.Context) (interface{}, error) {
	return []map[string]interface{}{{"id": 1}}, nil
}
func (d *dummyQueryset) Count(context.Context) (int64, error) { return 1, nil }
func (d *dummyQueryset) Offset(int) interface{}               { return d }
func (d *dummyQueryset) Limit(int) interface{}                { return d }
func (d *dummyQueryset) Filter(interface{}) interface{}       { return d }
func (d *dummyQueryset) OrderBy(...interface{}) interface{}   { return d }
func (d *dummyQueryset) Get(_ context.Context, id int64) (interface{}, error) {
	return map[string]interface{}{"id": id}, nil
}
func (d *dummyQueryset) Delete(context.Context, interface{}) error { return nil }

func TestConcurrentActionPermissionIsolation(t *testing.T) {
	qs := &dummyQueryset{}
	vs := NewBaseViewSet(
		func() Serializer { return NewBaseSerializer(nil) },
		qs,
		map[string]interface{}{},
	)
	vs.Permissions = []permissions.Permission{listOnlyPermission{}}

	router := NewRouter("/api")
	router.Register("items", vs)
	handler := forgehttp.NewRouter()
	router.RegisterRoutes(handler)

	const numRequests = 50
	var wg sync.WaitGroup
	wg.Add(numRequests * 2)

	listStatusCodes := make([]int, numRequests)
	destroyStatusCodes := make([]int, numRequests)

	for i := 0; i < numRequests; i++ {
		idx := i
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/items/", nil)
			handler.ServeHTTP(rec, req)
			listStatusCodes[idx] = rec.Code
		}()

		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/items/1", nil)
			handler.ServeHTTP(rec, req)
			destroyStatusCodes[idx] = rec.Code
		}()
	}

	wg.Wait()

	for i := 0; i < numRequests; i++ {
		require.Equal(t, http.StatusOK, listStatusCodes[i], "list request %d should be allowed", i)
		assert.Equal(t, http.StatusForbidden, destroyStatusCodes[i], "destroy request %d should be denied", i)
	}
}
