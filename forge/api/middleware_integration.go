package api

import (
	"net/http"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/exceptions"
)

// APIMiddleware creates middleware for API requests
func APIMiddleware() core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Request is already wrapped by core if needed
			_ = core.NewRequest(r)

			// Recover from panics
			defer func() {
				if err := recover(); err != nil {
					// Convert panic to exception
					if apiErr, ok := err.(*exceptions.APIException); ok {
						exceptions.HandleExceptionHTTP(w, r, apiErr, nil)
					} else {
						exceptions.HandleExceptionHTTP(w, r, exceptions.NewAPIException(
							http.StatusInternalServerError,
							"internal_error",
							"Internal server error",
							nil,
						), nil)
					}
				}
			}()

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// AuthenticationMiddleware creates middleware for authentication
func AuthenticationMiddleware(authClasses []authentication.Authentication) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(authClasses) > 0 {
				result, err := authentication.AuthenticateRequest(r, authClasses)
				if err != nil {
					exceptions.HandleExceptionHTTP(w, r, exceptions.NewAuthenticationFailed(err.Error()), nil)
					return
				}

				if result != nil {
					authentication.SetUserOnRequest(r, result.User)
					authentication.SetAuthOnRequest(r, result.Auth)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
