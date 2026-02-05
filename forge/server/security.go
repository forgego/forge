package server

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
)

// CSRF wraps gorilla/csrf with framework-specific methods
type CSRF struct {
	protect func(http.Handler) http.Handler
	secret  []byte
}

// NewCSRF creates a new CSRF protector
func NewCSRF(secretKey []byte, opts ...csrf.Option) *CSRF {
	protect := csrf.Protect(secretKey, opts...)
	return &CSRF{
		protect: protect,
		secret:  secretKey,
	}
}

// Middleware returns the CSRF protection middleware
func (c *CSRF) Middleware() func(http.Handler) http.Handler {
	return c.protect
}

// Token returns the CSRF token for the current request
func (c *CSRF) Token(r *http.Request) string {
	return csrf.Token(r)
}

// TemplateField returns a hidden input field for forms
func (c *CSRF) TemplateField(r *http.Request) string {
	return string(csrf.TemplateField(r))
}

// DefaultCSRFOptions returns default CSRF options suitable for framework
func DefaultCSRFOptions() []csrf.Option {
	return []csrf.Option{
		csrf.Secure(false), // Set to true in production with HTTPS
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.Path("/"),
	}
}

// ProductionCSRFOptions returns CSRF options for production
func ProductionCSRFOptions() []csrf.Option {
	return []csrf.Option{
		csrf.Secure(true), // Require HTTPS
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.Path("/"),
	}
}

// SessionManager wraps scs.SessionManager with framework-specific methods
type SessionManager struct {
	*scs.SessionManager
}

// NewSessionManager creates a new session manager
func NewSessionManager(secretKey []byte) *SessionManager {
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "forge_session"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = false // Set to true in production
	sessionManager.Cookie.SameSite = http.SameSiteStrictMode
	sessionManager.Cookie.Path = "/"

	return &SessionManager{
		SessionManager: sessionManager,
	}
}

// Middleware returns the session middleware
func (sm *SessionManager) Middleware() func(http.Handler) http.Handler {
	return sm.SessionManager.LoadAndSave
}

// Get retrieves a value from the session
func (sm *SessionManager) Get(r *http.Request, key string) interface{} {
	return sm.SessionManager.Get(r.Context(), key)
}

// Put stores a value in the session
func (sm *SessionManager) Put(r *http.Request, key string, val interface{}) {
	sm.SessionManager.Put(r.Context(), key, val)
}

// Remove removes a value from the session
func (sm *SessionManager) Remove(r *http.Request, key string) {
	sm.SessionManager.Remove(r.Context(), key)
}

// Exists checks if a key exists in the session
func (sm *SessionManager) Exists(r *http.Request, key string) bool {
	return sm.SessionManager.Exists(r.Context(), key)
}

// Destroy destroys the session
func (sm *SessionManager) Destroy(r *http.Request) error {
	return sm.SessionManager.Destroy(r.Context())
}

// RenewToken renews the session token
func (sm *SessionManager) RenewToken(r *http.Request) error {
	return sm.SessionManager.RenewToken(r.Context())
}

// SQLInjection provides SQL injection prevention utilities
type SQLInjection struct{}

// NewSQLInjection creates a new SQL injection protector
func NewSQLInjection() *SQLInjection {
	return &SQLInjection{}
}

// ValidateInput validates input to prevent SQL injection
func (s *SQLInjection) ValidateInput(input string) error {
	// Check for SQL injection patterns
	dangerousPatterns := []string{
		";",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"exec",
		"execute",
		"union",
		"select",
		"insert",
		"update",
		"delete",
		"drop",
		"create",
		"alter",
		"truncate",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(inputLower, pattern) {
			return fmt.Errorf("potentially dangerous input detected: %s", pattern)
		}
	}

	return nil
}

// SanitizeIdentifier sanitizes SQL identifiers (table/column names)
func (s *SQLInjection) SanitizeIdentifier(identifier string) string {
	// Only allow alphanumeric and underscore
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return reg.ReplaceAllString(identifier, "")
}

// EnsureParameterized ensures a query uses parameterized queries
func (s *SQLInjection) EnsureParameterized(query string) error {
	// Check for string concatenation in SQL
	if strings.Contains(query, "+") || strings.Contains(query, "||") {
		return fmt.Errorf("query appears to use string concatenation - use parameterized queries instead")
	}

	// Check for fmt.Sprintf patterns
	if strings.Contains(query, "%s") || strings.Contains(query, "%d") {
		return fmt.Errorf("query appears to use string formatting - use parameterized queries instead")
	}

	return nil
}

// QueryLogger interface for security audit logging
type QueryLogger interface {
	LogQuery(query string, args []interface{}, timestamp time.Time, source string)
}

// DefaultQueryLogger logs queries to the application log
type DefaultQueryLogger struct {
	enabled bool
}

// NewDefaultQueryLogger creates a new default query logger
func NewDefaultQueryLogger(enabled bool) *DefaultQueryLogger {
	return &DefaultQueryLogger{enabled: enabled}
}

