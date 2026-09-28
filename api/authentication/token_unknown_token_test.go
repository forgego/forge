package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type w0TestUser struct {
	ID string
}

func TestTokenAuthentication_UnknownToken(t *testing.T) {
	validUser := &w0TestUser{ID: "u123"}
	tokenAuth := NewTokenAuthentication(func(token string) (interface{}, error) {
		if token == "valid-token" {
			return validUser, nil
		}
		return nil, nil // unknown token
	})

	tests := []struct {
		name       string
		authHeader string
		expectErr  bool
		expectUser interface{}
	}{
		{
			name:       "unknown token returns error",
			authHeader: "Token unknown-token",
			expectErr:  true,
		},
		{
			name:       "valid token returns user",
			authHeader: "Token valid-token",
			expectErr:  false,
			expectUser: validUser,
		},
		{
			name:       "no header returns nil, nil",
			authHeader: "",
			expectErr:  false,
			expectUser: nil,
		},
		{
			name:       "Bearer scheme returns nil, nil",
			authHeader: "Bearer some-token",
			expectErr:  false,
			expectUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			res, err := tokenAuth.Authenticate(req)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, res)

				// AuthenticateRequest with [TokenAuthentication] returns that error
				reqAll := httptest.NewRequest(http.MethodGet, "/test", nil)
				reqAll.Header.Set("Authorization", tt.authHeader)
				allRes, allErr := AuthenticateRequest(reqAll, []Authentication{tokenAuth})
				assert.Error(t, allErr)
				assert.Nil(t, allRes)
				assert.Equal(t, err, allErr)
			} else if tt.expectUser != nil {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, tt.expectUser, res.User)
			} else {
				// No header or Bearer scheme
				assert.NoError(t, err)
				assert.Nil(t, res)

				// AuthenticateRequest should fall through (nil, nil)
				reqAll := httptest.NewRequest(http.MethodGet, "/test", nil)
				if tt.authHeader != "" {
					reqAll.Header.Set("Authorization", tt.authHeader)
				}
				allRes, allErr := AuthenticateRequest(reqAll, []Authentication{tokenAuth})
				assert.NoError(t, allErr)
				assert.Nil(t, allRes)
			}
		})
	}
}
