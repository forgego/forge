package authentication

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"hash"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthenticationAcceptsValidSignedToken(t *testing.T) {
	secret := []byte("test-secret-that-is-long-enough")
	lookupCalls := 0
	auth := NewJWTAuthentication(secret, func(claims JWTClaims) (interface{}, error) {
		lookupCalls++
		assert.Equal(t, "user-1", claims["sub"])
		return "user-1", nil
	})
	token := signedHS256Token(t, secret, map[string]interface{}{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	result, err := auth.Authenticate(bearerRequest(token))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "user-1", result.User)
	assert.Equal(t, token, result.Auth)
	assert.Equal(t, 1, lookupCalls)
}

func TestJWTAuthenticationRejectsInvalidTokensBeforeUserLookup(t *testing.T) {
	secret := []byte("test-secret-that-is-long-enough")
	validToken := signedHS256Token(t, secret, map[string]interface{}{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tamperedPayload := tamperJWTPayload(t, validToken, map[string]interface{}{
		"sub": "attacker",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	expiredToken := signedHS256Token(t, secret, map[string]interface{}{
		"sub": "user-1",
		"exp": time.Now().Add(-time.Second).Unix(),
	})
	noneToken := unsignedToken(t, map[string]interface{}{
		"sub": "attacker",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	nonAllowedAlgorithmToken := signedToken(t, secret, "HS512", map[string]interface{}{
		"sub": "attacker",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, sha512.New)
	invalidClaimsToken := signedHS256Token(t, secret, map[string]interface{}{
		"sub": "user-1",
		"exp": "not-a-unix-timestamp",
	})

	for name, token := range map[string]string{
		"tampered payload":      tamperedPayload,
		"expired token":         expiredToken,
		"alg none":              noneToken,
		"non-allowed algorithm": nonAllowedAlgorithmToken,
		"malformed claims":      invalidClaimsToken,
	} {
		t.Run(name, func(t *testing.T) {
			lookupCalls := 0
			auth := NewJWTAuthentication(secret, func(JWTClaims) (interface{}, error) {
				lookupCalls++
				return "unexpected", nil
			})

			result, err := auth.Authenticate(bearerRequest(token))

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Zero(t, lookupCalls)
		})
	}
}

func TestJWTAuthenticationRejectsMalformedToken(t *testing.T) {
	lookupCalls := 0
	auth := NewJWTAuthentication([]byte("test-secret-that-is-long-enough"), func(JWTClaims) (interface{}, error) {
		lookupCalls++
		return "unexpected", nil
	})

	result, err := auth.Authenticate(bearerRequest("not-a-jwt"))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, lookupCalls)
}

func TestJWTAuthenticationRejectsNullClaimsBeforeUserLookup(t *testing.T) {
	secret := []byte("test-secret-that-is-long-enough")
	lookupCalls := 0
	auth := NewJWTAuthentication(secret, func(JWTClaims) (interface{}, error) {
		lookupCalls++
		return "unexpected", nil
	})
	token := signedRawHS256Token(t, secret, []byte("null"))

	result, err := auth.Authenticate(bearerRequest(token))

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Zero(t, lookupCalls)
}

func signedRawHS256Token(t *testing.T, secret, payload []byte) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString(mustJSON(t, map[string]string{"alg": "HS256", "typ": "JWT"}))
	signingInput := header + "." + base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, secret)
	_, err := mac.Write([]byte(signingInput))
	require.NoError(t, err)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func bearerRequest(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func signedHS256Token(t *testing.T, secret []byte, claims map[string]interface{}) string {
	t.Helper()
	return signedToken(t, secret, "HS256", claims, sha256.New)
}

func signedToken(t *testing.T, secret []byte, algorithm string, claims map[string]interface{}, hash func() hash.Hash) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString(mustJSON(t, map[string]string{"alg": algorithm, "typ": "JWT"}))
	payload := base64.RawURLEncoding.EncodeToString(mustJSON(t, claims))
	signingInput := header + "." + payload

	mac := hmac.New(hash, secret)
	_, err := mac.Write([]byte(signingInput))
	require.NoError(t, err)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func unsignedToken(t *testing.T, claims map[string]interface{}) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString(mustJSON(t, map[string]string{"alg": "none", "typ": "JWT"}))
	payload := base64.RawURLEncoding.EncodeToString(mustJSON(t, claims))
	return header + "." + payload + "."
}

func tamperJWTPayload(t *testing.T, token string, claims map[string]interface{}) string {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	parts[1] = base64.RawURLEncoding.EncodeToString(mustJSON(t, claims))
	return strings.Join(parts, ".")
}

func mustJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return encoded
}
