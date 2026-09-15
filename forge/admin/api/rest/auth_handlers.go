package rest

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func loginKeys(remoteAddr, username string) (string, string) {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		ip = remoteAddr
	}
	return "ip:" + ip, "user:" + strings.ToLower(username)
}

func (r *Router) checkLoginRateLimit(w http.ResponseWriter, ipKey, userKey string) bool {
	if r.loginLimiter == nil {
		return true
	}
	blockedIP, remIP := r.loginLimiter.blocked(ipKey)
	blockedUser, remUser := r.loginLimiter.blocked(userKey)
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

func (r *Router) issueAdminSession(w http.ResponseWriter, username string) {
	token, err := r.sessions.Issue(username, 24*time.Hour)
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

	ipKey, userKey := loginKeys(req.RemoteAddr, username)
	if !r.checkLoginRateLimit(w, ipKey, userKey) {
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
		if r.loginLimiter != nil {
			r.loginLimiter.fail(ipKey)
			r.loginLimiter.fail(userKey)
		}
		respondError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password", nil)
		return
	}

	if r.loginLimiter != nil {
		r.loginLimiter.success(ipKey)
		r.loginLimiter.success(userKey)
	}
	r.issueAdminSession(w, canonicalUser)
}

// handleLogout revokes the current session token
func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	token, err := bearerToken(req.Header.Get("Authorization"))
	if err != nil {
		respondError(w, http.StatusUnauthorized, "authentication_required", "Authentication required", nil)
		return
	}

	r.sessions.Revoke(token)
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
