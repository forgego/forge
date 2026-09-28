package authentication

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

// validateJWTToken validates a JWT token signed with HS256.
func validateJWTToken(tokenString string, secretKey []byte) (JWTClaims, error) {
	if len(secretKey) == 0 {
		return nil, errors.New("JWT secret key is not configured")
	}

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("unexpected JWT signing method")
			}
			return secretKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, errors.New("invalid JWT token")
	}
	if err := validateJWTClaimsPayload(tokenString); err != nil {
		return nil, err
	}

	return JWTClaims(claims), nil
}

func validateJWTClaimsPayload(tokenString string) error {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return errors.New("invalid JWT token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid JWT claims payload")
	}

	var rawClaims map[string]json.RawMessage
	if err := json.Unmarshal(payload, &rawClaims); err != nil || rawClaims == nil {
		return errors.New("JWT claims must be a JSON object")
	}

	return nil
}

// Authenticate attempts to authenticate using JWT from Authorization header
func (a *JWTAuthentication) Authenticate(r *http.Request) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, nil
	}

	// A non-Bearer scheme is not applicable to this authenticator. A malformed
	// Bearer credential is applicable, but invalid, and must not fall through.
	parts := strings.Fields(authHeader)
	if len(parts) == 0 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, nil
	}
	if len(parts) != 2 || parts[1] == "" {
		return nil, errors.New("invalid Bearer authorization header")
	}

	tokenString := parts[1]

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
