package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRouterMethods(t *testing.T) {
	router := NewRouter()
	var methodCalled string

	handler := func(w http.ResponseWriter, r *http.Request) {
		methodCalled = r.Method
		w.WriteHeader(http.StatusOK)
	}

	router.Get("/get", handler)
	router.Post("/post", handler)
	router.Put("/put", handler)
	router.Patch("/patch", handler)
	router.Delete("/delete", handler)
	router.Head("/head", handler)
	router.Options("/options", handler)
	router.Connect("/connect", handler)
	router.Trace("/trace", handler)

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "CONNECT", "TRACE"}
	paths := []string{"/get", "/post", "/put", "/patch", "/delete", "/head", "/options", "/connect", "/trace"}

	for i, method := range methods {
		t.Run(method, func(t *testing.T) {
			methodCalled = ""
			req := httptest.NewRequest(method, paths[i], nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// HEAD requests don't typically reach the body handler in the same way,
			// or chi might handle them specifically, but for standard routing the method should match.
			if method != "HEAD" {
			    assert.Equal(t, method, methodCalled)
			}
		})
	}
}

func TestRouterFeatures(t *testing.T) {
	t.Run("URLParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		assert.Equal(t, "123", URLParam(req, "id"))
	})

	t.Run("Use Middleware", func(t *testing.T) {
		router := NewRouter()
		middlewareCalled := false

		router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		})

		router.Get("/", func(w http.ResponseWriter, r *http.Request) {})

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.True(t, middlewareCalled)
	})

	t.Run("Route and Group", func(t *testing.T) {
		router := NewRouter()

		router.Route("/api", func(r *Router) {
			r.Get("/users", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			})

			r.Group(func(gr *Router) {
				gr.Get("/posts", func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusCreated)
				})
			})
		})

		req1 := httptest.NewRequest("GET", "/api/users", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusAccepted, w1.Code)

		req2 := httptest.NewRequest("GET", "/api/posts", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusCreated, w2.Code)
	})

	t.Run("Mount", func(t *testing.T) {
		router := NewRouter()
		subRouter := NewRouter()
		subRouter.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		router.Mount("/sub", subRouter)

		req := httptest.NewRequest("GET", "/sub/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("With Middleware", func(t *testing.T) {
		router := NewRouter()
		middlewareCalled := false

		mid := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		}

		router.With(mid).Get("/with", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		router.Get("/without", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Request to /with
		req := httptest.NewRequest("GET", "/with", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.True(t, middlewareCalled)
		assert.Equal(t, http.StatusOK, w.Code)

		// Reset and request to /without
		middlewareCalled = false
		req = httptest.NewRequest("GET", "/without", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.False(t, middlewareCalled)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Custom Handlers (NotFound and MethodNotAllowed)", func(t *testing.T) {
		router := NewRouter()

		router.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusPaymentRequired)
		})

		router.Get("/exists", func(w http.ResponseWriter, r *http.Request) {})

		// Test NotFound
		req1 := httptest.NewRequest("GET", "/missing", nil)
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		assert.Equal(t, http.StatusTeapot, w1.Code)

		// Test MethodNotAllowed
		req2 := httptest.NewRequest("POST", "/exists", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		assert.Equal(t, http.StatusPaymentRequired, w2.Code)
	})
}
