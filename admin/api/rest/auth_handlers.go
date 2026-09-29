package rest

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/forgego/forge/netutil"
)

func decodeLoginPayload(req *http.Request) (string, string, error) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil || decoder.More() {
		return "", "", errors.New("invalid login payload")
	}
	return strings.TrimSpace(payload.Username), payload.Password, nil
}

// loginKeys returns the limiter keys for a login attempt. The client IP
// honors X-Forwarded-For / X-Real-IP only when the TCP peer is a configured
// trusted proxy (server.trusted_proxies); otherwise it is the peer address.
// Without that, every client behind a reverse proxy would share one key.
func loginKeys(req *http.Request, username string) (string, string) {
	ip := netutil.ClientIP(req, netutil.TrustedProxies())
	return "ip:" + ip, "user:" + strings.ToLower(username)
}

func (r *Router) checkLoginRateLimit(w http.ResponseWriter, req *http.Request, ipKey, userKey string) bool {
	if r.loginAttempts == nil {
		return true
	}
	ctx := req.Context()
	blockedIP, remIP, errIP := r.loginAttempts.Blocked(ctx, ipKey)
	blockedUser, remUser, errUser := r.loginAttempts.Blocked(ctx, userKey)
	if errIP != nil || errUser != nil {
		// Without the counts the lockout cannot be enforced, so the
		// attempt is refused rather than let through unchecked.
		log.Printf("forge/admin: login lockout store: %v", errors.Join(errIP, errUser))
		respondError(w, http.StatusServiceUnavailable, "login_unavailable", "Login is temporarily unavailable", nil)
		return false
	}
	if !blockedIP && !blockedUser {
		return true
	}
	longer := remIP
	if blockedUser && remUser > longer {
		longer = remUser
	}
	secs := int(math.Ceil(longer.Seconds()))
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	respondError(w, http.StatusTooManyRequests, "too_many_attempts", "Too many failed login attempts. Try again later.", nil)
	return false
}

// recordLogin records a failed or successful login for each limiter key.
// A store error is logged: the login's outcome stands either way.
func (r *Router) recordLogin(ctx context.Context, failed bool, keys ...string) {
	if r.loginAttempts == nil {
		return
	}
	for _, key := range keys {
		var err error
		if failed {
			err = r.loginAttempts.Failed(ctx, key)
		} else {
			err = r.loginAttempts.Succeeded(ctx, key)
		}
		if err != nil {
			log.Printf("forge/admin: login lockout store: %v", err)
		}
	}
}

func (r *Router) authenticateAdmin(ctx context.Context, username, password string) (string, error) {
	expUser, expPass := adminCredentials()
	hasEnv := expUser != "" && expPass != ""

	if r.authenticator == nil && !hasEnv {
		return "", errAdminLoginDisabled
	}

	if r.authenticator != nil {
		canonicalUser, err := r.authenticator.AuthenticateAdmin(ctx, username, password)
		if err == nil {
			return canonicalUser, nil
		}
		if !isInvalidLogin(err) {
			return "", err
		}
	}

	if hasEnv && secureEqual(username, expUser) && secureEqual(password, expPass) {
		return expUser, nil
	}

	return "", ErrInvalidLogin
}

func (r *Router) issueAdminSession(w http.ResponseWriter, req *http.Request, username string) {
	token, err := r.tokens.IssueToken(req.Context(), username, 24*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "login_failed", "Could not create session token", nil)
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]string{
			"name": username,
			"role": "superuser",
		},
		"expires_at": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	})
}

// handleLogin handles admin login
func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	username, password, err := decodeLoginPayload(req)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid_body", "Invalid login payload", nil)
		return
	}

	ipKey, userKey := loginKeys(req, username)
	if !r.checkLoginRateLimit(w, req, ipKey, userKey) {
		return
	}
	if username == "" || password == "" {
		respondError(w, http.StatusBadRequest, "invalid_credentials", "Username and password are required", nil)
		return
	}

	canonicalUser, err := r.authenticateAdmin(req.Context(), username, password)
	if errors.Is(err, errAdminLoginDisabled) {
		respondError(w, http.StatusServiceUnavailable, "admin_login_disabled", "Admin login is not configured (set FORGE_ADMIN_USERNAME and FORGE_ADMIN_PASSWORD)", nil)
		return
	}
	if err != nil && !isInvalidLogin(err) {
		respondError(w, http.StatusInternalServerError, "login_failed", "Authentication failed", nil)
		return
	}
	if err != nil {
		r.recordLogin(req.Context(), true, ipKey, userKey)
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password", nil)
		return
	}

	r.recordLogin(req.Context(), false, ipKey, userKey)
	r.issueAdminSession(w, req, canonicalUser)
}

// handleLogout revokes the current session token
func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	token, err := bearerToken(req.Header.Get("Authorization"))
	if err != nil {
		respondError(w, http.StatusUnauthorized, "authentication_required", "Authentication required", nil)
		return
	}

	if err := r.tokens.RevokeToken(req.Context(), token); err != nil {
		respondError(w, http.StatusServiceUnavailable, "token_store_unavailable", "Could not revoke the token", nil)
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "logged out successfully",
	})
}

func adminCredentials() (string, string) {
	// No defaults: admin login stays disabled until both variables are set.
	username := strings.TrimSpace(os.Getenv("FORGE_ADMIN_USERNAME"))
	password := strings.TrimRight(os.Getenv("FORGE_ADMIN_PASSWORD"), "\r\n")

	return username, password
}

func secureEqual(a, b string) bool {
	// Hash only to normalize lengths before constant-time comparison. This is not
	// password storage: the configured credential remains the source of truth.
	ah := sha256.Sum256([]byte(a))
	bh := sha256.Sum256([]byte(b))
	return subtle.ConstantTimeCompare(ah[:], bh[:]) == 1
}
