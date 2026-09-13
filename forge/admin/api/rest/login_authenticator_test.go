package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuthenticator struct {
	authenticateFunc func(ctx context.Context, username, password string) (string, error)
}

func (f *fakeAuthenticator) AuthenticateAdmin(ctx context.Context, username, password string) (string, error) {
	if f.authenticateFunc != nil {
		return f.authenticateFunc(ctx, username, password)
	}
	return "", ErrInvalidLogin
}

type customDeniedError struct{}

func (e *customDeniedError) Error() string      { return "custom denied" }
func (e *customDeniedError) InvalidLogin() bool { return true }

func TestHandleLogin_AuthenticatorSuccess(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "")
	t.Setenv("FORGE_ADMIN_PASSWORD", "")

	router := NewRouter(core.NewRegistry())
	router.SetLoginAuthenticator(&fakeAuthenticator{
		authenticateFunc: func(ctx context.Context, username, password string) (string, error) {
			if username == "staffuser" && password == "correctpassword" {
				return "staffuser", nil
			}
			return "", ErrInvalidLogin
		},
	})

	body := `{"username":"staffuser","password":"correctpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))

	token, ok := res["token"].(string)
	require.True(t, ok)
	require.NotEmpty(t, token)

	userObj, ok := res["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "staffuser", userObj["name"])
	assert.Equal(t, "superuser", userObj["role"])

	session, ok := router.sessions.Validate(token)
	require.True(t, ok)
	assert.Equal(t, "staffuser", session.Username)
}

func TestHandleLogin_AuthenticatorInvalid_FallsBackToEnvWhenSet(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "envadmin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "envsecret")

	router := NewRouter(core.NewRegistry())
	router.SetLoginAuthenticator(&fakeAuthenticator{
		authenticateFunc: func(ctx context.Context, username, password string) (string, error) {
			return "", ErrInvalidLogin
		},
	})

	body := `{"username":"envadmin","password":"envsecret"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	userObj, ok := res["user"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "envadmin", userObj["name"])
}

func TestHandleLogin_AuthenticatorInvalid_FallsBackToEnvWithWrongPassword(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "envadmin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "envsecret")

	router := NewRouter(core.NewRegistry())
	router.SetLoginAuthenticator(&fakeAuthenticator{
		authenticateFunc: func(ctx context.Context, username, password string) (string, error) {
			return "", ErrInvalidLogin
		},
	})

	body := `{"username":"envadmin","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	errObj, ok := res["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "invalid_credentials", errObj["code"])
	assert.Equal(t, "Invalid username or password", errObj["message"])
}

func TestHandleLogin_AuthenticatorInvalid_NoEnvConfigured(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "")
	t.Setenv("FORGE_ADMIN_PASSWORD", "")

	router := NewRouter(core.NewRegistry())
	router.SetLoginAuthenticator(&fakeAuthenticator{
		authenticateFunc: func(ctx context.Context, username, password string) (string, error) {
			return "", &customDeniedError{}
		},
	})

	body := `{"username":"someone","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	errObj, ok := res["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "invalid_credentials", errObj["code"])
	assert.Equal(t, "Invalid username or password", errObj["message"])
}

func TestHandleLogin_AuthenticatorInfraError(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "")
	t.Setenv("FORGE_ADMIN_PASSWORD", "")

	router := NewRouter(core.NewRegistry())
	router.SetLoginAuthenticator(&fakeAuthenticator{
		authenticateFunc: func(ctx context.Context, username, password string) (string, error) {
			return "", errors.New("db connection failure")
		},
	})

	body := `{"username":"anyone","password":"anypassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.RemoteAddr = "10.0.0.99:1234"
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// Verify limiter was NOT failed on infra error
	if router.loginLimiter != nil {
		blocked, _ := router.loginLimiter.blocked("ip:10.0.0.99")
		assert.False(t, blocked)
		router.loginLimiter.mu.Lock()
		_, exists := router.loginLimiter.failures["ip:10.0.0.99"]
		router.loginLimiter.mu.Unlock()
		assert.False(t, exists, "infra error should not record limiter failure")
	}
}

func TestHandleLogin_NoAuthenticatorAndNoEnv(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "")
	t.Setenv("FORGE_ADMIN_PASSWORD", "")

	router := NewRouter(core.NewRegistry())

	body := `{"username":"admin","password":"password"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.handleLogin(rec, req)
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var res map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	errObj, ok := res["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "admin_login_disabled", errObj["code"])
}
