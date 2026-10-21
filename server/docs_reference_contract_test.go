package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/config"
)

func TestDefaultMiddlewareReference_A239(t *testing.T) {
	srv, err := NewServer(config.NewConfig(), &config.Settings{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.RegisterRoutes(func(r *Router) {
		r.Get("/deadline", func(w http.ResponseWriter, req *http.Request) {
			_, hasDeadline := req.Context().Deadline()
			if hasDeadline {
				t.Error("NewServer unexpectedly installed Timeout")
			}
		})
	})
	srv.Handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/deadline", nil))
}

func TestCSRFSecretControlsMiddleware_A255(t *testing.T) {
	for _, tc := range []struct {
		name, secret string
		want         int
	}{
		{"no secret", "", http.StatusNoContent},
		{"secret", "a-csrf-secret-of-at-least-32-bytes", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv, err := NewServer(config.NewConfig(), &config.Settings{Security: config.SecuritySettings{CSRFSecretKey: tc.secret}}, nil)
			if err != nil {
				t.Fatal(err)
			}
			srv.RegisterRoutes(func(r *Router) {
				r.Post("/write", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
			})
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/write", nil))
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestProductionSecretValidationReference_A257(t *testing.T) {
	valid := config.SecuritySettings{
		SecretKey:     "unique-production-app-secret-12345678",
		CSRFSecretKey: "unique-production-csrf-secret-1234567",
		SessionSecret: "unique-production-session-secret-1234",
	}
	for _, tc := range []struct {
		name      string
		keys      config.SecuritySettings
		wantError bool
	}{
		{"distinct keys", valid, false},
		{"short app key", config.SecuritySettings{SecretKey: "short-value", CSRFSecretKey: valid.CSRFSecretKey, SessionSecret: valid.SessionSecret}, true},
		{"short csrf key", config.SecuritySettings{SecretKey: valid.SecretKey, CSRFSecretKey: "short-value", SessionSecret: valid.SessionSecret}, true},
		{"short session key", config.SecuritySettings{SecretKey: valid.SecretKey, CSRFSecretKey: valid.CSRFSecretKey, SessionSecret: "short-value"}, true},
		{"reused keys", config.SecuritySettings{SecretKey: valid.SecretKey, CSRFSecretKey: valid.SecretKey, SessionSecret: valid.SessionSecret}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := &Server{config: config.NewConfig(), settings: &config.Settings{
				App: config.AppSettings{Env: "production"}, Security: tc.keys,
			}}
			err := srv.validateProductionSecrets()
			if (err != nil) != tc.wantError {
				t.Fatalf("validation error = %v, want error = %t", err, tc.wantError)
			}
		})
	}
}

func TestOptionalResponseMiddlewareReference(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("server.cors.enabled", true)
	cfg.Set("server.cors.allowed_origins", []string{"https://store.example.com"})
	srv, err := NewServer(cfg, &config.Settings{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	srv.RegisterRoutes(func(r *Router) {
		r.Get("/plain", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/plain", nil)
	req.Header.Set("Origin", "https://store.example.com")
	srv.Handler.ServeHTTP(rec, req)
	t.Run("A-262", func(t *testing.T) {
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("CORS installed by default: %q", got)
		}
	})
	t.Run("A-263", func(t *testing.T) {
		if got := rec.Header().Get("X-Content-Type-Options"); got != "" {
			t.Fatalf("SecureHeaders installed by default: %q", got)
		}
	})
}
