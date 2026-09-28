package core

import (
	"context"
)

// Context keys for storing API-specific values
type contextKey string

const (
	// UserKey is the context key for the authenticated user
	UserKey contextKey = "api_user"
	// AuthKey is the context key for authentication credentials
	AuthKey contextKey = "api_auth"
	// ViewSetKey is the context key for the current viewset
	ViewSetKey contextKey = "api_viewset"
	// ActionKey is the context key for the current action
	ActionKey contextKey = "api_action"
)

// WithUser adds a user to the context
func WithUser(ctx context.Context, user interface{}) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

// UserFromContext retrieves the user from context
func UserFromContext(ctx context.Context) (interface{}, bool) {
	user := ctx.Value(UserKey)
	return user, user != nil
}

// WithAuth adds authentication credentials to the context
func WithAuth(ctx context.Context, auth interface{}) context.Context {
	return context.WithValue(ctx, AuthKey, auth)
}

// AuthFromContext retrieves authentication credentials from context
func AuthFromContext(ctx context.Context) (interface{}, bool) {
	auth := ctx.Value(AuthKey)
	return auth, auth != nil
}

// WithViewSet adds a viewset to the context
func WithViewSet(ctx context.Context, viewset interface{}) context.Context {
	return context.WithValue(ctx, ViewSetKey, viewset)
}

// ViewSetFromContext retrieves the viewset from context
func ViewSetFromContext(ctx context.Context) (interface{}, bool) {
	viewset := ctx.Value(ViewSetKey)
	return viewset, viewset != nil
}

// WithAction adds an action name to the context
func WithAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, ActionKey, action)
}

// ActionFromContext retrieves the action name from context
func ActionFromContext(ctx context.Context) (string, bool) {
	action := ctx.Value(ActionKey)
	if action == nil {
		return "", false
	}
	if s, ok := action.(string); ok {
		return s, true
	}
	return "", false
}
