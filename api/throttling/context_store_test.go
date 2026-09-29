package throttling

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ctxKey struct{}

// contextStore records the context its AllowContext receives.
type contextStore struct {
	got         context.Context
	err         error
	allowCalled bool
}

func (s *contextStore) Allow(string) (bool, time.Duration) {
	s.allowCalled = true
	return false, time.Minute
}

func (s *contextStore) AllowContext(ctx context.Context, _ string) (bool, time.Duration, error) {
	s.got = ctx
	if s.err != nil {
		return false, 0, s.err
	}
	return false, time.Minute, nil
}

func TestThrottle_ContextStoreGetsTheRequestContext(t *testing.T) {
	store := &contextStore{}
	throttle := NewAnonRateThrottle("1/min").WithStore(store)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "request"))

	allowed, retry, err := throttle.AllowRequest(req, nil)
	require.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, time.Minute, retry)
	assert.False(t, store.allowCalled, "a ContextStore is called through AllowContext")
	require.NotNil(t, store.got)
	assert.Equal(t, "request", store.got.Value(ctxKey{}), "the store must be bound to the request's context")
}

func TestThrottle_ContextStoreErrorAllowsTheRequest(t *testing.T) {
	store := &contextStore{err: errors.New("database unreachable")}
	throttle := NewUserRateThrottle("1/min").WithStore(store)

	allowed, retry, err := throttle.AllowRequest(httptest.NewRequest(http.MethodGet, "/", nil), nil)
	require.NoError(t, err)
	assert.True(t, allowed, "an unreachable store must not refuse traffic")
	assert.Zero(t, retry)
}

func TestThrottle_ContextStoreRefusesAGoneClient(t *testing.T) {
	store := &contextStore{err: context.Canceled}
	throttle := NewAnonRateThrottle("1/min").WithStore(store)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	allowed, _, err := throttle.AllowRequest(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx), nil)
	require.NoError(t, err)
	assert.False(t, allowed, "a client that disconnected mid-check is not let through uncounted")
}
