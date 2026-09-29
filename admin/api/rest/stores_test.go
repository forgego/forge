package rest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errStoreDown = errors.New("store down")

type failingStores struct{}

func (failingStores) IssueToken(context.Context, string, time.Duration) (string, error) {
	return "", errStoreDown
}
func (failingStores) ValidateToken(context.Context, string) (string, bool, error) {
	return "", false, errStoreDown
}
func (failingStores) RevokeToken(context.Context, string) error { return errStoreDown }
func (failingStores) ListSavedViews(context.Context, string, string) ([]SavedView, error) {
	return nil, errStoreDown
}
func (failingStores) SaveView(context.Context, string, string, SavedViewRequest) (SavedView, bool, error) {
	return SavedView{}, false, errStoreDown
}
func (failingStores) DeleteSavedView(context.Context, string, string, string) (bool, error) {
	return false, errStoreDown
}
func (failingStores) Blocked(context.Context, string) (bool, time.Duration, error) {
	return false, 0, errStoreDown
}
func (failingStores) Failed(context.Context, string) error    { return errStoreDown }
func (failingStores) Succeeded(context.Context, string) error { return errStoreDown }

type recordingTokens struct {
	issued map[string]string
}

func (r *recordingTokens) IssueToken(_ context.Context, username string, _ time.Duration) (string, error) {
	token := "tok-" + username
	r.issued[token] = username
	return token, nil
}
func (r *recordingTokens) ValidateToken(_ context.Context, token string) (string, bool, error) {
	username, ok := r.issued[token]
	return username, ok, nil
}
func (r *recordingTokens) RevokeToken(_ context.Context, token string) error {
	delete(r.issued, token)
	return nil
}

func serveRouter(router *Router) http.Handler {
	mux := chi.NewRouter()
	router.RegisterRoutes(mux)
	return mux
}

func TestRouterSetStores_HandlersUseTheConfiguredStores(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")
	tokens := &recordingTokens{issued: map[string]string{}}
	router := NewRouter(core.NewRegistry()).SetStores(Stores{Tokens: tokens})
	handler := serveRouter(router)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`)))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Contains(t, rec.Body.String(), `"token":"tok-admin"`)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	req.Header.Set("Authorization", "Bearer tok-admin")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	_, ok := router.sessions.Validate("tok-admin")
	assert.False(t, ok, "the in-memory token store is no longer used")
}

func TestRouterSetStores_StoreErrorsAreNotAuthenticationFailures(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "admin")
	t.Setenv("FORGE_ADMIN_PASSWORD", "secret")
	router := NewRouter(core.NewRegistry()).SetStores(Stores{
		Tokens: failingStores{}, SavedViews: failingStores{}, LoginAttempts: failingStores{},
	})
	handler := serveRouter(router)

	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	req.Header.Set("Authorization", "Bearer anything")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "an unreachable token store is not a 401")

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"admin","password":"secret"}`)))
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code, "without lockout counts login is refused, not let through")
}

func TestRouterSetStores_NilFieldsKeepMemoryDefaults(t *testing.T) {
	router := NewRouter(core.NewRegistry())
	router.SetStores(Stores{})
	assert.Same(t, router.sessions, router.tokens)
	assert.Same(t, router.views, router.savedViews)
	assert.Same(t, router.loginLimiter, router.loginAttempts)
}
