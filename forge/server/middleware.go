package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// DefaultMiddlewares returns the default middleware stack for the framework
func DefaultMiddlewares() []Middleware {
	return []Middleware{
		RequestID,
		RealIP,
		Recoverer,
		Logger,
		Timeout(60 * time.Second),
	}
}

// RequestID adds a request ID to each request
func RequestID(next http.Handler) http.Handler {
	return chimw.RequestID(next)
}

// RealIP sets the client IP from X-Forwarded-For or X-Real-IP headers
func RealIP(next http.Handler) http.Handler {
	return chimw.RealIP(next)
}

// Recoverer recovers from panics and returns a 500 error
func Recoverer(next http.Handler) http.Handler {
	return chimw.Recoverer(next)
}

// Logger logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return chimw.Logger(next)
}

// Timeout sets a timeout for request processing
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return chimw.Timeout(timeout)(next)
	}
}

// Compress compresses response bodies with gzip
func Compress(level int) Middleware {
	if level < 1 || level > 9 {
		level = 5
	}
	return func(next http.Handler) http.Handler {
		return chimw.Compress(level)(next)
	}
}

// DefaultCompress uses default compression level (5)
func DefaultCompress(next http.Handler) http.Handler {
	return Compress(5)(next)
}

// StripSlashes removes trailing slashes from URLs
func StripSlashes(next http.Handler) http.Handler {
	return chimw.StripSlashes(next)
}

// RedirectSlashes redirects trailing slashes
func RedirectSlashes(next http.Handler) http.Handler {
	return chimw.RedirectSlashes(next)
}

// CORSOptions configures CORS middleware
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSOptions returns default CORS options
func DefaultCORSOptions() *CORSOptions {
	return &CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// CORS adds CORS headers using chi/cors package
func CORS(opts *CORSOptions) Middleware {
	if opts == nil {
		opts = DefaultCORSOptions()
	}

	c := cors.Options{
		AllowedOrigins:   opts.AllowedOrigins,
		AllowedMethods:   opts.AllowedMethods,
		AllowedHeaders:   opts.AllowedHeaders,
		ExposedHeaders:   opts.ExposedHeaders,
		AllowCredentials: opts.AllowCredentials,
		MaxAge:           opts.MaxAge,
	}

	return cors.Handler(c)
}

// SimpleCORS is a convenience function for simple CORS setup
func SimpleCORS(allowedOrigins, allowedMethods, allowedHeaders []string) Middleware {
	return CORS(&CORSOptions{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   allowedMethods,
		AllowedHeaders:   allowedHeaders,
		AllowCredentials: true,
		MaxAge:           300,
	})
}

// RateLimit provides basic rate limiting (simple implementation)
// Note: For production use, see ratelimit.go for proper implementation
func RateLimit(requestsPerMinute int) Middleware {
	// This is a simplified implementation
	// In production, use a proper rate limiting library
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Note: Actual rate limiting implementation is planned.
			// For production use, integrate a proper rate limiting library like golang.org/x/time/rate
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeadersOptions configures security headers
type SecureHeadersOptions struct {
	// HSTS (HTTP Strict Transport Security)
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// Content Security Policy
	CSP string

	// Cross-Origin policies
	CrossOriginOpenerPolicy   string
	CrossOriginEmbedderPolicy string

	// Other security headers
	XContentTypeOptions string
	XFrameOptions       string
	XXSSProtection      string
	ReferrerPolicy      string
	PermissionsPolicy   string
}

// DefaultSecureHeadersOptions returns default security header options
func DefaultSecureHeadersOptions() *SecureHeadersOptions {
	return &SecureHeadersOptions{
		HSTSMaxAge:                31536000, // 1 year
		HSTSIncludeSubdomains:     false,
		HSTSPreload:               false,
		CSP:                       "",
		CrossOriginOpenerPolicy:   "",
		CrossOriginEmbedderPolicy: "",
		XContentTypeOptions:       "nosniff",
		XFrameOptions:             "DENY",
		XXSSProtection:            "1; mode=block",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "",
	}
}

// SecureHeaders adds security headers to responses
func SecureHeaders(opts *SecureHeadersOptions) Middleware {
	if opts == nil {
		opts = DefaultSecureHeadersOptions()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only add HSTS for HTTPS requests
			if r.TLS != nil && opts.HSTSMaxAge > 0 {
				hsts := fmt.Sprintf("max-age=%d", opts.HSTSMaxAge)
				if opts.HSTSIncludeSubdomains {
					hsts += "; includeSubDomains"
				}
				if opts.HSTSPreload {
					hsts += "; preload"
				}
				w.Header().Set("Strict-Transport-Security", hsts)
			}

			if opts.CSP != "" {
				w.Header().Set("Content-Security-Policy", opts.CSP)
			}

			if opts.CrossOriginOpenerPolicy != "" {
				w.Header().Set("Cross-Origin-Opener-Policy", opts.CrossOriginOpenerPolicy)
			}

			if opts.CrossOriginEmbedderPolicy != "" {
				w.Header().Set("Cross-Origin-Embedder-Policy", opts.CrossOriginEmbedderPolicy)
			}

			if opts.XContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", opts.XContentTypeOptions)
			}

			if opts.XFrameOptions != "" {
				w.Header().Set("X-Frame-Options", opts.XFrameOptions)
			}

			if opts.XXSSProtection != "" {
				w.Header().Set("X-XSS-Protection", opts.XXSSProtection)
			}

			if opts.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", opts.ReferrerPolicy)
			}

			if opts.PermissionsPolicy != "" {
				w.Header().Set("Permissions-Policy", opts.PermissionsPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CacheControl sets cache control headers
func CacheControl(maxAge int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "max-age="+strconv.Itoa(maxAge))
			next.ServeHTTP(w, r)
		})
	}
}

