package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/forgego/forge/config"
	"github.com/forgego/forge/netutil"
	"github.com/forgego/forge/stores/storestest"
)

// Metrics expose measured server uptime.
func TestMetricsHandlerReportsParseableUptime(t *testing.T) {
	w := httptest.NewRecorder()
	MetricsHandler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || result["uptime"] == "" {
		t.Fatalf("metrics=%d %v", w.Code, result)
	}
	if _, err := time.ParseDuration(result["uptime"]); err != nil {
		t.Fatal(err)
	}
}

// Production sessions and CSRF cookies have Secure set.
func TestProductionEnablesSecureCookies(t *testing.T) {
	settings := productionSettingsWith(secureTestSecurity())
	settings.Server.HealthCheckPath = "/health"
	srv, err := NewServer(config.NewConfig(), settings, nil, WithDatabase(storestest.OpenMemory(t)))
	if err != nil {
		t.Fatal(err)
	}
	if srv.sessions == nil || !srv.sessions.Cookie.Secure {
		t.Fatal("session cookie is not Secure")
	}
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	for _, cookie := range w.Result().Cookies() {
		if !cookie.Secure {
			t.Fatalf("%s is not Secure", cookie.Name)
		}
	}
}

// Forwarding headers are accepted only from configured proxy peers.
func TestOnlyTrustedProxiesSupplyForwardingHeaders(t *testing.T) {
	if err := netutil.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = netutil.SetTrustedProxies(nil) })
	for _, tc := range []struct{ remote, want string }{{"10.1.2.3:9000", "198.51.100.8"}, {"203.0.113.3:9000", "203.0.113.3"}} {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = tc.remote
		r.Header.Set("X-Forwarded-For", "198.51.100.8")
		if got := netutil.ClientIP(r, netutil.TrustedProxies()); got != tc.want {
			t.Fatalf("peer %s resolved to %s", tc.remote, got)
		}
	}
}

// Missing production secrets are rejected before binding a listener.
func TestProductionRejectsMissingSecretsBeforeListening(t *testing.T) {
	_, err := NewServer(config.NewConfig(), productionSettingsWith(config.SecuritySettings{}), nil)
	if err == nil || !strings.Contains(err.Error(), "production requires explicit") {
		t.Fatalf("NewServer error=%v", err)
	}
}

// Debug mode is rejected on the same pre-listen path.
func TestProductionRejectsDebugBeforeListening(t *testing.T) {
	s := productionSettingsWith(secureTestSecurity())
	s.App.Debug = true
	_, err := NewServer(config.NewConfig(), s, nil)
	if err == nil || !strings.Contains(err.Error(), "app.debug=false") {
		t.Fatalf("NewServer error=%v", err)
	}
}

// Liveness is independent of failing readiness checks.
func TestLivenessReturns200DespiteFailingChecks(t *testing.T) {
	RegisterHealthCheckFunc("unhealthy-for-liveness", func(context.Context) error { return context.DeadlineExceeded })
	t.Cleanup(func() { UnregisterHealthCheck("unhealthy-for-liveness") })
	w := httptest.NewRecorder()
	LivenessHandler().ServeHTTP(w, httptest.NewRequest("GET", "/health/live", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("liveness=%d", w.Code)
	}
}

// NewServer does not install a database or other health check.
func TestNewServerRegistersNoHealthCheckByDefault(t *testing.T) {
	globalRegistry.mu.RLock()
	before := len(globalRegistry.checkers)
	globalRegistry.mu.RUnlock()
	_, err := NewServer(config.NewConfig(), &config.Settings{App: config.AppSettings{Env: "test"}}, nil, WithDatabase(storestest.OpenMemory(t)))
	if err != nil {
		t.Fatal(err)
	}
	globalRegistry.mu.RLock()
	after := len(globalRegistry.checkers)
	globalRegistry.mu.RUnlock()
	if after != before {
		t.Fatalf("health checks changed from %d to %d", before, after)
	}
}

// Shutdown accepts a bounded context and completes with no active requests.
func TestServerGracefulShutdownCompletesWithoutRequests(t *testing.T) {
	s := &config.Settings{App: config.AppSettings{Env: "test"}}
	srv, err := NewServer(config.NewConfig(), s, nil, WithDatabase(storestest.OpenMemory(t)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}
