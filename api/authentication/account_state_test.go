package authentication

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/identity/models"
)

type accountSessionLookup func(*http.Request) (interface{}, error)

func (f accountSessionLookup) GetUser(r *http.Request) (interface{}, error) { return f(r) }

func TestAuthenticatorsRejectUnavailableIdentityUsers(t *testing.T) {
	for _, state := range []struct {
		name   string
		active bool
		locked bool
	}{
		{name: "inactive", active: false},
		{name: "locked", active: true, locked: true},
	} {
		t.Run(state.name, func(t *testing.T) {
			user := &models.User{ID: 1, IsActive: state.active, IsLocked: state.locked}
			basic := NewBasicAuthentication(func(string, string) (interface{}, error) { return user, nil })
			basicReq := httptest.NewRequest("GET", "/", nil)
			basicReq.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte("user:pass")))
			if _, err := basic.Authenticate(basicReq); err == nil {
				t.Fatal("Basic authentication accepted unavailable user")
			}

			token := NewTokenAuthentication(func(string) (interface{}, error) { return user, nil })
			tokenReq := httptest.NewRequest("GET", "/", nil)
			tokenReq.Header.Set("Authorization", "Token token")
			if _, err := token.Authenticate(tokenReq); err == nil {
				t.Fatal("token authentication accepted unavailable user")
			}

			session := NewSessionAuthentication(accountSessionLookup(func(*http.Request) (interface{}, error) { return user, nil }))
			if _, err := session.Authenticate(httptest.NewRequest("GET", "/", nil)); err == nil {
				t.Fatal("session authentication accepted unavailable user")
			}

			apiKey := NewAPIKeyAuthentication(func(string) (interface{}, error) { return user, nil })
			apiKeyReq := httptest.NewRequest("GET", "/?api_key=key", nil)
			if _, err := apiKey.Authenticate(apiKeyReq); err == nil {
				t.Fatal("API key authentication accepted unavailable user")
			}

			jwt := NewJWTAuthentication([]byte("secret"), func(JWTClaims) (interface{}, error) { return user, nil })
			jwt.ValidateToken = func(string, []byte) (JWTClaims, error) { return JWTClaims{"sub": 1}, nil }
			jwtReq := httptest.NewRequest("GET", "/", nil)
			jwtReq.Header.Set("Authorization", "Bearer token")
			if _, err := jwt.Authenticate(jwtReq); err == nil {
				t.Fatal("JWT authentication accepted unavailable user")
			}
		})
	}
}
