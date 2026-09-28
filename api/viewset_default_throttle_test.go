package api

import (
	"net/http"
	"testing"

	"github.com/forgego/forge/api/throttling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseViewSetDefaultThrottles(t *testing.T) {
	saved := GetSettings()
	t.Cleanup(func() { SetSettings(saved) })
	SetSettings(DefaultSettings())

	vs := newAccessCheckViewSet()
	require.Nil(t, vs.Throttles)

	SetDefaultThrottles(denyingThrottle{})

	response := serveAccessCheckRequest(vs, http.MethodGet, "/api/items/")
	assert.Equal(t, http.StatusTooManyRequests, response.Code, response.Body.String())

	// Non-nil empty slice disables throttling
	vsEmpty := newAccessCheckViewSet()
	vsEmpty.Throttles = []throttling.Throttle{}
	respEmpty := serveAccessCheckRequest(vsEmpty, http.MethodGet, "/api/items/")
	assert.Equal(t, http.StatusOK, respEmpty.Code, respEmpty.Body.String())
}
