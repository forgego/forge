package server

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/netutil"
	"github.com/forgego/forge/stores/storestest"
	"github.com/gorilla/csrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityDefaults(t *testing.T) {
	// (1) CSRF is enforced for state-changing requests authenticated by a cookie session
	t.Run("CSRFEnforcedForCookieSession", func(t *testing.T) {
		cfg := config.NewConfig()
		settings := &config.Settings{
			App: config.AppSettings{Env: "test"},
			Security: config.SecuritySettings{
				SessionSecret: strings.Repeat("s", 32),
				CSRFSecretKey: strings.Repeat("c", 32),
			},
		}
		srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		var csrfToken string
		srv.RegisterRoutes(func(r *Router) {
			r.Get("/login", func(w http.ResponseWriter, req *http.Request) {
				srv.SessionManager().Put(req, "user", "alice")
				csrfToken = csrf.Token(req)
				w.Header().Set("X-CSRF-Token", csrfToken)
				w.WriteHeader(http.StatusOK)
			})
			r.Post("/transfer", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})

		// 1. Establish session and get cookies/token
		rec1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "https://example.com/login", nil)
		srv.Handler.ServeHTTP(rec1, req1)
		require.Equal(t, http.StatusOK, rec1.Code)

		var sessionCookie *http.Cookie
		var csrfCookie *http.Cookie
		for _, c := range rec1.Result().Cookies() {
			if c.Name == "forge_session" {
				sessionCookie = c
			}
			if c.Name == "_gorilla_csrf" {
				csrfCookie = c
			}
		}
		require.NotNil(t, sessionCookie, "session cookie must be set")
		require.NotEmpty(t, csrfToken, "csrf token must be generated")

		// 2. State-changing request with session cookie but WITHOUT CSRF token -> MUST be denied (403)
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodPost, "https://example.com/transfer", nil)
		req2.Header.Set("Referer", "https://example.com/")
		req2.AddCookie(sessionCookie)
		if csrfCookie != nil {
			req2.AddCookie(csrfCookie)
		}
		srv.Handler.ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusForbidden, rec2.Code, "state-changing request with session cookie and no CSRF token must be rejected")

		// 3. State-changing request with session cookie AND valid CSRF token -> MUST be accepted (200)
		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodPost, "https://example.com/transfer", nil)
		req3.Header.Set("Referer", "https://example.com/")
		req3.AddCookie(sessionCookie)
		if csrfCookie != nil {
			req3.AddCookie(csrfCookie)
		}
		req3.Header.Set("X-CSRF-Token", csrfToken)
		srv.Handler.ServeHTTP(rec3, req3)
		t.Logf("body: %s", rec3.Body.String())
		assert.Equal(t, http.StatusOK, rec3.Code, "state-changing request with session cookie and valid CSRF token must succeed")
	})

	// (2) session cookies are Secure (when TLS or production), HttpOnly and SameSite=Lax or Strict
	t.Run("SessionCookiesSecureHttpOnlySameSite", func(t *testing.T) {
		t.Run("Production", func(t *testing.T) {
			cfg := config.NewConfig()
			settings := &config.Settings{
				App: config.AppSettings{Env: "production"},
				Security: config.SecuritySettings{
					SecretKey:     strings.Repeat("k", 32),
					SessionSecret: strings.Repeat("s", 32),
					CSRFSecretKey: strings.Repeat("c", 32),
				},
			}
			srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
			require.NoError(t, err)

			srv.RegisterRoutes(func(r *Router) {
				r.Get("/set-session", func(w http.ResponseWriter, req *http.Request) {
					srv.SessionManager().Put(req, "authenticated", true)
					w.WriteHeader(http.StatusOK)
				})
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/set-session", nil)
			srv.Handler.ServeHTTP(rec, req)

			var sessionCookie *http.Cookie
			for _, c := range rec.Result().Cookies() {
				if c.Name == "forge_session" {
					sessionCookie = c
					break
				}
			}
			require.NotNil(t, sessionCookie, "session cookie must be set")
			assert.True(t, sessionCookie.Secure, "session cookie must be Secure in production")
			assert.True(t, sessionCookie.HttpOnly, "session cookie must be HttpOnly")
			assert.True(t, sessionCookie.SameSite == http.SameSiteStrictMode || sessionCookie.SameSite == http.SameSiteLaxMode, "session cookie SameSite must be Strict or Lax")
		})

		t.Run("TLSInDevelopment", func(t *testing.T) {
			cfg := config.NewConfig()
			settings := &config.Settings{
				App: config.AppSettings{Env: "production"},
				Security: config.SecuritySettings{
					SecretKey:     strings.Repeat("k", 32),
					SessionSecret: strings.Repeat("s", 32),
					CSRFSecretKey: strings.Repeat("c", 32),
				},
			}
			srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
			require.NoError(t, err)

			srv.RegisterRoutes(func(r *Router) {
				r.Get("/set-session", func(w http.ResponseWriter, req *http.Request) {
					srv.SessionManager().Put(req, "authenticated", true)
					w.WriteHeader(http.StatusOK)
				})
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/set-session", nil)
			req.TLS = &tls.ConnectionState{HandshakeComplete: true}
			srv.Handler.ServeHTTP(rec, req)

			var sessionCookie *http.Cookie
			for _, c := range rec.Result().Cookies() {
				if c.Name == "forge_session" {
					sessionCookie = c
					break
				}
			}
			require.NotNil(t, sessionCookie, "session cookie must be set")
			assert.True(t, sessionCookie.Secure, "session cookie must be Secure")
			assert.True(t, sessionCookie.HttpOnly, "session cookie must be HttpOnly")
			assert.True(t, sessionCookie.SameSite == http.SameSiteStrictMode || sessionCookie.SameSite == http.SameSiteLaxMode, "session cookie SameSite must be Strict or Lax")
		})

		t.Run("PlainHTTPInDevelopment", func(t *testing.T) {
			cfg := config.NewConfig()
			settings := &config.Settings{
				App: config.AppSettings{Env: "development"},
				Security: config.SecuritySettings{
					SessionSecret: strings.Repeat("s", 32),
					CSRFSecretKey: strings.Repeat("c", 32),
				},
			}
			srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
			require.NoError(t, err)

			srv.RegisterRoutes(func(r *Router) {
				r.Get("/set-session", func(w http.ResponseWriter, req *http.Request) {
					srv.SessionManager().Put(req, "authenticated", true)
					w.WriteHeader(http.StatusOK)
				})
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/set-session", nil)
			srv.Handler.ServeHTTP(rec, req)

			var sessionCookie *http.Cookie
			for _, c := range rec.Result().Cookies() {
				if c.Name == "forge_session" {
					sessionCookie = c
					break
				}
			}
			require.NotNil(t, sessionCookie, "session cookie must be set")
			assert.False(t, sessionCookie.Secure, "session cookie must not be Secure on plain HTTP in development")
			assert.True(t, sessionCookie.HttpOnly, "session cookie must be HttpOnly")
			assert.True(t, sessionCookie.SameSite == http.SameSiteStrictMode || sessionCookie.SameSite == http.SameSiteLaxMode, "session cookie SameSite must be Strict or Lax")
		})
	})

	// (3) CORS denies cross-origin requests by default (no Access-Control-Allow-Origin for an unknown origin)
	t.Run("CORSDefaultDeniesCrossOrigin", func(t *testing.T) {
		cfg := config.NewConfig()
		settings := &config.Settings{App: config.AppSettings{Env: "test"}}
		srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		srv.RegisterRoutes(func(r *Router) {
			r.Get("/api/data", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})

		// 1. By default, server sends NO Access-Control-Allow-Origin for unknown origin
		rec1 := httptest.NewRecorder()
		req1 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req1.Header.Set("Origin", "https://malicious.example.com")
		srv.Handler.ServeHTTP(rec1, req1)
		assert.Empty(t, rec1.Header().Get("Access-Control-Allow-Origin"), "unknown origin must not receive Access-Control-Allow-Origin by default")

		// 2. Preflight request from unknown origin is not allowed
		rec2 := httptest.NewRecorder()
		req2 := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
		req2.Header.Set("Origin", "https://malicious.example.com")
		req2.Header.Set("Access-Control-Request-Method", "GET")
		srv.Handler.ServeHTTP(rec2, req2)
		assert.Empty(t, rec2.Header().Get("Access-Control-Allow-Origin"), "unknown origin preflight must not receive Access-Control-Allow-Origin")

		// 3. CORS middleware with configured origins denies unknown origin
		corsHandler := CORS(&CORSOptions{
			AllowedOrigins: []string{"https://app.example.com"},
			AllowedMethods: []string{"GET", "POST"},
		})(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		rec3 := httptest.NewRecorder()
		req3 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req3.Header.Set("Origin", "https://unknown.example.com")
		corsHandler.ServeHTTP(rec3, req3)
		assert.Empty(t, rec3.Header().Get("Access-Control-Allow-Origin"), "CORS middleware must deny unknown origin")

		// 4. DefaultCORSOptions() denies unknown origins (no wildcard or empty allowed origins)
		defaultCORSHandler := CORS(nil)(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		rec4 := httptest.NewRecorder()
		req4 := httptest.NewRequest(http.MethodGet, "/api/data", nil)
		req4.Header.Set("Origin", "https://unknown.example.com")
		defaultCORSHandler.ServeHTTP(rec4, req4)
		assert.Empty(t, rec4.Header().Get("Access-Control-Allow-Origin"), "default CORS options must not allow unknown origins")
	})

	// (4) X-Forwarded-For/Proto are ignored unless the peer is a configured trusted proxy
	t.Run("ForwardingHeadersIgnoredUnlessTrustedProxy", func(t *testing.T) {
		t.Cleanup(func() { _ = netutil.SetTrustedProxies(nil) })

		cfg := config.NewConfig()
		settings := &config.Settings{
			App: config.AppSettings{Env: "test"},
			Server: config.ServerSettings{
				TrustedProxies: []string{"10.0.0.0/8"},
			},
		}
		srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		var recordedRemoteAddr string
		var recordedScheme string
		var recordedTLS bool

		srv.RegisterRoutes(func(r *Router) {
			r.Get("/check-headers", func(w http.ResponseWriter, req *http.Request) {
				recordedRemoteAddr = req.RemoteAddr
				recordedScheme = req.URL.Scheme
				recordedTLS = req.TLS != nil
				w.WriteHeader(http.StatusOK)
			})
		})

		// (a) Untrusted peer sends X-Forwarded-For and X-Forwarded-Proto -> MUST BE IGNORED
		{
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/check-headers", nil)
			req.RemoteAddr = "203.0.113.10:4321"
			req.Header.Set("X-Forwarded-For", "198.51.100.5")
			req.Header.Set("X-Forwarded-Proto", "https")
			srv.Handler.ServeHTTP(rec, req)

			assert.Equal(t, "203.0.113.10:4321", recordedRemoteAddr, "untrusted peer RemoteAddr must not be rewritten")
			assert.False(t, recordedTLS, "untrusted peer X-Forwarded-Proto must not enable TLS")
			assert.NotEqual(t, "https", recordedScheme, "untrusted peer X-Forwarded-Proto must not set scheme to https")
		}

		// (b) Configured trusted proxy sends X-Forwarded-For and X-Forwarded-Proto -> MUST BE HONORED
		{
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/check-headers", nil)
			req.RemoteAddr = "10.1.2.3:5555"
			req.Header.Set("X-Forwarded-For", "198.51.100.5")
			req.Header.Set("X-Forwarded-Proto", "https")
			srv.Handler.ServeHTTP(rec, req)

			assert.Equal(t, "198.51.100.5", recordedRemoteAddr, "trusted proxy X-Forwarded-For must be honored")
			assert.True(t, recordedTLS || recordedScheme == "https", "trusted proxy X-Forwarded-Proto must be honored")
		}

		// (c) Standalone RealIP middleware: untrusted vs trusted
		{
			var seenAddr string
			var seenTLS bool
			var seenScheme string
			h := RealIP(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				seenAddr = req.RemoteAddr
				seenTLS = req.TLS != nil
				seenScheme = req.URL.Scheme
				w.WriteHeader(http.StatusOK)
			}))

			// Untrusted peer
			reqUntrusted := httptest.NewRequest(http.MethodGet, "/", nil)
			reqUntrusted.RemoteAddr = "203.0.113.20:9999"
			reqUntrusted.Header.Set("X-Forwarded-For", "1.1.1.1")
			reqUntrusted.Header.Set("X-Forwarded-Proto", "https")
			h.ServeHTTP(httptest.NewRecorder(), reqUntrusted)
			assert.Equal(t, "203.0.113.20:9999", seenAddr)
			assert.False(t, seenTLS)
			assert.NotEqual(t, "https", seenScheme)

			// Trusted peer
			reqTrusted := httptest.NewRequest(http.MethodGet, "/", nil)
			reqTrusted.RemoteAddr = "10.2.3.4:8888"
			reqTrusted.Header.Set("X-Forwarded-For", "198.51.100.77")
			reqTrusted.Header.Set("X-Forwarded-Proto", "https")
			h.ServeHTTP(httptest.NewRecorder(), reqTrusted)
			assert.Equal(t, "198.51.100.77", seenAddr)
			assert.True(t, seenTLS || seenScheme == "https")
		}
	})

	// (5) sensitive responses (auth/session endpoints) carry Cache-Control: no-store
	t.Run("SensitiveResponsesCacheControlNoStore", func(t *testing.T) {
		cfg := config.NewConfig()
		settings := &config.Settings{
			App: config.AppSettings{Env: "test"},
			Security: config.SecuritySettings{
				SessionSecret:   strings.Repeat("s", 32),
				CSRFSecretKey:   strings.Repeat("c", 32),
				CSRFExemptPaths: []string{"/session"},
			},
		}
		srv, err := NewServer(cfg, settings, nil, WithDatabase(storestest.OpenMemory(t)))
		require.NoError(t, err)

		srv.RegisterRoutes(func(r *Router) {
			// Session login endpoint: creates session
			r.Post("/session/login", func(w http.ResponseWriter, req *http.Request) {
				srv.SessionManager().Put(req, "user_id", "42")
				w.WriteHeader(http.StatusOK)
			})
			// Session logout endpoint: destroys session
			r.Post("/session/logout", func(w http.ResponseWriter, req *http.Request) {
				_ = srv.SessionManager().Destroy(req)
				w.WriteHeader(http.StatusOK)
			})
			// Sensitive auth endpoint wrapped with NoCache
			r.With(NoCache).Get("/auth/userinfo", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			// Regular public endpoint
			r.Get("/public/hello", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		})

		// (a) Endpoint modifying session must carry Cache-Control: no-store
		{
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/session/login", nil)
			srv.Handler.ServeHTTP(rec, req)
			cc := rec.Header().Get("Cache-Control")
			assert.Contains(t, cc, "no-store", "session-establishing endpoint must carry Cache-Control: no-store")
		}

		// (b) Endpoint destroying session must carry Cache-Control: no-store
		{
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/session/logout", nil)
			srv.Handler.ServeHTTP(rec, req)
			cc := rec.Header().Get("Cache-Control")
			assert.Contains(t, cc, "no-store", "session-destroying endpoint must carry Cache-Control: no-store")
		}

		// (c) Sensitive auth endpoint with NoCache must carry Cache-Control: no-store
		{
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/auth/userinfo", nil)
			srv.Handler.ServeHTTP(rec, req)
			cc := rec.Header().Get("Cache-Control")
			assert.Contains(t, cc, "no-store", "sensitive auth endpoint must carry Cache-Control: no-store")
		}

		// (d) Response helper NoCache() must carry Cache-Control: no-store
		{
			rec := httptest.NewRecorder()
			resp := NewResponse(rec)
			resp.NoCache()
			assert.Contains(t, rec.Header().Get("Cache-Control"), "no-store")
		}
	})
}