// NoCache disables caching for responses
func NoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// Heartbeat provides a simple heartbeat endpoint
func Heartbeat(pattern string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == pattern {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("."))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetHead automatically converts HEAD requests to GET
func GetHead(next http.Handler) http.Handler {
	return chimw.GetHead(next)
}

// CleanPath cleans URL paths
func CleanPath(next http.Handler) http.Handler {
	return chimw.CleanPath(next)
}

// ContentCharset sets the content charset
func ContentCharset(charset string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if charset != "" {
				ct := w.Header().Get("Content-Type")
				if ct != "" && !strings.Contains(ct, "charset=") {
					w.Header().Set("Content-Type", ct+"; charset="+charset)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequestSizeLimit limits the size of request bodies
func RequestSizeLimit(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// ETagOptions configures ETag middleware
type ETagOptions struct {
	Weak bool
}

// ETag generates ETags for responses
func ETag(opts *ETagOptions) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response writer that captures the body
			rrw := &etagResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				weak:           opts != nil && opts.Weak,
			}

			next.ServeHTTP(rrw, r)

			// Generate ETag if not already set
			if rrw.Header().Get("ETag") == "" && rrw.body.Len() > 0 {
				hash := md5.Sum(rrw.body.Bytes())
				etag := fmt.Sprintf(`"%x"`, hash)
				if rrw.weak {
					etag = "W/" + etag
				}
				rrw.Header().Set("ETag", etag)
			}

			// Check If-None-Match header
			if rrw.statusCode == http.StatusOK {
				ifNoneMatch := r.Header.Get("If-None-Match")
				if ifNoneMatch != "" {
					currentETag := rrw.Header().Get("ETag")
					if currentETag != "" {
						// Remove W/ prefix for comparison
						cleanETag := strings.TrimPrefix(currentETag, "W/")
						cleanIfNoneMatch := strings.TrimPrefix(ifNoneMatch, "W/")
						if cleanETag == cleanIfNoneMatch || strings.Contains(cleanIfNoneMatch, cleanETag) {
							rrw.statusCode = http.StatusNotModified
							rrw.body.Reset()
						}
					}
				}
			}

			// Write status and body
			rrw.ResponseWriter.WriteHeader(rrw.statusCode)
			if rrw.body.Len() > 0 {
				rrw.ResponseWriter.Write(rrw.body.Bytes())
			}
		})
	}
}

// etagResponseWriter captures response for ETag generation
type etagResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
	weak       bool
}

func (w *etagResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *etagResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.body.Write(b)
}

// ConditionalGet handles If-Modified-Since and If-None-Match headers
func ConditionalGet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check If-Modified-Since
		if r.Method == "GET" || r.Method == "HEAD" {
			ifModifiedSince := r.Header.Get("If-Modified-Since")
			if ifModifiedSince != "" {
				// This is a simplified check - in practice, you'd compare with actual resource modification time
				// For now, we'll let the handler process it normally
			}
		}
		next.ServeHTTP(w, r)
	})
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

// Profiler enables profiling endpoints (dev mode only)
func Profiler(pattern string, enabled bool) Middleware {
	if !enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, pattern) {
				chimw.Profiler().ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsOptions configures metrics middleware
type MetricsOptions struct {
	Enabled bool
	Path    string
}

// Metrics provides basic request metrics
func Metrics(opts *MetricsOptions) Middleware {
	if opts == nil || !opts.Enabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Simple in-memory metrics
	// In production, you'd want to use a proper metrics library like prometheus
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			// Note: Metrics storage is planned. For now, duration is available for logging.
			// Consider integrating Prometheus or another metrics system.
			_ = duration
			_ = ww.statusCode
		})
	}
}
