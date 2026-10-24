package api

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mountedRoute struct {
	Method string
	Path   string
}

func recordMountedRoutes(r *forgehttp.Router) []mountedRoute {
	var routes []mountedRoute
	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routes = append(routes, mountedRoute{
			Method: method,
			Path:   route,
		})
		return nil
	})
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path != routes[j].Path {
			return routes[i].Path < routes[j].Path
		}
		return routes[i].Method < routes[j].Method
	})
	return routes
}

func TestRouteDescriptors(t *testing.T) {
	router := NewRouter("/api/v1")
	router.Register("items", createRouterTestViewSet())
	router.Register("categories", createRouterTestViewSet())

	router.Action("items", "publish", ActionConfig{
		Methods: []string{http.MethodPost},
		Detail:  true,
	}, func(w http.ResponseWriter, r *http.Request) {})

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	mounted := recordMountedRoutes(httpRouter)

	// Snapshot of mounted routes before refactoring.
	// This set must not change after refactoring.
	expectedSnapshot := []mountedRoute{
		{Method: "GET", Path: "/api/v1/categories/"},
		{Method: "OPTIONS", Path: "/api/v1/categories/"},
		{Method: "POST", Path: "/api/v1/categories/"},
		{Method: "DELETE", Path: "/api/v1/categories/{id}"},
		{Method: "GET", Path: "/api/v1/categories/{id}"},
		{Method: "PATCH", Path: "/api/v1/categories/{id}"},
		{Method: "PUT", Path: "/api/v1/categories/{id}"},
		{Method: "GET", Path: "/api/v1/items/"},
		{Method: "OPTIONS", Path: "/api/v1/items/"},
		{Method: "POST", Path: "/api/v1/items/"},
		{Method: "DELETE", Path: "/api/v1/items/{id}"},
		{Method: "GET", Path: "/api/v1/items/{id}"},
		{Method: "PATCH", Path: "/api/v1/items/{id}"},
		{Method: "PUT", Path: "/api/v1/items/{id}"},
		{Method: "POST", Path: "/api/v1/items/{id}/publish"},
	}

	require.Equal(t, expectedSnapshot, mounted, "mounted paths must match expected snapshot")
	assert.Len(t, mounted, 15)

	// Assert that Routes() returns descriptors whose (method, path) set equals the mounted set.
	descriptors := router.Routes()
	require.Len(t, descriptors, len(mounted))

	var descriptorRoutes []mountedRoute
	for _, d := range descriptors {
		descriptorRoutes = append(descriptorRoutes, mountedRoute{
			Method: d.Method,
			Path:   d.Path,
		})
	}
	sort.Slice(descriptorRoutes, func(i, j int) bool {
		if descriptorRoutes[i].Path != descriptorRoutes[j].Path {
			return descriptorRoutes[i].Path < descriptorRoutes[j].Path
		}
		return descriptorRoutes[i].Method < descriptorRoutes[j].Method
	})
	assert.Equal(t, mounted, descriptorRoutes, "mounted routes set must equal Routes() projection")

	// Verify descriptor details.
	expectedDescriptors := []RouteDescriptor{
		{Method: "GET", Path: "/api/v1/categories/", Resource: "categories", Action: "list", Detail: false},
		{Method: "POST", Path: "/api/v1/categories/", Resource: "categories", Action: "create", Detail: false},
		{Method: "GET", Path: "/api/v1/categories/{id}", Resource: "categories", Action: "retrieve", Detail: true},
		{Method: "PUT", Path: "/api/v1/categories/{id}", Resource: "categories", Action: "update", Detail: true},
		{Method: "PATCH", Path: "/api/v1/categories/{id}", Resource: "categories", Action: "partial_update", Detail: true},
		{Method: "DELETE", Path: "/api/v1/categories/{id}", Resource: "categories", Action: "destroy", Detail: true},
		{Method: "OPTIONS", Path: "/api/v1/categories/", Resource: "categories", Action: "metadata", Detail: false},
		{Method: "GET", Path: "/api/v1/items/", Resource: "items", Action: "list", Detail: false},
		{Method: "POST", Path: "/api/v1/items/", Resource: "items", Action: "create", Detail: false},
		{Method: "GET", Path: "/api/v1/items/{id}", Resource: "items", Action: "retrieve", Detail: true},
		{Method: "PUT", Path: "/api/v1/items/{id}", Resource: "items", Action: "update", Detail: true},
		{Method: "PATCH", Path: "/api/v1/items/{id}", Resource: "items", Action: "partial_update", Detail: true},
		{Method: "DELETE", Path: "/api/v1/items/{id}", Resource: "items", Action: "destroy", Detail: true},
		{Method: "OPTIONS", Path: "/api/v1/items/", Resource: "items", Action: "metadata", Detail: false},
		{Method: "POST", Path: "/api/v1/items/{id}/publish", Resource: "items", Action: "publish", Detail: true},
	}
	assert.Equal(t, expectedDescriptors, descriptors)
}

func TestRouteDescriptors_CollectionActionAndCustomRoute(t *testing.T) {
	router := NewRouter("/api/v1")
	router.Register("items", createRouterTestViewSet())

	router.Action("items", "export", ActionConfig{
		Methods: []string{http.MethodGet, http.MethodPost},
		Detail:  false,
	}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	mounted := recordMountedRoutes(httpRouter)
	descriptors := router.Routes()
	require.Len(t, descriptors, len(mounted))

	var descriptorRoutes []mountedRoute
	for _, d := range descriptors {
		descriptorRoutes = append(descriptorRoutes, mountedRoute{
			Method: d.Method,
			Path:   d.Path,
		})
	}
	sort.Slice(descriptorRoutes, func(i, j int) bool {
		if descriptorRoutes[i].Path != descriptorRoutes[j].Path {
			return descriptorRoutes[i].Path < descriptorRoutes[j].Path
		}
		return descriptorRoutes[i].Method < descriptorRoutes[j].Method
	})
	assert.Equal(t, mounted, descriptorRoutes)

	// Verify collection action and custom route are properly routed and reachable.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items/export", nil)
	httpRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	httpRouter.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}
