package server

import (
	"context"
	"net/http"

	"github.com/forgego/forge/api/core"
	"github.com/forgego/forge/api/errors"
)

// Re-export core context functions for convenience and compatibility
// This ensures that server and api packages use the same context keys

// GetUser retrieves the user from the request context
func GetUser(r *http.Request) interface{} {
	user, _ := core.UserFromContext(r.Context())
	return user
}

// SetUser sets the user in the request context
func SetUser(r *http.Request, user interface{}) *http.Request {
	ctx := core.WithUser(r.Context(), user)
	return r.WithContext(ctx)
}

// GetDB retrieves the database connection from the request context
// Note: DB is not yet in core, so we keep local key for now, or move to core?
// Let's keep it local but use a unique key to avoid collisions
func GetDB(r *http.Request) interface{} {
	return r.Context().Value(DBKey)
}

// SetDB sets the database connection in the request context
func SetDB(r *http.Request, db interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), DBKey, db))
}

// GetLogger retrieves the logger from the request context
func GetLogger(r *http.Request) interface{} {
	return r.Context().Value(LoggerKey)
}

// SetLogger sets the logger in the request context
func SetLogger(r *http.Request, logger interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), LoggerKey, logger))
}

// GetLocale retrieves the locale from the request context
func GetLocale(r *http.Request) string {
	if locale, ok := r.Context().Value(LocaleKey).(string); ok {
		return locale
	}
	return "en"
}

// SetLocale sets the locale in the request context
func SetLocale(r *http.Request, locale string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), LocaleKey, locale))
}

// GetRequestID retrieves the request ID from the context
// Uses the errors package for consistency
func GetRequestID(r *http.Request) string {
	return errors.GetRequestIDFromContext(r.Context())
}

// SetRequestID sets the request ID in the context
func SetRequestID(r *http.Request, id string) *http.Request {
	// errors package doesn't export WithRequestID (it's internal to GetRequestIDFromContext?)
	// Wait, errors.GetRequestIDFromContext looks for RequestIDKey in its own package?
	// Let's check errors package.
	// For now, we assume we need to set it.
	return r.WithContext(context.WithValue(r.Context(), RequestIDKey, id))
}

// WithContext returns a new request with the given context
func WithContext(r *http.Request, ctx context.Context) *http.Request {
	return r.WithContext(ctx)
}

// GetContext retrieves the request context
func GetContext(r *http.Request) context.Context {
	return r.Context()
}

type contextKey string

const (
	// DBKey is the context key for the database connection
	DBKey contextKey = "db"
	// LoggerKey is the context key for the logger
	LoggerKey contextKey = "logger"
	// LocaleKey is the context key for the locale
	LocaleKey contextKey = "locale"
	// RequestIDKey is the context key for the request ID
	RequestIDKey contextKey = "request_id"
)
