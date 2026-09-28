package errors

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDHeader is the standard header name for request IDs
const RequestIDHeader = "X-Request-ID"

// RequestIDKey is the context key type for request ID
type requestIDKey struct{}

// GetRequestIDFromContext extracts the request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestIDFromRequest extracts the request ID from the request header or context
func GetRequestIDFromRequest(r *http.Request) string {
	// First check context
	if id := GetRequestIDFromContext(r.Context()); id != "" {
		return id
	}
	// Then check header
	return r.Header.Get(RequestIDHeader)
}

// RequestIDMiddleware creates middleware that generates and propagates request IDs
func RequestIDMiddleware(headerName string, generateIfMissing bool) func(http.Handler) http.Handler {
	if headerName == "" {
		headerName = RequestIDHeader
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from header or generate one
			requestID := r.Header.Get(headerName)

			if requestID == "" && generateIfMissing {
				requestID = uuid.New().String()
			}

			// Add to context
			if requestID != "" {
				ctx := WithRequestID(r.Context(), requestID)
				r = r.WithContext(ctx)

				// Add to response header
				w.Header().Set(headerName, requestID)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DefaultRequestIDMiddleware creates the default request ID middleware
func DefaultRequestIDMiddleware() func(http.Handler) http.Handler {
	return RequestIDMiddleware(RequestIDHeader, true)
}
