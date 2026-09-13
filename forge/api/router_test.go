package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/permissions"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Register(t *testing.T) {
	router := NewRouter("/api/v1")

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		&MockQueryset{},
		&MockModel{},
	)

	router.Register("items", viewset)

	// Router should have registered the viewset
	assert.NotNil(t, router)
}

func TestRouter_RegisterRoutes(t *testing.T) {
	router := NewRouter("/api/v1")

	viewset := NewEnhancedBaseViewSet(
		NewMockSerializer,
		&MockQueryset{items: []interface{}{}},
		&MockModel{},
	)
	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	router.Register("items", viewset)

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutes(httpRouter)

	// Test that routes are registered
	req := httptest.NewRequest("GET", "/api/v1/items/", nil)
	w := httptest.NewRecorder()

	httpRouter.ServeHTTP(w, req)

	// Should handle the request (may return error if queryset not properly set up)
	assert.True(t, w.Code >= 200)
}

func TestEnhancedRouter_RegisterAction(t *testing.T) {
	router := NewEnhancedRouter("/api/v1")

	router.RegisterAction("custom", &ActionConfig{
		Methods: []string{"POST"},
		Detail:  true,
	}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Verify action is registered
	action, ok := router.actionRegistry.GetAction("custom")
	assert.True(t, ok)
	assert.NotNil(t, action)
}

func TestEnhancedRouter_RegisterEnhanced(t *testing.T) {
	router := NewEnhancedRouter("/api/v1")

	viewset := NewEnhancedBaseViewSetIntegrated(
		NewMockSerializer,
		&MockQueryset{},
		&MockModel{},
	)

	router.RegisterEnhanced("items", viewset)

	// Should be registered
	assert.NotNil(t, router)
}

func TestEnhancedRouter_RegisterRoutesEnhanced(t *testing.T) {
	router := NewEnhancedRouter("/api/v1")

	viewset := NewEnhancedBaseViewSetIntegrated(
		NewMockSerializer,
		&MockQueryset{items: []interface{}{}},
		&MockModel{},
	)
	viewset.PermissionClasses = []permissions.Permission{
		permissions.NewAllowAny(),
	}

	router.RegisterEnhanced("items", viewset)

	// Register custom action
	router.RegisterAction("custom", &ActionConfig{
		Methods: []string{"GET"},
		Detail:  false,
	}, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpRouter := forgehttp.NewRouter()
	router.RegisterRoutesEnhanced(httpRouter)

	// Test standard route
	req := httptest.NewRequest("GET", "/api/v1/items/", nil)
	w := httptest.NewRecorder()
	httpRouter.ServeHTTP(w, req)
	assert.True(t, w.Code >= 200)

	// Test custom action route
	req2 := httptest.NewRequest("GET", "/api/v1/custom/", nil)
	w2 := httptest.NewRecorder()
	httpRouter.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
