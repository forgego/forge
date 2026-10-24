package log

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/forgego/forge/api/errors"
	"go.uber.org/zap"
)

var sensitiveQueryParams = map[string]struct{}{
	"api_key":       {},
	"apikey":        {},
	"key":           {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"signature":     {},
	"session":       {},
	"session_key":   {},
	"authorization": {},
}

func redactQuery(raw string) string {
	if raw == "" {
		return ""
	}
	prefix := ""
	if strings.HasPrefix(raw, "?") {
		prefix = "?"
		raw = raw[1:]
	}
	vals, err := url.ParseQuery(raw)
	if err != nil {
		return "[unparseable]"
	}
	hasSensitive := false
	for k := range vals {
		if _, ok := sensitiveQueryParams[strings.ToLower(k)]; ok {
			hasSensitive = true
			break
		}
	}
	if !hasSensitive {
		return prefix + raw
	}
	for k, vs := range vals {
		if _, ok := sensitiveQueryParams[strings.ToLower(k)]; ok {
			for i := range vs {
				vs[i] = "REDACTED"
			}
			vals[k] = vs
		}
	}
	return prefix + vals.Encode()
}

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
			ww := wrapResponseWriter(w)

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			// Get request ID from context
			requestID := errors.GetRequestIDFromContext(r.Context())

			fields := []zap.Field{
				zap.String("method", sanitizeLogString(r.Method)),
				zap.String("path", sanitizeLogString(r.URL.Path)),
				zap.String("query", sanitizeLogString(redactQuery(r.URL.RawQuery))),
				zap.Int("status", ww.status()),
				zap.Int("bytes", ww.written()),
				zap.Duration("duration", duration),
				zap.String("ip", sanitizeLogString(r.RemoteAddr)),
				zap.String("user_agent", sanitizeLogString(r.UserAgent())),
			}

			// Add request ID if available
			if requestID != "" {
				fields = append(fields, zap.String("request_id", sanitizeLogString(requestID)))
			}

			// Log at appropriate level
			if ww.status() >= 500 {
				logger.Error("HTTP request", fields...)
			} else if ww.status() >= 400 {
				logger.Warn("HTTP request", fields...)
			} else {
				logger.Info("HTTP request", fields...)
			}
		})
	}
}

type loggingResponseWriter interface {
	http.ResponseWriter
	status() int
	written() int
	Unwrap() http.ResponseWriter
}

// responseWriter wraps http.ResponseWriter to capture response metadata and
// centralize optional HTTP writer capability forwarding.
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
	wroteHeader  bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.wroteHeader = true
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(p)
	rw.bytesWritten += n
	return n, err
}

func (rw *responseWriter) status() int {
	if rw.statusCode == 0 {
		return http.StatusOK
	}
	return rw.statusCode
}

func (rw *responseWriter) written() int { return rw.bytesWritten }

func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("hijacker not supported: %w", http.ErrNotSupported)
}

func (rw *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func wrapResponseWriter(w http.ResponseWriter) loggingResponseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}
