package log

import (
	"net/http"
	"strings"
	"time"

	"github.com/forgego/forge/api/errors"
	"go.uber.org/zap"
)

func sanitizeLogString(s string) string {
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	return s
}

// Middleware creates a logging middleware that logs HTTP requests with request ID
func Middleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			// Get request ID from context
			requestID := errors.GetRequestIDFromContext(r.Context())

			fields := []zap.Field{
				zap.String("method", sanitizeLogString(r.Method)),
				zap.String("path", sanitizeLogString(r.URL.Path)),
				zap.String("query", sanitizeLogString(r.URL.RawQuery)),
				zap.Int("status", ww.statusCode),
				zap.Duration("duration", duration),
				zap.String("ip", sanitizeLogString(r.RemoteAddr)),
				zap.String("user_agent", sanitizeLogString(r.UserAgent())),
			}

			// Add request ID if available
			if requestID != "" {
				fields = append(fields, zap.String("request_id", sanitizeLogString(requestID)))
			}

			// Log at appropriate level
			if ww.statusCode >= 500 {
				logger.Error("HTTP request", fields...)
			} else if ww.statusCode >= 400 {
				logger.Warn("HTTP request", fields...)
			} else {
				logger.Info("HTTP request", fields...)
			}
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
