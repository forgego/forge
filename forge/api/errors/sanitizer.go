package errors

import (
	"fmt"
	"regexp"
	"strings"
)

// Sanitizer sanitizes errors to prevent sensitive data leakage
type Sanitizer struct {
	hideStackTraces  bool
	hideDatabaseErrs bool
	redactPII        bool
	piiPatterns      []*regexp.Regexp
}

// SanitizerConfig configures the sanitizer
type SanitizerConfig struct {
	HideStackTraces  bool
	HideDatabaseErrs bool
	RedactPII        bool
	PIIPatterns      []string
}

// NewSanitizer creates a new sanitizer with the given configuration
func NewSanitizer(config SanitizerConfig) *Sanitizer {
	s := &Sanitizer{
		hideStackTraces:  config.HideStackTraces,
		hideDatabaseErrs: config.HideDatabaseErrs,
		redactPII:        config.RedactPII,
		piiPatterns:      make([]*regexp.Regexp, 0),
	}

	// Default PII patterns
	defaultPatterns := []string{
		`(?i)password`,
		`(?i)token`,
		`(?i)secret`,
		`(?i)api[_-]?key`,
		`(?i)auth[_-]?token`,
		`(?i)access[_-]?token`,
		`(?i)refresh[_-]?token`,
		`(?i)session[_-]?id`,
		`(?i)credit[_-]?card`,
		`(?i)ssn`,
		`(?i)social[_-]?security`,
		`(?i)email`,
		`(?i)phone`,
		`(?i)address`,
	}

	// Combine default and custom patterns
	allPatterns := append(defaultPatterns, config.PIIPatterns...)

	// Compile regex patterns
	for _, pattern := range allPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			s.piiPatterns = append(s.piiPatterns, re)
		}
	}

	return s
}

// DefaultSanitizer returns a sanitizer with secure defaults
// Stack traces and database errors are ALWAYS hidden
func DefaultSanitizer() *Sanitizer {
	return NewSanitizer(SanitizerConfig{
		HideStackTraces:  true, // Always true - never expose stack traces
		HideDatabaseErrs: true, // Always true - never expose DB errors
		RedactPII:        true,
		PIIPatterns:      []string{},
	})
}

// SanitizeError sanitizes an error message
func (s *Sanitizer) SanitizeError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()

	// Remove stack traces (always, regardless of config)
	msg = s.removeStackTraces(msg)

	// Sanitize database errors
	if s.hideDatabaseErrs {
		msg = s.sanitizeDatabaseErrors(msg)
	}

	// Redact PII
	if s.redactPII {
		msg = s.redactPIIFromMessage(msg)
	}

	return msg
}

// SanitizeString sanitizes a string message
func (s *Sanitizer) SanitizeString(msg string) string {
	// Remove stack traces
	msg = s.removeStackTraces(msg)

	// Sanitize database errors
	if s.hideDatabaseErrs {
		msg = s.sanitizeDatabaseErrors(msg)
	}

	// Redact PII
	if s.redactPII {
		msg = s.redactPIIFromMessage(msg)
	}

	return msg
}

// removeStackTraces removes stack trace information from error messages
func (s *Sanitizer) removeStackTraces(msg string) string {
	// Remove common stack trace patterns
	patterns := []string{
		`(?s)goroutine \d+.*?\(0x[0-9a-f]+\)`,
		`(?s)runtime\.panic.*?\n`,
		`(?s)panic:.*?\n`,
		`(?s)at .*?\(.*?\)`,
		`(?s)/.*?\.go:\d+`,
		`(?s)\[0x[0-9a-f]+\]`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		msg = re.ReplaceAllString(msg, "")
	}

	// Remove lines that look like file paths with line numbers
	re := regexp.MustCompile(`(?m)^\s*[a-zA-Z]:?[\\/].*?\.go:\d+.*$`)
	msg = re.ReplaceAllString(msg, "")

	return strings.TrimSpace(msg)
}

// sanitizeDatabaseErrors sanitizes database error messages
func (s *Sanitizer) sanitizeDatabaseErrors(msg string) string {
	// Common database error patterns to sanitize
	dbPatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{
			regexp.MustCompile(`(?i)connection.*refused`),
			"Database connection error",
		},
		{
			regexp.MustCompile(`(?i)duplicate key.*`),
			"Duplicate entry",
		},
		{
			regexp.MustCompile(`(?i)foreign key.*`),
			"Referential integrity error",
		},
		{
			regexp.MustCompile(`(?i)violates.*constraint`),
			"Data constraint violation",
		},
		{
			regexp.MustCompile(`(?i)column.*does not exist`),
			"Database schema error",
		},
		{
			regexp.MustCompile(`(?i)table.*does not exist`),
			"Database schema error",
		},
		{
			regexp.MustCompile(`(?i)pq:.*`),
			"Database error occurred",
		},
		{
			regexp.MustCompile(`(?i)sql:.*`),
			"Database error occurred",
		},
		{
			regexp.MustCompile(`(?i)mysql.*error.*`),
			"Database error occurred",
		},
		{
			regexp.MustCompile(`(?i)postgres.*error.*`),
			"Database error occurred",
		},
	}

	for _, dbPattern := range dbPatterns {
		if dbPattern.pattern.MatchString(msg) {
			return dbPattern.replacement
		}
	}

	return msg
}

// redactPIIFromMessage redacts PII from error messages
func (s *Sanitizer) redactPIIFromMessage(msg string) string {
	for _, pattern := range s.piiPatterns {
		// Find matches and redact them
		msg = pattern.ReplaceAllStringFunc(msg, func(match string) string {
			// Redact the value but keep the key
			if strings.Contains(strings.ToLower(match), "=") {
				parts := strings.SplitN(match, "=", 2)
				if len(parts) == 2 {
					return fmt.Sprintf("%s=[REDACTED]", parts[0])
				}
			}
			return "[REDACTED]"
		})
	}

	return msg
}

// SanitizeMap sanitizes a map of values, redacting PII
func (s *Sanitizer) SanitizeMap(data map[string]interface{}) map[string]interface{} {
	if !s.redactPII {
		return data
	}

	sanitized := make(map[string]interface{})
	for k, v := range data {
		keyLower := strings.ToLower(k)

		// Check if key matches PII patterns
		isPII := false
		for _, pattern := range s.piiPatterns {
			if pattern.MatchString(keyLower) {
				isPII = true
				break
			}
		}

		if isPII {
			sanitized[k] = "[REDACTED]"
		} else {
			// Recursively sanitize nested maps
			if nestedMap, ok := v.(map[string]interface{}); ok {
				sanitized[k] = s.SanitizeMap(nestedMap)
			} else {
				sanitized[k] = v
			}
		}
	}

	return sanitized
}
