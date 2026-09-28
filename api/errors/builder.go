package errors

import (
	"go.uber.org/zap"
)

// ErrorHandlerBuilder provides a fluent interface for building error handlers
type ErrorHandlerBuilder struct {
	config *HandlerConfig
}

// NewErrorHandlerBuilder creates a new error handler builder
func NewErrorHandlerBuilder() *ErrorHandlerBuilder {
	return &ErrorHandlerBuilder{
		config: DefaultHandlerConfig(),
	}
}

// WithLogger sets the logger
func (b *ErrorHandlerBuilder) WithLogger(logger *zap.Logger) *ErrorHandlerBuilder {
	b.config.Logger = logger
	return b
}

// WithTypeBaseURL sets the base URL for problem type URIs
func (b *ErrorHandlerBuilder) WithTypeBaseURL(url string) *ErrorHandlerBuilder {
	b.config.TypeBaseURL = url
	return b
}

// WithLinkHeader sets the Link header URL
func (b *ErrorHandlerBuilder) WithLinkHeader(url string) *ErrorHandlerBuilder {
	b.config.IncludeLinkHeader = true
	b.config.LinkHeaderURL = url
	return b
}

// WithSanitizer sets a custom sanitizer
func (b *ErrorHandlerBuilder) WithSanitizer(sanitizer *Sanitizer) *ErrorHandlerBuilder {
	b.config.Sanitizer = sanitizer
	return b
}

// WithPanicHandling enables/disables panic handling
func (b *ErrorHandlerBuilder) WithPanicHandling(enabled bool) *ErrorHandlerBuilder {
	b.config.HandlePanics = enabled
	return b
}

// Build creates the error handler
func (b *ErrorHandlerBuilder) Build() *Handler {
	return NewHandler(b.config)
}

// MustBuild creates the error handler and panics on error
func (b *ErrorHandlerBuilder) MustBuild() *Handler {
	return b.Build()
}

// QuickErrorHandler creates a quick error handler for development
func QuickErrorHandler(logger *zap.Logger) *Handler {
	builder := NewErrorHandlerBuilder()
	if logger != nil {
		builder.WithLogger(logger)
	}
	builder.WithTypeBaseURL("https://api.example.com/problems")
	return builder.Build()
}

// ProductionErrorHandler creates a production-ready error handler
func ProductionErrorHandler(logger *zap.Logger, typeBaseURL, linkURL string) *Handler {
	builder := NewErrorHandlerBuilder()
	if logger != nil {
		builder.WithLogger(logger)
	}
	builder.WithTypeBaseURL(typeBaseURL)
	if linkURL != "" {
		builder.WithLinkHeader(linkURL)
	}
	// Use default sanitizer (always enabled in production)
	return builder.Build()
}
