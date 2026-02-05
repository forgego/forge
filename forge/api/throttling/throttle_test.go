package throttling

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache()

	// Set value
	err := cache.Set("test_key", 42, time.Minute)
	require.NoError(t, err)

	// Get value
	value, err := cache.GetInt("test_key")
	require.NoError(t, err)
	assert.Equal(t, 42, value)
}

func TestMemoryCache_Expiration(t *testing.T) {
	cache := NewMemoryCache()

	// Set with short TTL
	err := cache.Set("test_key", 42, 100*time.Millisecond)
	require.NoError(t, err)

	// Should exist immediately
	value, err := cache.GetInt("test_key")
	require.NoError(t, err)
	assert.Equal(t, 42, value)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	value, err = cache.GetInt("test_key")
	require.NoError(t, err)
	assert.Equal(t, 0, value) // Returns 0 when not found
}

func TestMemoryCache_GetTTL(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set("test_key", 42, time.Minute)
	require.NoError(t, err)

	ttl := cache.GetTTL("test_key")
	assert.True(t, ttl > 50*time.Second) // Should be close to 1 minute
	assert.True(t, ttl <= time.Minute)
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set("test_key", 42, time.Minute)
	cache.Delete("test_key")

	value, err := cache.GetInt("test_key")
	require.NoError(t, err)
	assert.Equal(t, 0, value)
}

func TestMemoryCache_Exists(t *testing.T) {
	cache := NewMemoryCache()

	// Check non-existent key
	_, err := cache.GetInt("test_key")
	require.NoError(t, err)
	val, _ := cache.GetInt("test_key")
	assert.Equal(t, 0, val)

	cache.Set("test_key", 42, time.Minute)
	val, _ = cache.GetInt("test_key")
	assert.Equal(t, 42, val)

	cache.Delete("test_key")
	val, _ = cache.GetInt("test_key")
	assert.Equal(t, 0, val)
}

func TestParseRate(t *testing.T) {
	tests := []struct {
		rate     string
		limit    int
		duration time.Duration
		hasError bool
	}{
		{"100/hour", 100, time.Hour, false},
		{"1000/day", 1000, 24 * time.Hour, false},
		{"50/minute", 50, time.Minute, false},
		{"10/second", 10, time.Second, false},
		{"invalid", 0, 0, true},
		{"100", 0, 0, true},
		{"100/invalid", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.rate, func(t *testing.T) {
			limit, duration, err := parseRate(tt.rate)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.limit, limit)
				assert.Equal(t, tt.duration, duration)
			}
		})
	}
}

func TestAnonRateThrottle_AllowRequest(t *testing.T) {
	cache := NewMemoryCache()
	throttle := NewAnonRateThrottle("2/hour", cache)
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{}

	// First request should be allowed
	allowed, wait, err := throttle.AllowRequest(req, view)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, time.Duration(0), wait)

	// Second request should be allowed
	allowed, wait, err = throttle.AllowRequest(req, view)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Third request should be throttled
	allowed, wait, err = throttle.AllowRequest(req, view)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.True(t, wait > 0)
}

func TestAnonRateThrottle_GetScope(t *testing.T) {
	throttle := NewAnonRateThrottle("100/hour", nil)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	scope := throttle.GetScope(req, nil)
	assert.NotEmpty(t, scope)
}

func TestUserRateThrottle_AllowRequest_NotAuthenticated(t *testing.T) {
	cache := NewMemoryCache()
	throttle := NewUserRateThrottle("100/hour", cache)
	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{}

	// Not authenticated, should allow (let AnonRateThrottle handle it)
	allowed, wait, err := throttle.AllowRequest(req, view)
	require.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, time.Duration(0), wait)
}

func TestCheckThrottles_AllPass(t *testing.T) {
	cache := NewMemoryCache()
	throttles := []Throttle{
		NewAnonRateThrottle("100/hour", cache),
		NewUserRateThrottle("1000/day", cache),
	}

	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{}

	err := CheckThrottles(req, view, throttles)
	assert.NoError(t, err)
}

func TestCheckThrottles_Throttled(t *testing.T) {
	cache := NewMemoryCache()
	throttles := []Throttle{
		NewAnonRateThrottle("1/hour", cache), // Very strict
	}

	req := httptest.NewRequest("GET", "/test", nil)
	view := &MockViewSet{}

	// First request
	err := CheckThrottles(req, view, throttles)
	assert.NoError(t, err)

	// Second request should be throttled
	err = CheckThrottles(req, view, throttles)
	assert.Error(t, err)
	assert.IsType(t, &ThrottledError{}, err)
}

func TestThrottledError(t *testing.T) {
	err := NewThrottledError(5 * time.Minute)
	assert.Equal(t, "Request was throttled", err.Error())
	assert.Equal(t, 5*time.Minute, err.WaitDuration)
}

// MockViewSet for testing
type MockViewSet struct{}

func (m *MockViewSet) GetAction() string     { return "list" }
func (m *MockViewSet) GetDetail() bool       { return false }
func (m *MockViewSet) GetModel() interface{} { return nil }
