package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/forgego/forge/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ipCounter atomic.Uint32

func TestRateLimitByIP_UntrustedHeaderSpoofing(t *testing.T) {
	// Reset trusted proxies to empty (trust no one)
	err := netutil.SetTrustedProxies(nil)
	require.NoError(t, err)

	handler := RateLimitByIP(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	remoteAddr := fmt.Sprintf("203.0.113.%d:1234", ipCounter.Add(1)%250+1)

	// Request 1: RemoteAddr with XFF "1.1.1.1" -> allowed
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = remoteAddr
	req1.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Request 2: Same RemoteAddr, DIFFERENT X-Forwarded-For "2.2.2.2" -> must get 429
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = remoteAddr
	req2.Header.Set("X-Forwarded-For", "2.2.2.2")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code, "second request from same RemoteAddr with spoofed X-Forwarded-For must be rate limited")
}

func TestRealIP_TrustedProxiesOnly(t *testing.T) {
	require.NoError(t, netutil.SetTrustedProxies([]string{"10.0.0.0/8"}))
	t.Cleanup(func() { _ = netutil.SetTrustedProxies(nil) })

	var seen string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.RemoteAddr
	}))

	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xRealIP    string
		want       string
	}{
		{"untrusted peer spoofs X-Forwarded-For", "203.0.113.9:4444", "1.2.3.4", "", "203.0.113.9:4444"},
		{"untrusted peer spoofs X-Real-IP", "203.0.113.9:4444", "", "1.2.3.4", "203.0.113.9:4444"},
		{"trusted proxy forwards X-Forwarded-For", "10.1.2.3:5555", "198.51.100.7, 10.0.0.2", "", "198.51.100.7"},
		{"trusted proxy forwards X-Real-IP", "10.1.2.3:5555", "", "198.51.100.8", "198.51.100.8"},
		{"trusted proxy without headers keeps peer", "10.1.2.3:5555", "", "", "10.1.2.3:5555"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)
			assert.Equal(t, tt.want, seen)
		})
	}
}

func TestRealIP_NoTrustedProxiesIgnoresHeaders(t *testing.T) {
	require.NoError(t, netutil.SetTrustedProxies(nil))

	var seen string
	handler := RealIP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.RemoteAddr
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9000"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "5.6.7.8")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	assert.Equal(t, "127.0.0.1:9000", seen)
}
