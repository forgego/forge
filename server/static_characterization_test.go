package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestStaticFS_Characterization(t *testing.T) {
	files := fstest.MapFS{
		"index.html":      &fstest.MapFile{Data: []byte("root index")},
		"docs/index.html": &fstest.MapFile{Data: []byte("docs index")},
		"app.js":          &fstest.MapFile{Data: []byte("console.log(1)")},
		"secret.txt":      &fstest.MapFile{Data: []byte("secret")},
	}

	for _, tc := range []struct {
		name, url, body, contentType, cacheControl string
		status                                     int
		opts                                       []StaticOption
	}{
		{"file content and type", "/app.js", "console.log(1)", "application/javascript; charset=utf-8", "public, max-age=60", http.StatusOK, []StaticOption{WithMaxAge(60)}},
		{"directory index", "/docs/", "docs index", "text/html; charset=utf-8", "public, max-age=3600", http.StatusOK, nil},
		{"missing file", "/missing", "404 page not found\n", "text/plain; charset=utf-8", "", http.StatusNotFound, nil},
		{"cleaned traversal path", "/../secret.txt", "secret", "text/plain; charset=utf-8", "public, max-age=3600", http.StatusOK, nil},
		{"prefix mismatch", "/app.js", "404 page not found\n", "text/plain; charset=utf-8", "", http.StatusNotFound, []StaticOption{WithPrefix("/assets")}},
		{"fallback", "/client/route", "root index", "text/html; charset=utf-8", "public, max-age=3600", http.StatusOK, []StaticOption{WithFallback("index.html")}},
		{"cache disabled", "/app.js", "console.log(1)", "application/javascript; charset=utf-8", "no-store, no-cache, must-revalidate, private", http.StatusOK, []StaticOption{WithDisableCache(true)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			StaticFS("", files, tc.opts...).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.url, nil))
			assert.Equal(t, tc.status, recorder.Code)
			assert.Equal(t, tc.body, recorder.Body.String())
			assert.Equal(t, tc.contentType, recorder.Header().Get("Content-Type"))
			assert.Equal(t, tc.cacheControl, recorder.Header().Get("Cache-Control"))
		})
	}
}

func TestStaticFS_Characterization_Advanced(t *testing.T) {
	files := fstest.MapFS{
		"index.html":         &fstest.MapFile{Data: []byte("root index")},
		"docs/index.html":    &fstest.MapFile{Data: []byte("docs index")},
		"custom/custom.html": &fstest.MapFile{Data: []byte("custom index")},
		"app.js":             &fstest.MapFile{Data: []byte("console.log(1)")},
		"secret.txt":         &fstest.MapFile{Data: []byte("secret")},
		"empty.txt":          &fstest.MapFile{Data: []byte{}},
		"empty-dir/sub.keep": &fstest.MapFile{Data: []byte("keep")},
	}

	t.Run("show indexes directory listing", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
		StaticFS("", files, WithShowIndexes(true)).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "Index of /docs")
	})

	t.Run("directory without index falls back to fallback file", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/empty-dir/", nil)
		StaticFS("", files, WithFallback("index.html")).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "root index", rec.Body.String())
	})

	t.Run("directory without index and no fallback returns 404", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/empty-dir/", nil)
		StaticFS("", files).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("custom index files", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/custom/", nil)
		StaticFS("", files, WithIndexFiles("custom.html")).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "custom index", rec.Body.String())
	})

	t.Run("ETag and If-None-Match 304", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		handler := StaticFS("", files)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		etag := rec.Header().Get("ETag")
		assert.NotEmpty(t, etag)

		// Exact match
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		req2.Header.Set("If-None-Match", etag)
		handler.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusNotModified, rec2.Code)

		// Wildcard match
		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		req3.Header.Set("If-None-Match", "*")
		handler.ServeHTTP(rec3, req3)
		assert.Equal(t, http.StatusNotModified, rec3.Code)
	})

	t.Run("If-Modified-Since 304", func(t *testing.T) {
		handler := StaticFS("", files)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		handler.ServeHTTP(rec, req)
		lastMod := rec.Header().Get("Last-Modified")
		assert.NotEmpty(t, lastMod)

		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		req2.Header.Set("If-Modified-Since", lastMod)
		handler.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusNotModified, rec2.Code)
	})

	t.Run("empty file returns 200 without accept ranges", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/empty.txt", nil)
		StaticFS("", files).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Empty(t, rec.Header().Get("Accept-Ranges"))
		assert.Empty(t, rec.Body.String())
	})

	t.Run("pattern prefix routing", func(t *testing.T) {
		handler := StaticFS("/assets", files)

		// Match
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "console.log(1)", rec.Body.String())

		// Mismatch
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodGet, "/other/app.js", nil)
		handler.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusNotFound, rec2.Code)
	})

	t.Run("range request partial content", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/secret.txt", nil)
		req.Header.Set("Range", "bytes=0-2")
		StaticFS("", files).ServeHTTP(rec, req)
		assert.Equal(t, http.StatusPartialContent, rec.Code)
		assert.Equal(t, "bytes 0-2/6", rec.Header().Get("Content-Range"))
		assert.Equal(t, "3", rec.Header().Get("Content-Length"))
		assert.Equal(t, "sec", rec.Body.String())
	})
}
