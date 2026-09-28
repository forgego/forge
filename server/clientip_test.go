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
