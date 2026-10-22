package server

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/spf13/viper"
)

func documentedServer(t *testing.T, settings *config.Settings) *Server {
	t.Helper()
	srv, err := NewServer(&config.Config{Viper: viper.New()}, settings, nil)
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

func TestServerServesConfiguredStaticFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.txt"), []byte("static content"), 0o600); err != nil {
		t.Fatal(err)
	}
	srv := documentedServer(t, &config.Settings{Server: config.ServerSettings{StaticFilesPath: dir}})
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/static/app.txt", nil))
	if w.Code != http.StatusOK || w.Body.String() != "static content" {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
}

func TestCORSMiddlewareAnswersAllowedPreflight(t *testing.T) {
	h := CORS(&CORSOptions{AllowedOrigins: []string{"https://app.example.test"}, AllowedMethods: []string{"GET", "POST"}})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://app.example.test")
	r.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.test" || !strings.Contains(w.Header().Get("Access-Control-Allow-Methods"), "POST") {
		t.Fatalf("headers=%v", w.Header())
	}
}

func TestCompressMiddlewareGzipsAcceptedResponse(t *testing.T) {
	h := Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, strings.Repeat("hello", 200))
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("encoding=%q", w.Header().Get("Content-Encoding"))
	}
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil || string(decoded) != strings.Repeat("hello", 200) {
		t.Fatalf("decoded=%q err=%v", decoded, err)
	}
}

func TestSecureHeadersMiddlewareSetsBrowserProtections(t *testing.T) {
	h := SecureHeaders(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("headers=%v", w.Header())
	}
}

func TestSlashMiddlewareStripsAndRedirectsTrailingSlash(t *testing.T) {
	path := ""
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { path = r.URL.Path; w.WriteHeader(http.StatusOK) })
	w := httptest.NewRecorder()
	StripSlashes(next).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/items/", nil))
	if w.Code != http.StatusOK || path != "/items" {
		t.Fatalf("status=%d path=%q", w.Code, path)
	}
	w = httptest.NewRecorder()
	RedirectSlashes(next).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/items/", nil))
	if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "/items" {
		t.Fatalf("status=%d location=%q", w.Code, w.Header().Get("Location"))
	}
}

func TestMetricsEndpointReportsUptime(t *testing.T) {
	srv := documentedServer(t, &config.Settings{Server: config.ServerSettings{MetricsEnabled: true, MetricsPath: "/metrics"}})
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || body["uptime"] == nil {
		t.Fatalf("status=%d body=%v", w.Code, body)
	}
}

func TestProfilingRequiresFlagAndDebugMode(t *testing.T) {
	for _, tc := range []struct{ flag, debug, enabled bool }{{false, false, false}, {true, false, false}, {false, true, false}, {true, true, true}} {
		srv := documentedServer(t, &config.Settings{App: config.AppSettings{Env: "development", Debug: tc.debug}, Server: config.ServerSettings{EnableProfiling: tc.flag, HealthCheckPath: "/health"}})
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil))
		if (w.Code == http.StatusOK) != tc.enabled {
			t.Fatalf("flag=%v debug=%v status=%d", tc.flag, tc.debug, w.Code)
		}
	}
}

func TestSessionCookieIsHttpOnly(t *testing.T) {
	sm := NewSessionManager(nil)
	if !sm.Cookie.HttpOnly {
		t.Fatal("session cookie is script-readable")
	}
}

func TestCSRFRejectsStateChangingMethodsWithoutToken(t *testing.T) {
	settings := &config.Settings{Security: config.SecuritySettings{CSRFSecretKey: strings.Repeat("c", 32)}}
	srv := documentedServer(t, settings)
	srv.RegisterRoutes(func(r *Router) {
		accepted := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }
		r.Post("/submit", accepted)
		r.Put("/submit", accepted)
		r.Patch("/submit", accepted)
		r.Delete("/submit", accepted)
	})
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, httptest.NewRequest(method, "/submit", nil))
		if w.Code != http.StatusForbidden {
			t.Fatalf("method=%s status=%d body=%q", method, w.Code, w.Body.String())
		}
	}
}

func TestCSRFTokenCookieIsHttpOnly(t *testing.T) {
	csrf := NewCSRF([]byte(strings.Repeat("c", 32)), DefaultCSRFOptions()...)
	h := csrf.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, csrf.Token(r)) }))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/form", nil))
	if w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
	found := false
	for _, cookie := range w.Result().Cookies() {
		if cookie.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Fatalf("no HttpOnly CSRF cookie: %v", w.Result().Cookies())
	}
}
