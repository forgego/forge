package authentication

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

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
	// CurrentKeyID is written into tokens signed by SignToken.
	CurrentKeyID string
	// VerificationKeys are previous keys accepted only through VerifyUntil.
	VerificationKeys []JWTVerificationKey
	// UserLookup is a function that looks up a user from JWT claims
	// Should return (user, nil) if found, (nil, error) if error
	UserLookup func(claims JWTClaims) (interface{}, error)
	// ValidateToken is a function to validate the JWT token
	// Should return (claims, nil) if valid, (nil, error) if invalid
	ValidateToken func(tokenString string, secretKey []byte) (JWTClaims, error)
}

// JWTVerificationKey is a previous JWT key accepted only until VerifyUntil.
// Key material is never included in errors or logs.
type JWTVerificationKey struct {
	ID          string
	Key         []byte
	VerifyUntil time.Time
}

// NewJWTAuthentication creates a new JWT authentication instance
func NewJWTAuthentication(secretKey []byte, userLookup func(claims JWTClaims) (interface{}, error)) *JWTAuthentication {
	return &JWTAuthentication{
		SecretKey:  secretKey,
		UserLookup: userLookup,
	}
}

// NewJWTAuthenticationWithKeys configures a current signing key and previous,
// verify-only keys. Previous keys are rejected at and after VerifyUntil.
func NewJWTAuthenticationWithKeys(currentKeyID string, currentKey []byte, previous []JWTVerificationKey, userLookup func(JWTClaims) (interface{}, error)) *JWTAuthentication {
	auth := NewJWTAuthentication(currentKey, userLookup)
	auth.CurrentKeyID = currentKeyID
	auth.VerificationKeys = append([]JWTVerificationKey(nil), previous...)
	return auth
}

// SignToken signs claims with the current key and its key ID. Previous keys
// configured for verification are never used to issue tokens.
func (a *JWTAuthentication) SignToken(claims JWTClaims) (string, error) {
	if len(a.SecretKey) == 0 {
		return "", errors.New("JWT secret key is not configured")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	if a.CurrentKeyID != "" {
		token.Header["kid"] = a.CurrentKeyID
	}
	return token.SignedString(a.SecretKey)
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

func validateJWTTokenWithKeys(tokenString string, currentID string, currentKey []byte, previous []JWTVerificationKey) (JWTClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid JWT token")
	}
	var header struct {
		KeyID string `json:"kid"`
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(decoded, &header) != nil {
		return nil, errors.New("invalid JWT header")
	}
	if header.KeyID == "" {
		if len(previous) > 0 || currentID != "" {
			return nil, errors.New("JWT key ID is required")
		}
		return validateJWTToken(tokenString, currentKey)
	}
	if header.KeyID == currentID {
		return validateJWTToken(tokenString, currentKey)
	}
	now := time.Now()
	for _, key := range previous {
		if key.ID == header.KeyID {
			if !now.Before(key.VerifyUntil) {
				return nil, errors.New("JWT signing key has expired")
			}
			return validateJWTToken(tokenString, key.Key)
		}
	}
	return nil, errors.New("unknown JWT key ID")
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

	var claims JWTClaims
	var err error
	if a.ValidateToken == nil && (a.CurrentKeyID != "" || len(a.VerificationKeys) > 0) {
		claims, err = validateJWTTokenWithKeys(tokenString, a.CurrentKeyID, a.SecretKey, a.VerificationKeys)
	} else {
		claims, err = validateFunc(tokenString, a.SecretKey)
	}
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
	if err := validateAccountState(user); err != nil {
		return nil, err
	}

	return NewAuthResult(user, tokenString), nil
}

// AuthenticateHeader returns the WWW-Authenticate header value
func (a *JWTAuthentication) AuthenticateHeader(r *http.Request) string {
	return "Bearer"
}