// LogQuery logs a query for security auditing
func (l *DefaultQueryLogger) LogQuery(query string, args []interface{}, timestamp time.Time, source string) {
	if !l.enabled {
		return
	}
	// Log with structured format for security monitoring
	// In production, this should go to a secure audit log system
	fmt.Printf("[AUDIT] %s | Source: %s | Query: %s | Args: %v\n",
		timestamp.Format(time.RFC3339),
		source,
		query,
		args,
	)
}

// LogQuery logs a query for security auditing
func (s *SQLInjection) LogQuery(query string, args []interface{}) {
	// Use default logger for now
	logger := NewDefaultQueryLogger(true)
	logger.LogQuery(query, args, time.Now(), "sql_injection_check")
}

// XSS provides XSS protection utilities
type XSS struct{}

// NewXSS creates a new XSS protector
func NewXSS() *XSS {
	return &XSS{}
}

// EscapeHTML escapes HTML special characters
func (x *XSS) EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// HTMLPolicy represents allowed HTML elements and attributes
type HTMLPolicy struct {
	AllowedTags       map[string]bool
	AllowedAttributes map[string][]string
	StripComments     bool
	StripScripts      bool
}

// DefaultHTMLPolicy returns a safe default HTML policy
func DefaultHTMLPolicy() *HTMLPolicy {
	return &HTMLPolicy{
		AllowedTags: map[string]bool{
			"p": true, "br": true, "strong": true, "em": true, "u": true,
			"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
			"ul": true, "ol": true, "li": true, "a": true, "img": true,
			"blockquote": true, "code": true, "pre": true, "span": true, "div": true,
		},
		AllowedAttributes: map[string][]string{
			"a":   {"href", "title", "target"},
			"img": {"src", "alt", "title", "width", "height"},
		},
		StripComments: true,
		StripScripts:  true,
	}
}

// SanitizeHTML sanitizes HTML content according to policy
func (x *XSS) SanitizeHTML(htmlContent string) string {
	policy := DefaultHTMLPolicy()
	
	// Strip comments
	if policy.StripComments {
		commentRe := regexp.MustCompile(`<!--.*?-->`)
		htmlContent = commentRe.ReplaceAllString(htmlContent, "")
	}
	
	// Strip script tags and their content
	if policy.StripScripts {
		scriptRe := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
		htmlContent = scriptRe.ReplaceAllString(htmlContent, "")
	}
	
	// Strip event handlers (onclick, onerror, etc.)
	eventHandlerRe := regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["'][^"']*["']`)
	htmlContent = eventHandlerRe.ReplaceAllString(htmlContent, "")
	
	// Strip javascript: URLs
	jsUrlRe := regexp.MustCompile(`(?i)javascript:`)
	htmlContent = jsUrlRe.ReplaceAllString(htmlContent, "")
	
	// Strip data: URLs (can be used for XSS)
	dataUrlRe := regexp.MustCompile(`(?i)data:text/html`)
	htmlContent = dataUrlRe.ReplaceAllString(htmlContent, "")
	
	// Strip style tags with expression()
	styleExprRe := regexp.MustCompile(`(?i)<style[^>]*>.*?expression\(.*?\).*?</style>`)
	htmlContent = styleExprRe.ReplaceAllString(htmlContent, "")
	
	// Remove iframe, embed, object tags
	// Note: Go regexp doesn't support backreferences, so we handle each tag separately
	dangerousTags := []string{"iframe", "embed", "object", "applet", "meta", "link", "base"}
	for _, tag := range dangerousTags {
		// Remove paired tags
		pattern := fmt.Sprintf(`(?i)<%s[^>]*>.*?</%s>`, tag, tag)
		tagRe := regexp.MustCompile(pattern)
		htmlContent = tagRe.ReplaceAllString(htmlContent, "")
		
		// Remove self-closing tags
		selfPattern := fmt.Sprintf(`(?i)<%s[^>]*/>`, tag)
		selfRe := regexp.MustCompile(selfPattern)
		htmlContent = selfRe.ReplaceAllString(htmlContent, "")
	}
	
	return htmlContent
}

// SanitizeHTMLStrict sanitizes HTML by escaping everything (safe but loses formatting)
func (x *XSS) SanitizeHTMLStrict(htmlContent string) string {
	return x.EscapeHTML(htmlContent)
}

// SafeString represents a string that is safe to output without escaping
type SafeString string

// String returns the string value
func (s SafeString) String() string {
	return string(s)
}

// MarkSafe marks a string as safe (trusted content)
func MarkSafe(s string) SafeString {
	return SafeString(s)
}

// ContentSecurityPolicy generates CSP headers
func (x *XSS) ContentSecurityPolicy() map[string]string {
	return map[string]string{
		"Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
	}
}

// SanitizeInput sanitizes user input
func (x *XSS) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Escape HTML
	input = x.EscapeHTML(input)

	return input
}
