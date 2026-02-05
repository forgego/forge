package authentication

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// JWTClaims represents JWT claims
type JWTClaims map[string]interface{}

// JWTAuthentication authenticates requests using JWT tokens
// Format: Authorization: Bearer <token>
// Note: This is a simplified JWT implementation. For production, use a proper JWT library.
type JWTAuthentication struct {
	// SecretKey is the secret key for signing/verifying tokens
	SecretKey []byte
	// UserLookup is a function that looks up a user from JWT claims
	// Should return (user, nil) if found, (nil, error) if error
	UserLookup func(claims JWTClaims) (interface{}, error)
	// ValidateToken is a function to validate the JWT token
	// Should return (claims, nil) if valid, (nil, error) if invalid
	ValidateToken func(tokenString string, secretKey []byte) (JWTClaims, error)
}

// NewJWTAuthentication creates a new JWT authentication instance
func NewJWTAuthentication(secretKey []byte, userLookup func(claims JWTClaims) (interface{}, error)) *JWTAuthentication {
	return &JWTAuthentication{
		SecretKey:     secretKey,
		UserLookup:    userLookup,
		ValidateToken: validateJWTToken, // Default validation
	}
}

// validateJWTToken validates a JWT token (simplified implementation)
// For production, use a proper JWT library like github.com/golang-jwt/jwt
func validateJWTToken(tokenString string, secretKey []byte) (JWTClaims, error) {
	// Split token into parts (header.payload.signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Decode payload (simplified - in production, verify signature)
	payload := parts[1]
	// Add padding if needed
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var claims JWTClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, errors.New("token expired")
		}
	}

	return claims, nil
}

// Authenticate attempts to authenticate using JWT from Authorization header
func (a *JWTAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	// Parse "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, nil
	}

	tokenString := parts[1]
	if tokenString == "" {
		return nil, nil
	}

	// Validate token
	validateFunc := a.ValidateToken
	if validateFunc == nil {
		validateFunc = validateJWTToken
	}

	claims, err := validateFunc(tokenString, a.SecretKey)
	if err != nil {
		return nil, err
	}

	// Lookup user from claims
	if a.UserLookup == nil {
		return nil, errors.New("user lookup function not configured")
	}

	user, err := a.UserLookup(claims)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return NewAuthResult(user, tokenString), nil
}

// AuthenticateHeader returns the WWW-Authenticate header value
func (a *JWTAuthentication) AuthenticateHeader(r *http.Request) string {
	return "Bearer"
}
