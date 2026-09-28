package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	apiReq := NewRequest(req)

	assert.NotNil(t, apiReq)
	assert.Equal(t, req, apiReq.Request)
}

func TestRequest_User(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	apiReq := NewRequest(req)

	user := "test-user"
	apiReq.SetUser(user)

	assert.Equal(t, user, apiReq.User())
}

func TestRequest_Auth(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	apiReq := NewRequest(req)

	auth := "token123"
	apiReq.SetAuth(auth)

	assert.Equal(t, auth, apiReq.Auth())
}

func TestRequest_IsAuthenticated(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	apiReq := NewRequest(req)

	assert.False(t, apiReq.IsAuthenticated())

	apiReq.SetUser("user")
	assert.True(t, apiReq.IsAuthenticated())
}

func TestNewResponse(t *testing.T) {
	w := httptest.NewRecorder()
	apiResp := NewResponse(w)

	assert.NotNil(t, apiResp)
	assert.Equal(t, w, apiResp.ResponseWriter)
}

func TestResponse_Status(t *testing.T) {
	w := httptest.NewRecorder()
	apiResp := NewResponse(w)

	apiResp.Status(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, apiResp.statusCode)
}

func TestResponse_Header(t *testing.T) {
	w := httptest.NewRecorder()
	apiResp := NewResponse(w)

	apiResp.Header("X-Custom", "value")

	// Headers are stored, will be set when JSON is called
	assert.Equal(t, "value", apiResp.headers["X-Custom"])
}

func TestResponse_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	apiResp := NewResponse(w)

	data := map[string]interface{}{
		"name": "John",
	}

	err := apiResp.Status(http.StatusOK).JSON(data)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestResponse_Error(t *testing.T) {
	w := httptest.NewRecorder()
	apiResp := NewResponse(w)

	err := apiResp.Error(http.StatusBadRequest, "Error message", nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestWithUser(t *testing.T) {
	ctx := context.Background()
	user := "test-user"

	ctx = WithUser(ctx, user)

	retrieved, ok := UserFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, user, retrieved)
}

func TestWithAuth(t *testing.T) {
	ctx := context.Background()
	auth := "token123"

	ctx = WithAuth(ctx, auth)

	retrieved, ok := AuthFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, auth, retrieved)
}

func TestWithViewSet(t *testing.T) {
	ctx := context.Background()
	viewset := "test-viewset"

	ctx = WithViewSet(ctx, viewset)

	retrieved, ok := ViewSetFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, viewset, retrieved)
}

func TestWithAction(t *testing.T) {
	ctx := context.Background()
	action := "list"

	ctx = WithAction(ctx, action)

	retrieved, ok := ActionFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, action, retrieved)
}

func TestUserFromContext_NotSet(t *testing.T) {
	ctx := context.Background()

	_, ok := UserFromContext(ctx)
	assert.False(t, ok)
}

func TestAuthFromContext_NotSet(t *testing.T) {
	ctx := context.Background()

	_, ok := AuthFromContext(ctx)
	assert.False(t, ok)
}
