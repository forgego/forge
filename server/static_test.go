package server

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestStaticFS_WithIndexTransform(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><head><title>hello</title></head><body><h1>hello world</h1></body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('original app.js');"),
		},
	}

	upperTransform := func(b []byte) []byte {
		return bytes.ToUpper(b)
	}

	handler := StaticFS("", fsys,
		WithFallback("index.html"),
		WithIndexTransform(upperTransform),
	)

	// 1. GET "/" -> transformed body
	{
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		expectedBody := "<HTML><HEAD><TITLE>HELLO</TITLE></HEAD><BODY><H1>HELLO WORLD</H1></BODY></HTML>"
		assert.Equal(t, expectedBody, rec.Body.String())
		assert.Equal(t, fmt.Sprintf("%d", len(expectedBody)), rec.Header().Get("Content-Length"))
	}

	// 2. GET "/deep/route" -> fallback, transformed body
	{
		req := httptest.NewRequest(http.MethodGet, "/deep/route", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		expectedBody := "<HTML><HEAD><TITLE>HELLO</TITLE></HEAD><BODY><H1>HELLO WORLD</H1></BODY></HTML>"
		assert.Equal(t, expectedBody, rec.Body.String())
		assert.Equal(t, fmt.Sprintf("%d", len(expectedBody)), rec.Header().Get("Content-Length"))
	}

	// 3. GET "/assets/app.js" -> original bytes (NOT transformed)
	{
		req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/javascript; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "console.log('original app.js');", rec.Body.String())
	}

	// 4. GET "/index.html" -> direct index request, transformed body
	{
		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		expectedBody := "<HTML><HEAD><TITLE>HELLO</TITLE></HEAD><BODY><H1>HELLO WORLD</H1></BODY></HTML>"
		assert.Equal(t, expectedBody, rec.Body.String())
	}

	// 5. HEAD "/" -> headers only, no body
	{
		req := httptest.NewRequest(http.MethodHead, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		expectedLen := len("<HTML><HEAD><TITLE>HELLO</TITLE></HEAD><BODY><H1>HELLO WORLD</H1></BODY></HTML>")
		assert.Equal(t, fmt.Sprintf("%d", expectedLen), rec.Header().Get("Content-Length"))
		assert.Empty(t, rec.Body.String())
	}
}

func TestStaticFS_WithIndexTransform_CacheHeaders(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><head></head><body>test</body></html>"),
		},
	}

	t.Run("DisableCache true", func(t *testing.T) {
		handler := StaticFS("", fsys,
			WithFallback("index.html"),
			WithDisableCache(true),
			WithIndexTransform(func(b []byte) []byte {
				return bytes.ToUpper(b)
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "no-store, no-cache, must-revalidate, private", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))
		assert.Equal(t, "0", rec.Header().Get("Expires"))
	})

	t.Run("DisableCache false", func(t *testing.T) {
		handler := StaticFS("", fsys,
			WithFallback("index.html"),
			WithDisableCache(false),
			WithMaxAge(7200),
			WithIndexTransform(func(b []byte) []byte {
				return bytes.ToUpper(b)
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=7200", rec.Header().Get("Cache-Control"))
		assert.NotEmpty(t, rec.Header().Get("ETag"))
		assert.NotEmpty(t, rec.Header().Get("Last-Modified"))
	})
}
