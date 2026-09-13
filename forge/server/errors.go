package server

import (
	"github.com/forgego/forge/api/errors"
	"go.uber.org/zap"
)

// ErrorHandlerOptions configures error handling middleware
type ErrorHandlerOptions struct {
	// Logger for error logging
	Logger *zap.Logger
	// HandlerConfig configures the error handler
	HandlerConfig *errors.HandlerConfig
}

// DefaultErrorHandlerOptions returns default error handler options
func DefaultErrorHandlerOptions() *ErrorHandlerOptions {
	config := errors.DefaultHandlerConfig()
	return &ErrorHandlerOptions{
		Logger:        nil,
		HandlerConfig: config,
	}
}

// ErrorHandler creates middleware that handles errors and panics using the new error system
func ErrorHandler(opts *ErrorHandlerOptions) Middleware {
	if opts == nil {
		opts = DefaultErrorHandlerOptions()
	}

	// Create error handler
	if opts.HandlerConfig == nil {
		opts.HandlerConfig = errors.DefaultHandlerConfig()
	}

	// Set logger if provided
	if opts.Logger != nil {
		opts.HandlerConfig.Logger = opts.Logger
	}

	handler := errors.NewHandler(opts.HandlerConfig)

	return handler.Middleware()
}

// RequestIDMiddleware creates middleware for request ID handling
func RequestIDMiddleware() Middleware {
	return errors.DefaultRequestIDMiddleware()
}
