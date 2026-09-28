package core

import "net/http"

// Middleware is a function that wraps an HTTP handler
type Middleware func(http.Handler) http.Handler

// Chain chains multiple middlewares together
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		// Process in reverse order (last middleware wraps first)
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies middleware to a handler
func Apply(handler http.Handler, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(handler)
}

// ApplyFunc applies middleware to a handler function
func ApplyFunc(handler http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(handler)
}
