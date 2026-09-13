package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestAdminIndexTransform(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		input    string
		expected string
	}{
		{
			name:     "prefix /admin: meta content /admin injected; asset URLs unchanged",
			prefix:   "/admin",
			input:    `<!DOCTYPE html><html><head><title>Forge Admin</title><script src="/admin/assets/index-x.js"></script><link rel="stylesheet" href="/admin/assets/a.css"></head><body><div id="root"></div></body></html>`,
			expected: `<!DOCTYPE html><html><head><meta name="forge-admin-prefix" content="/admin"><title>Forge Admin</title><script src="/admin/assets/index-x.js"></script><link rel="stylesheet" href="/admin/assets/a.css"></head><body><div id="root"></div></body></html>`,
		},
		{
			name:     "prefix /backoffice: asset URLs rewritten; meta content /backoffice",
			prefix:   "/backoffice",
			input:    `<!DOCTYPE html><html><head><title>Forge Admin</title><script src="/admin/assets/index-x.js"></script><link rel="stylesheet" href="/admin/assets/a.css"></head><body><div id="root"></div></body></html>`,
			expected: `<!DOCTYPE html><html><head><meta name="forge-admin-prefix" content="/backoffice"><title>Forge Admin</title><script src="/backoffice/assets/index-x.js"></script><link rel="stylesheet" href="/backoffice/assets/a.css"></head><body><div id="root"></div></body></html>`,
		},
		{
			name:     "prefix /backoffice: single quotes rewritten",
			prefix:   "/backoffice",
			input:    `<!DOCTYPE html><html><head><script src='/admin/assets/index-x.js'></script></head><body></body></html>`,
			expected: `<!DOCTYPE html><html><head><meta name="forge-admin-prefix" content="/backoffice"><script src='/backoffice/assets/index-x.js'></script></head><body></body></html>`,
		},
		{
			name:     "prefix empty: asset rewritten to /assets/...; meta content empty",
			prefix:   "",
			input:    `<!DOCTYPE html><html><head><title>Forge Admin</title><script src="/admin/assets/index-x.js"></script><link rel="stylesheet" href="/admin/assets/a.css"></head><body><div id="root"></div></body></html>`,
			expected: `<!DOCTYPE html><html><head><meta name="forge-admin-prefix" content=""><title>Forge Admin</title><script src="/assets/index-x.js"></script><link rel="stylesheet" href="/assets/a.css"></head><body><div id="root"></div></body></html>`,
		},
		{
			name:     "prefix with a quote character is escaped in the meta tag",
			prefix:   `/custom"prefix`,
			input:    `<!DOCTYPE html><html><head><title>Forge Admin</title></head><body></body></html>`,
			expected: `<!DOCTYPE html><html><head><meta name="forge-admin-prefix" content="/custom&#34;prefix"><title>Forge Admin</title></head><body></body></html>`,
		},
		{
			name:     "document without <head> still gets the meta tag",
			prefix:   "/admin",
			input:    `<div>Hello world</div>`,
			expected: `<meta name="forge-admin-prefix" content="/admin"><div>Hello world</div>`,
		},
		{
			name:     "head tag case-insensitive (HEAD)",
			prefix:   "/backoffice",
			input:    `<HTML><HEAD><TITLE>Test</TITLE></HEAD><BODY></BODY></HTML>`,
			expected: `<HTML><HEAD><meta name="forge-admin-prefix" content="/backoffice"><TITLE>Test</TITLE></HEAD><BODY></BODY></HTML>`,
		},
		{
			name:     "head tag with attributes (<head lang=\"en\">)",
			prefix:   "/backoffice",
			input:    `<html><head lang="en"><title>Test</title></head><body></body></html>`,
			expected: `<html><head lang="en"><meta name="forge-admin-prefix" content="/backoffice"><title>Test</title></head><body></body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := adminIndexTransform(tt.prefix)
			actual := string(fn([]byte(tt.input)))
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestNormalizeAdminPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/", ""},
		{"///", ""},
		{"   ", ""},
		{"/admin", "/admin"},
		{"/admin/", "/admin"},
		{"admin", "/admin"},
		{"admin/", "/admin"},
		{"/backoffice", "/backoffice"},
		{"/backoffice/", "/backoffice"},
		{"backoffice", "/backoffice"},
		{"  /backoffice/  ", "/backoffice"},
		{"/custom/mount/path/", "/custom/mount/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeAdminPrefix(tt.input))
		})
	}
}

func TestSite_Handler_CustomPrefixServing(t *testing.T) {
	site := NewSite("test")
	uiCfg := site.GetUIConfig()
	uiCfg.Source = UISourceEmbedded
	uiCfg.Prefix = "/backoffice"
	uiCfg.EmbedFS = fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(`<!DOCTYPE html><html><head><title>Admin</title><script src="/admin/assets/main.js"></script></head><body></body></html>`),
		},
		"assets/main.js": &fstest.MapFile{
			Data: []byte(`console.log("main.js");`),
		},
	}
	site.WithUIConfig(uiCfg)
	handler := site.Handler()

	// 1. GET /backoffice/ -> served transformed index.html with meta tag and rewritten URL
	{
		req := httptest.NewRequest(http.MethodGet, "/backoffice/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), `<meta name="forge-admin-prefix" content="/backoffice">`)
		assert.Contains(t, rec.Body.String(), `src="/backoffice/assets/main.js"`)
		assert.NotContains(t, rec.Body.String(), `src="/admin/assets/main.js"`)
		assert.Equal(t, "no-store, no-cache, must-revalidate, private", rec.Header().Get("Cache-Control"))
	}

	// 2. GET /backoffice/assets/main.js -> served untransformed asset
	{
		req := httptest.NewRequest(http.MethodGet, "/backoffice/assets/main.js", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/javascript; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, `console.log("main.js");`, rec.Body.String())
	}

	// 3. GET /backoffice/deep/spa/route -> fallback to transformed index.html
	{
		req := httptest.NewRequest(http.MethodGet, "/backoffice/deep/spa/route", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), `<meta name="forge-admin-prefix" content="/backoffice">`)
		assert.Contains(t, rec.Body.String(), `src="/backoffice/assets/main.js"`)
	}
}
