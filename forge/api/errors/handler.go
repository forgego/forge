package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// HandlerConfig configures the error handler
type HandlerConfig struct {
	// Sanitizer for error sanitization (required)
	Sanitizer *Sanitizer
	// Logger for error logging (optional)
	Logger *zap.Logger
	// TypeBaseURL is the base URL for problem type URIs
	TypeBaseURL string
	// IncludeLinkHeader includes Link header pointing to error documentation
	IncludeLinkHeader bool
	// LinkHeaderURL is the URL for the Link header
	LinkHeaderURL string
	// HandlePanics enables panic recovery
	HandlePanics bool
}

// DefaultHandlerConfig returns a default handler configuration
func DefaultHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		Sanitizer:         DefaultSanitizer(),
		Logger:            nil,
		TypeBaseURL:       "https://api.example.com/problems",
		IncludeLinkHeader: true,
		LinkHeaderURL:     "https://api.example.com/docs/errors",
		HandlePanics:      true,
	}
}

// Handler is the centralized error handler
type Handler struct {
	config *HandlerConfig
	mapper *ErrorMapper
}

// NewHandler creates a new error handler
func NewHandler(config *HandlerConfig) *Handler {
	if config == nil {
		config = DefaultHandlerConfig()
	}
	if config.Sanitizer == nil {
		config.Sanitizer = DefaultSanitizer()
	}

	// Set the global type base URL
	if config.TypeBaseURL != "" {
		SetTypeBaseURL(config.TypeBaseURL)
	}

	return &Handler{
		config: config,
		mapper: NewErrorMapper(config.Sanitizer),
	}
}

// Middleware returns HTTP middleware that handles errors
func (h *Handler) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer that captures errors
			rrw := &errorResponseWriter{
				ResponseWriter: w,
				handler:        h,
				request:        r,
			}

			// Handle panics if enabled
			if h.config.HandlePanics {
				defer func() {
					if rec := recover(); rec != nil {
						rrw.handlePanic(rec)
					}
				}()
			}

			// Call next handler
			next.ServeHTTP(rrw, r)
		})
	}
}

// HandleError handles an error and writes the response
func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Get instance URI from request
	instance := r.URL.Path
	if r.URL.RawQuery != "" {
		instance += "?" + r.URL.RawQuery
	}

	// Map error to Problem Details
	problem := h.mapper.MapError(err, instance)

	// Add request ID if available in context
	if requestID := GetRequestIDFromContext(r.Context()); requestID != "" {
		problem.WithMeta("request_id", requestID)
	}

	// Log the error (with full details, including stack trace if available)
	h.logError(err, problem, r)

	// Write response
	h.writeProblem(w, r, problem)
}

// HandlePanic handles a panic and writes the response
func (h *Handler) HandlePanic(w http.ResponseWriter, r *http.Request, rec interface{}) {
	// Get instance URI from request
	instance := r.URL.Path
	if r.URL.RawQuery != "" {
		instance += "?" + r.URL.RawQuery
	}

	// Map panic to Problem Details
	problem := h.mapper.MapPanic(rec, instance)

	// Add request ID if available
	if requestID := GetRequestIDFromContext(r.Context()); requestID != "" {
		problem.WithMeta("request_id", requestID)
	}

	// Log the panic with stack trace
	h.logPanic(rec, problem, r)

	// Write response
	h.writeProblem(w, r, problem)
}

// writeProblem writes the problem to the HTTP response
func (h *Handler) writeProblem(w http.ResponseWriter, r *http.Request, problem *Problem) {
	// Set headers
	if h.config.IncludeLinkHeader && h.config.LinkHeaderURL != "" {
		w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"help\"", h.config.LinkHeaderURL))
	}

	// Add Retry-After header if it's a rate limit error
	if problem.Status == http.StatusTooManyRequests {
		if retryAfter, ok := problem.Meta["retry_after_seconds"].(int); ok {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
		}
	}

	// Write problem response
	if err := problem.Write(w, r); err != nil {
		// Fallback to plain text if writing fails
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(problem.Status)
		fmt.Fprintf(w, "%s: %s", problem.Title, problem.Detail)
	}
}

// logError logs an error with full context
func (h *Handler) logError(err error, problem *Problem, r *http.Request) {
	if h.config.Logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("error_code", problem.Code),
		zap.String("error_type", string(problem.Type)),
		zap.Int("http_status", problem.Status),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("instance", problem.Instance),
	}

	// Add request ID if available
	if requestID := GetRequestIDFromContext(r.Context()); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	// Add error message (sanitized)
	fields = append(fields, zap.String("error_message", problem.Detail))

	// Log original error if it's not already sanitized
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	// Log at appropriate level
	if problem.Status >= 500 {
		h.config.Logger.Error("Internal server error", fields...)
	} else if problem.Status >= 400 {
		h.config.Logger.Warn("Client error", fields...)
	} else {
		h.config.Logger.Info("Error occurred", fields...)
	}
}

// logPanic logs a panic with stack trace
func (h *Handler) logPanic(rec interface{}, problem *Problem, r *http.Request) {
	if h.config.Logger == nil {
		return
	}

	stack := debug.Stack()

	fields := []zap.Field{
		zap.Any("panic", rec),
		zap.String("stack", string(stack)),
		zap.String("error_code", problem.Code),
		zap.String("error_type", string(problem.Type)),
		zap.Int("http_status", problem.Status),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("instance", problem.Instance),
	}

	// Add request ID if available
	if requestID := GetRequestIDFromContext(r.Context()); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	h.config.Logger.Error("Panic recovered", fields...)
}

// errorResponseWriter wraps http.ResponseWriter to capture errors
type errorResponseWriter struct {
	http.ResponseWriter
	handler    *Handler
	request    *http.Request
	written    bool
	statusCode int
}

func (w *errorResponseWriter) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
		if w.statusCode == 0 {
			w.statusCode = http.StatusOK
		}
		w.ResponseWriter.WriteHeader(w.statusCode)
	}
	return w.ResponseWriter.Write(b)
}

// handlePanic handles panics
func (w *errorResponseWriter) handlePanic(rec interface{}) {
	if w.written {
		return
	}
	w.handler.HandlePanic(w, w.request, rec)
}
