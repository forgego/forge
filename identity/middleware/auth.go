package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/identity/backends"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/service"
	forgehttp "github.com/forgego/forge/server"
	"github.com/gorilla/csrf"
)

// AuthenticationMiddleware authenticates requests
type AuthenticationMiddleware struct {
	backendRegistry backends.BackendRegistry
	sessionRepo     repository.SessionRepository
	userRepo        repository.UserRepository
	permissionSvc   service.PermissionService
	csrfWarningOnce sync.Once
}

// NewAuthenticationMiddleware creates a new authentication middleware
func NewAuthenticationMiddleware(
	backendRegistry backends.BackendRegistry,
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
) *AuthenticationMiddleware {
	return NewAuthenticationMiddlewareWithPermissionService(backendRegistry, sessionRepo, userRepo, nil)
}

// NewAuthenticationMiddlewareWithPermissionService creates a new authentication middleware
func NewAuthenticationMiddlewareWithPermissionService(
	backendRegistry backends.BackendRegistry,
	sessionRepo repository.SessionRepository,
	userRepo repository.UserRepository,
	permissionSvc service.PermissionService,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		backendRegistry: backendRegistry,
		sessionRepo:     sessionRepo,
		userRepo:        userRepo,
		permissionSvc:   permissionSvc,
	}
}

// RequireAuth middleware requires authentication
func (m *AuthenticationMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to authenticate
		user, err := m.authenticateRequest(ctx, r)
		if err != nil || user == nil {
			forgehttp.SendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Set user in context
		ctx = core.WithUser(ctx, user)
		*r = *r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware sets user if authenticated but doesn't require it
func (m *AuthenticationMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to authenticate (don't fail if not authenticated)
		user, _ := m.authenticateRequest(ctx, r)
		if user != nil {
			ctx = core.WithUser(ctx, user)
			*r = *r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

// authenticateRequest authenticates session cookies on unsafe methods only after CSRF validation ran.
// Header-based session keys and safe methods do not require CSRF validation.
func (m *AuthenticationMiddleware) authenticateRequest(ctx context.Context, r *http.Request) (*models.User, error) {
	// Try token-based authentication first (from Authorization header)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		token, err := extractTokenFromHeader(authHeader)
		if err == nil && token != "" {
			credentials := map[string]string{
				"token": token,
			}
			user, err := m.backendRegistry.Authenticate(ctx, credentials)
			if err == nil && user != nil {
				return user, nil
			}
		}
	}

	// Try session-based authentication
	sessionKey := r.Header.Get("X-Session-Key")
	cookieSession := false
	if sessionKey == "" {
		// Try to get from cookie
		cookie, err := r.Cookie("session_key")
		if err == nil {
			sessionKey = cookie.Value
			cookieSession = true
		}
	}

	if cookieSession && isUnsafeMethod(r.Method) && csrf.Token(r) == "" {
		m.csrfWarningOnce.Do(func() {
			slog.Warn("session cookie ignored on unsafe request: CSRF middleware is not mounted on this route", "method", r.Method, "path", sanitizeLogValue(r.URL.Path))
		})
		return nil, nil
	}

	if sessionKey != "" && m.sessionRepo != nil && m.userRepo != nil {
		session, err := m.sessionRepo.GetByKey(ctx, sessionKey)
		if err == nil && !session.IsExpired() {
			// Get user from session's user_id
			user, err := m.userRepo.GetByID(ctx, session.UserID)
			if err == nil && user != nil && user.IsActive && !user.IsLocked {
				return user, nil
			}
		}
	}

	return nil, nil
}

func sanitizeLogValue(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return false
	default:
		return true
	}
}

// RequirePermission middleware requires a specific permission
func (m *AuthenticationMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user from context
			user, ok := userFromContext(ctx)
			if !ok || user == nil {
				forgehttp.SendError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			// Check permission
			// Superusers have all permissions
			if user.IsSuperuser && user.IsActive {
				next.ServeHTTP(w, r)
				return
			}

			if !user.IsActive {
				forgehttp.SendError(w, http.StatusForbidden, "Permission denied")
				return
			}

			if m.permissionSvc != nil {
				hasPermission, err := m.permissionSvc.CheckPermission(ctx, user.ID, permission)
				if err != nil {
					forgehttp.SendError(w, http.StatusInternalServerError, "Permission check failed")
					return
				}
				if !hasPermission {
					forgehttp.SendError(w, http.StatusForbidden, "Permission denied")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Backward-compatible fallback for setups that haven't wired PermissionService yet.
			if !user.IsStaff {
				forgehttp.SendError(w, http.StatusForbidden, "Permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireStaff middleware requires staff status
func (m *AuthenticationMiddleware) RequireStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get user from context
		user, ok := userFromContext(ctx)
		if !ok || user == nil {
			forgehttp.SendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		if !user.IsStaff {
			forgehttp.SendError(w, http.StatusForbidden, "Staff access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireSuperuser middleware requires superuser status
func (m *AuthenticationMiddleware) RequireSuperuser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get user from context
		user, ok := userFromContext(ctx)
		if !ok || user == nil {
			forgehttp.SendError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		if !user.IsSuperuser {
			forgehttp.SendError(w, http.StatusForbidden, "Superuser access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func userFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := core.UserFromContext(ctx)
	if !ok {
		return nil, false
	}
	currentUser, ok := user.(*models.User)
	return currentUser, ok
}

// extractTokenFromHeader extracts token from Authorization header
func extractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid authorization header format")
	}

	if strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("unsupported authorization scheme")
	}

	return parts[1], nil
}
