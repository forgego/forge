package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/exceptions"
	"github.com/stretchr/testify/assert"
)

func TestAPIMiddleware(t *testing.T) {
	middleware := APIMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestAPIMiddleware_PanicRecovery(t *testing.T) {
	middleware := APIMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(exceptions.NewAPIException(
			http.StatusInternalServerError,
			"test_error",
			"Test panic",
			nil,
		))
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Should not panic, should handle gracefully
	wrapped.ServeHTTP(w, req)

	assert.True(t, w.Code >= 400)
}

func TestAPIMiddleware_GenericPanic(t *testing.T) {
	middleware := APIMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("generic panic")
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// Should handle generic panic
	assert.True(t, w.Code >= 500)
}

func TestAuthenticationMiddleware(t *testing.T) {
	mockAuth := &MockAuth{
		ShouldAuth: true,
		User:       &MockUser{ID: "123", Authenticated: true},
	}

	middleware := AuthenticationMiddleware([]authentication.Authentication{mockAuth})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is set
		_, ok := authentication.GetUserFromRequest(r)
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Authenticated"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Authenticated", w.Body.String())
}

func TestAuthenticationMiddleware_NoAuth(t *testing.T) {
	mockAuth := &MockAuth{
		ShouldAuth: false,
	}

	middleware := AuthenticationMiddleware([]authentication.Authentication{mockAuth})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	// Should continue without error
	assert.Equal(t, http.StatusOK, w.Code)
}

// MockAuth and MockUser are defined in viewset_enhanced_test.go

func TestCoreMiddleware_Chain(t *testing.T) {
	middleware1 := core.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "true")
			next.ServeHTTP(w, r)
		})
	})

	middleware2 := core.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "true")
			next.ServeHTTP(w, r)
		})
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	chained := core.Chain(middleware1, middleware2)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	chained.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("X-Middleware-1"))
	assert.Equal(t, "true", w.Header().Get("X-Middleware-2"))
}

func TestCoreMiddleware_Apply(t *testing.T) {
	middleware := core.Middleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Applied", "true")
			next.ServeHTTP(w, r)
		})
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	applied := core.Apply(handler, middleware)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	applied.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("X-Applied"))
}
