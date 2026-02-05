package handlers

import (
	"context"
	"net/http"

	forgehttp "github.com/forgego/forge/server"
	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/serializers"
	"github.com/forgego/forge/identity/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService     service.AuthService
	userService     service.UserService
	backendRegistry backends.BackendRegistry
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	authService service.AuthService,
	userService service.UserService,
	backendRegistry backends.BackendRegistry,
) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		userService:     userService,
		backendRegistry: backendRegistry,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Deserialize and validate
	serializer := serializers.NewRegisterSerializer(data)
	if err := serializer.Validate(); err != nil {
		forgehttp.SendJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": serializer.Errors(),
		})
		return
	}

	// Register user
	req := &service.RegisterRequest{
		Username: serializer.GetUsername(),
		Email:    serializer.GetEmail(),
		Password: serializer.GetPassword(),
	}

	user, err := h.userService.Register(ctx, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrEmailExists || err == service.ErrUsernameExists {
			status = http.StatusConflict
		}
		forgehttp.SendError(w, status, err.Error())
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	forgehttp.SendJSON(w, http.StatusCreated, response)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Deserialize and validate
	serializer := serializers.NewLoginSerializer(data)
	if err := serializer.Validate(); err != nil {
		forgehttp.SendJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": serializer.Errors(),
		})
		return
	}

	// Authenticate
	req := &service.AuthenticateRequest{
		UsernameOrEmail: serializer.GetUsernameOrEmail(),
		Password:        serializer.GetPassword(),
		RememberMe:      serializer.GetRememberMe(),
		IPAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
	}

	user, err := h.authService.Authenticate(ctx, req)
	if err != nil {
		status := http.StatusUnauthorized
		if err == service.ErrUserInactive || err == service.ErrUserLocked {
			status = http.StatusForbidden
		}
		forgehttp.SendError(w, status, err.Error())
		return
	}

	// Create session
	sessionReq := &service.CreateSessionRequest{
		UserID:     user.ID,
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
		RememberMe: serializer.GetRememberMe(),
	}

	session, err := h.authService.CreateSession(ctx, user.ID, sessionReq)
	if err != nil {
		forgehttp.SendError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	response["session_key"] = session.SessionKey
	forgehttp.SendJSON(w, http.StatusOK, response)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get session key from header or body
	sessionKey := r.Header.Get("X-Session-Key")
	if sessionKey == "" {
		var data map[string]interface{}
		if err := forgehttp.GetJSON(r, &data); err == nil {
			if key, ok := data["session_key"].(string); ok {
				sessionKey = key
			}
		}
	}

	if sessionKey == "" {
		forgehttp.SendError(w, http.StatusBadRequest, "Session key required")
		return
	}

	// Logout
	if err := h.authService.Logout(ctx, sessionKey); err != nil {
		forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (set by middleware)
	user, ok := GetUserFromContext(ctx)
	if !ok {
		forgehttp.SendError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	forgehttp.SendJSON(w, http.StatusOK, response)
}

// GetUserFromContext retrieves user from context
// This will be set by authentication middleware
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	// Try to get user from context (set by middleware)
	if user, ok := ctx.Value("user").(*models.User); ok {
		return user, true
	}

	// Try API core context
	if user, ok := ctx.Value("api_user").(*models.User); ok {
		return user, true
	}

	return nil, false
}
