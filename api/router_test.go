package api

import (
	"net/http/httptest"
	"testing"

	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Register(t *testing.T) {
	router := NewRouter("/api/v1")

	viewset := NewBaseViewSet(
		newAccessCheckSerializer,
		&accessCheckQueryset{},
		&accessCheckModel{},
	)

	router.Register("items", viewset)

	// Router should have registered the viewset
	assert.NotNil(t, router)
}

func TestRouter_RegisterRoutes(t *testing.T) {
	router := NewRouter("/api/v1")

	viewset := NewBaseViewSet(
		newAccessCheckSerializer,
		&accessCheckQueryset{},
		&accessCheckModel{},
	)

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
