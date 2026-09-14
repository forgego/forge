package throttling

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/api/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	mu           sync.Mutex
	calls        map[string]int
	maxCalls     int
	retryAfter   time.Duration
	receivedKeys []string
}

func newFakeStore(maxCalls int, retryAfter time.Duration) *fakeStore {
	return &fakeStore{
		calls:      make(map[string]int),
		maxCalls:   maxCalls,
		retryAfter: retryAfter,
	}
}

func (s *fakeStore) Allow(key string) (bool, time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.receivedKeys = append(s.receivedKeys, key)
	s.calls[key]++
	if s.calls[key] > s.maxCalls {
		return false, s.retryAfter
	}
	return true, 0
}

func TestUserRateThrottle_WithStore_DeniesAfterNCalls(t *testing.T) {
	store := newFakeStore(2, 45*time.Second)
	throttle := NewUserRateThrottle("100/hour").WithStore(store)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	authentication.SetUserOnRequest(request, &mockAuthUser{ID: "user-99"})

	// First 2 calls allowed
	for i := 0; i < 2; i++ {
		err := CheckThrottles(request, nil, []Throttle{throttle})
		require.NoError(t, err)
	}

	// 3rd call denied by store
	err := CheckThrottles(request, nil, []Throttle{throttle})
	require.Error(t, err)

	throttledErr, ok := err.(*ThrottledError)
	require.True(t, ok)
	assert.Equal(t, 45*time.Second, throttledErr.WaitDuration)

	// Verify key received by store
	store.mu.Lock()
	defer store.mu.Unlock()
	require.Len(t, store.receivedKeys, 3)
	assert.Equal(t, "throttle_user_user-99", store.receivedKeys[0])
}

func TestAnonRateThrottle_WithStore_DeniesAfterNCalls(t *testing.T) {
	store := newFakeStore(1, 30*time.Second)
	throttle := NewAnonRateThrottleWithStore("100/hour", store)

	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	request.RemoteAddr = "192.0.2.55:9999"

	// 1st call allowed
	err := CheckThrottles(request, nil, []Throttle{throttle})
	require.NoError(t, err)

	// 2nd call denied
	err = CheckThrottles(request, nil, []Throttle{throttle})
	require.Error(t, err)

	throttledErr, ok := err.(*ThrottledError)
	require.True(t, ok)
	assert.Equal(t, 30*time.Second, throttledErr.WaitDuration)

	// Verify key received by store
	store.mu.Lock()
	defer store.mu.Unlock()
	require.Len(t, store.receivedKeys, 2)
	assert.Equal(t, "throttle_anon_192.0.2.55", store.receivedKeys[0])
}

func TestDefaultConstructors_AllowLimitThenDeny(t *testing.T) {
	// User throttle default constructor allows limit then denies
	userThrottle := NewUserRateThrottle("2/minute")

	userReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	authentication.SetUserOnRequest(userReq, &mockAuthUser{ID: "burst-user"})

	for i := 0; i < 2; i++ {
		err := CheckThrottles(userReq, nil, []Throttle{userThrottle})
		require.NoError(t, err)
	}
	err := CheckThrottles(userReq, nil, []Throttle{userThrottle})
	require.Error(t, err)
	throttled, ok := err.(*ThrottledError)
	require.True(t, ok)
	assert.Positive(t, throttled.WaitDuration)

	// Anon throttle default constructor allows limit then denies
	anonThrottle := NewAnonRateThrottle("2/minute")

	anonReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	anonReq.RemoteAddr = "192.0.2.88:1234"

	for i := 0; i < 2; i++ {
		err := CheckThrottles(anonReq, nil, []Throttle{anonThrottle})
		require.NoError(t, err)
	}
	err = CheckThrottles(anonReq, nil, []Throttle{anonThrottle})
	require.Error(t, err)
	throttled, ok = err.(*ThrottledError)
	require.True(t, ok)
	assert.Positive(t, throttled.WaitDuration)
}
