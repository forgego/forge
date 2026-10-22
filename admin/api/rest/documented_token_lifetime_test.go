package rest

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
)

type lifetimeTokens struct{ ttl time.Duration }

func (s *lifetimeTokens) IssueToken(_ context.Context, _ string, ttl time.Duration) (string, error) {
	s.ttl = ttl
	return "sample-token", nil
}
func (*lifetimeTokens) ValidateToken(context.Context, string) (string, bool, error) {
	return "", false, nil
}
func (*lifetimeTokens) RevokeToken(context.Context, string) error { return nil }

// Successful admin login issues a token valid for 24 hours.
func TestAdminLoginIssues24HourBearerToken(t *testing.T) {
	t.Setenv("FORGE_ADMIN_USERNAME", "owner")
	t.Setenv("FORGE_ADMIN_PASSWORD", "synthetic-password")
	tokens := &lifetimeTokens{}
	r := NewRouter(core.NewRegistry()).SetStores(Stores{Tokens: tokens})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBufferString(`{"username":"owner","password":"synthetic-password"}`))
	r.handleLogin(w, req)
	if w.Code != http.StatusOK || tokens.ttl != 24*time.Hour {
		t.Fatalf("status=%d ttl=%s body=%s", w.Code, tokens.ttl, w.Body.String())
	}
}
