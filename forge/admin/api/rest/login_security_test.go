package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === BUG 2 TESTS ===

func TestHandleLogin_PasswordWhitespacePreserved(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "p w ")

	router := NewRouter(core.NewRegistry())

	// Exact password with trailing space succeeds
	req1 := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"p w "}`))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.handleLogin(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	// Trimmed password without trailing space fails (401)
	req2 := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"p w"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	router.handleLogin(rec2, req2)
	assert.Equal(t, http.StatusUnauthorized, rec2.Code)
}

func TestHandleLogin_PasswordTrailingNewlineTrimmedFromEnv(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "p w \r\n")

	router := NewRouter(core.NewRegistry())

	// Password with trailing space (before \r\n) matches
	req1 := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"p w "}`))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.handleLogin(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)
}

// === BUG 3 TESTS ===

func TestLoginLimiter_Unit(t *testing.T) {
	limiter := newLoginLimiter()
	assert.Equal(t, 5, limiter.max)
	assert.Equal(t, 15*time.Minute, limiter.window)

	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	key := "ip:127.0.0.1"

	// 4 failures should not block
	for i := 0; i < 4; i++ {
		limiter.fail(key)
		blocked, _ := limiter.blocked(key)
		assert.False(t, blocked)
	}

	// 5th failure blocks
	limiter.fail(key)
	blocked, rem := limiter.blocked(key)
	assert.True(t, blocked)
	assert.Equal(t, 15*time.Minute, rem)

	// Advance time by 10 minutes (still within 15 minute window)
	now = now.Add(10 * time.Minute)
	blocked, rem = limiter.blocked(key)
	assert.True(t, blocked)
	assert.Equal(t, 5*time.Minute, rem)

	// Advance time past 15 minute window -> unblocked
	now = now.Add(5*time.Minute + time.Second)
	blocked, _ = limiter.blocked(key)
	assert.False(t, blocked)

	// Success clears the counter
	limiter.fail(key)
	limiter.success(key)
	blocked, _ = limiter.blocked(key)
	assert.False(t, blocked)
	limiter.mu.Lock()
	_, exists := limiter.failures[key]
	limiter.mu.Unlock()
	assert.False(t, exists)
}

func TestHandleLogin_RateLimiting(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	router := NewRouter(core.NewRegistry())

	mockNow := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	router.loginLimiter.now = func() time.Time { return mockNow }

	// 5 failed attempts from IP 10.0.0.1 with user "admin"
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code, "attempt %d should be 401", i)
	}

	// 6th attempt with CORRECT password from same IP should be blocked (429)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusTooManyRequests, rec.Code)

		retryAfter := rec.Header().Get("Retry-After")
		require.NotEmpty(t, retryAfter)
		sec, err := strconv.Atoi(retryAfter)
		require.NoError(t, err)
		assert.Greater(t, sec, 0)

		var payload map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
		errPayload, ok := payload["error"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "too_many_attempts", errPayload["code"])
		assert.Equal(t, "Too many failed login attempts. Try again later.", errPayload["message"])
	}

	// Different username from the SAME IP is also blocked (ip key)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"otheruser","password":"secret"}`))
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusTooManyRequests, rec.Code)
	}

	// Admin user from a DIFFERENT IP is also blocked (user key)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
		req.RemoteAddr = "192.168.1.1:9999"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusTooManyRequests, rec.Code)
	}

	// After window passes, correct password succeeds (200)
	mockNow = mockNow.Add(15*time.Minute + time.Second)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
	}

	// Success cleared counters, so user can immediately fail 1 time and get 401 (not 429)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

// === BUG 4 TESTS ===

func TestAdminSessionStore_PurgeExpired(t *testing.T) {
	store := newAdminSessionStore()

	// Issue a token with a 1ms TTL
	_, err := store.Issue("alice", 1*time.Millisecond)
	require.NoError(t, err)

	// Sleep 5ms
	time.Sleep(5 * time.Millisecond)

	// Issue another
	_, err = store.Issue("bob", 1*time.Hour)
	require.NoError(t, err)

	// Internal map size must be 1
	store.mu.RLock()
	mapSize := len(store.sessions)
	store.mu.RUnlock()
	assert.Equal(t, 1, mapSize, "expired token should have been purged on next Issue")
}

func TestAdminSessionStore_PurgeRevoked(t *testing.T) {
	store := newAdminSessionStore()

	tok1, err := store.Issue("alice", 1*time.Hour)
	require.NoError(t, err)

	store.Revoke(tok1)

	// Reset lastPurge to allow immediate purge
	store.mu.Lock()
	store.lastPurge = time.Time{}
	store.mu.Unlock()

	_, err = store.Issue("bob", 1*time.Hour)
	require.NoError(t, err)

	store.mu.RLock()
	mapSize := len(store.sessions)
	store.mu.RUnlock()
	assert.Equal(t, 1, mapSize, "revoked token should be purged by purgeExpiredLocked")
}

func TestLoginLimiter_OpportunisticCleanup(t *testing.T) {
	limiter := newLoginLimiter()
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	// Populate 10001 expired entries
	for i := 0; i <= 10000; i++ {
		limiter.failures[fmt.Sprintf("ip:10.0.%d.%d", i/256, i%256)] = &failureWindow{
			count:   1,
			resetAt: now.Add(-1 * time.Minute), // expired
		}
	}
	assert.Greater(t, len(limiter.failures), 10000)

	// Next failure triggers opportunistic cleanup
	limiter.fail("ip:127.0.0.1")
	assert.Equal(t, 1, len(limiter.failures))
}

func TestHandleLogin_RemoteAddrFallback(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")

	router := NewRouter(core.NewRegistry())

	mockNow := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	router.loginLimiter.now = func() time.Time { return mockNow }

	// RemoteAddr without port
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"wrong"}`))
		req.RemoteAddr = "10.0.0.2" // No port
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.handleLogin(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	}

	// 6th attempt should be blocked based on "ip:10.0.0.2"
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`))
	req.RemoteAddr = "10.0.0.2"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.handleLogin(rec, req)
	require.Equal(t, http.StatusTooManyRequests, rec.Code)
}
