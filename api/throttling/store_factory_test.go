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

type factoryCall struct {
	name   string
	limit  int
	window time.Duration
}

func TestSetDefaultStoreFactory_SharedStoreForThrottlesWithoutOne(t *testing.T) {
	t.Cleanup(func() { SetDefaultStoreFactory(nil) })

	// Created before the factory is set, as package-level throttles are.
	anon := NewAnonRateThrottle("2/min")
	user := NewUserRateThrottle("5/hour")
	explicit := NewAnonRateThrottle("2/min").WithStore(newFakeStore(100, 0))

	var mu sync.Mutex
	var calls []factoryCall
	shared := newFakeStore(1, 30*time.Second)
	SetDefaultStoreFactory(func(name string, limit int, window time.Duration) Store {
		mu.Lock()
		defer mu.Unlock()
		calls = append(calls, factoryCall{name, limit, window})
		return shared
	})

	anonymous := httptest.NewRequest(http.MethodGet, "/", nil)
	anonymous.RemoteAddr = "10.0.0.5:1234"
	require.NoError(t, CheckThrottles(anonymous, nil, []Throttle{anon}))
	err := CheckThrottles(anonymous, nil, []Throttle{anon})
	require.Error(t, err, "the factory's store decides")
	assert.Equal(t, 30*time.Second, err.(*ThrottledError).WaitDuration)

	authenticated := httptest.NewRequest(http.MethodGet, "/", nil)
	authentication.SetUserOnRequest(authenticated, &mockAuthUser{ID: "u1"})
	require.NoError(t, CheckThrottles(authenticated, nil, []Throttle{user}))

	for range 3 {
		require.NoError(t, CheckThrottles(anonymous, nil, []Throttle{explicit}), "WithStore wins over the factory")
	}

	assert.Equal(t, []factoryCall{{"anon/2/min", 2, time.Minute}, {"user/5/hour", 5, time.Hour}}, calls,
		"one store per throttle, named by scope and rate")
	assert.Equal(t, []string{"throttle_anon_10.0.0.5", "throttle_anon_10.0.0.5", "throttle_user_u1"}, shared.receivedKeys)

	// nil restores the in-memory counter.
	SetDefaultStoreFactory(nil)
	fresh := httptest.NewRequest(http.MethodGet, "/", nil)
	fresh.RemoteAddr = "10.0.0.6:1234"
	require.NoError(t, CheckThrottles(fresh, nil, []Throttle{anon}))
	require.NoError(t, CheckThrottles(fresh, nil, []Throttle{anon}))
	require.Error(t, CheckThrottles(fresh, nil, []Throttle{anon}), "memory counter: 2/min")
	assert.Len(t, shared.receivedKeys, 3)
}
