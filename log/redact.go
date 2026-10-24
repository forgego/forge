package log

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sensitiveFieldNames = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"api_key":       {},
	"apikey":        {},
	"authorization": {},
}

func isSensitiveFieldName(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	if idx := strings.LastIndex(k, "."); idx >= 0 {
		k = k[idx+1:]
	}
	_, ok := sensitiveFieldNames[k]
	return ok
}

var (
	urlPattern             = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://[^\s"'<>,]+`)
	queryPattern           = regexp.MustCompile(`\?[a-zA-Z0-9_.~%!$&'()*+,;=:@/-]+`)
	keywordPasswordPattern = regexp.MustCompile(`(?i)(password\s*=\s*)(?:'[^']*'|"[^"]*"|[^\s,;]+)`)
)

func splitTrailingPunct(s string) (string, string) {
	cut := len(s)
	for cut > 0 {
		c := s[cut-1]
		if c == ':' || c == '.' || c == ',' || c == ';' || c == ')' || c == ']' || c == '>' || c == '"' || c == '\'' {
			cut--
		} else {
			break
		}
	}
	return s[:cut], s[cut:]
}

func redactString(s string) string {
	if s == "" {
		return ""
	}

	// 1. Redact URLs/DSNs with schemes (e.g. postgres://user:hunter2@host/db)
	s = urlPattern.ReplaceAllStringFunc(s, func(m string) string {
		candidate, punct := splitTrailingPunct(m)
		u, err := url.Parse(candidate)
		if err != nil {
			return m
		}
		modified := false
		if u.User != nil {
			if _, hasPass := u.User.Password(); hasPass {
				modified = true
			}
		}
		if u.RawQuery != "" {
			rq := redactQuery(u.RawQuery)
			if rq != u.RawQuery {
				u.RawQuery = rq
				modified = true
			}
		}
		if modified {
			return u.Redacted() + punct
		}
		return m
	})

	// 2. Redact query strings starting with ? (e.g. ?token=abc)
	s = queryPattern.ReplaceAllStringFunc(s, func(m string) string {
		candidate, punct := splitTrailingPunct(m)
		rq := redactQuery(candidate)
		return rq + punct
	})

	// 3. Redact keyword passwords (e.g. "password=hunter2")
	s = keywordPasswordPattern.ReplaceAllString(s, "${1}[REDACTED]")

	// 4. Redact standalone query parameters (e.g. "token=abc&page=2")
	if !strings.Contains(s, "://") && !strings.Contains(s, "/") && strings.Contains(s, "=") && !strings.Contains(s, " ") {
		s = redactQuery(s)
	}

	return s
}

func redactField(f zapcore.Field) zapcore.Field {
	if isSensitiveFieldName(f.Key) {
		return zap.String(f.Key, "[REDACTED]")
	}

	switch f.Type {
	case zapcore.StringType:
		redacted := redactString(f.String)
		if redacted != f.String {
			return zap.String(f.Key, redacted)
		}
	case zapcore.ErrorType:
		if err, ok := f.Interface.(error); ok && err != nil {
			redacted := redactString(err.Error())
			if redacted != err.Error() {
				return zap.NamedError(f.Key, errors.New(redacted))
			}
		}
	case zapcore.ByteStringType:
		if b, ok := f.Interface.([]byte); ok {
			s := string(b)
			redacted := redactString(s)
			if redacted != s {
				return zap.ByteString(f.Key, []byte(redacted))
			}
		}
	case zapcore.BinaryType:
		if b, ok := f.Interface.([]byte); ok {
			s := string(b)
			redacted := redactString(s)
			if redacted != s {
				return zap.Binary(f.Key, []byte(redacted))
			}
		}
	case zapcore.ReflectType:
		if s, ok := f.Interface.(string); ok {
			redacted := redactString(s)
			if redacted != s {
				return zap.Any(f.Key, redacted)
			}
		} else if err, ok := f.Interface.(error); ok && err != nil {
			redacted := redactString(err.Error())
			if redacted != err.Error() {
				return zap.Any(f.Key, errors.New(redacted))
			}
		} else if str, ok := f.Interface.(fmt.Stringer); ok && str != nil {
			s := str.String()
			redacted := redactString(s)
			if redacted != s {
				return zap.String(f.Key, redacted)
			}
		}
	}

	return f
}

func redactFields(fields []zapcore.Field) []zapcore.Field {
	if len(fields) == 0 {
		return fields
	}
	out := make([]zapcore.Field, len(fields))
	for i, f := range fields {
		out[i] = redactField(f)
	}
	return out
}

// redactionCore wraps a zapcore.Core to prevent sensitive fields or values from reaching outputs.
type redactionCore struct {
	core zapcore.Core
}

// NewRedactionCore returns a zapcore.Core that intercepts and redacts entries and fields.
func NewRedactionCore(core zapcore.Core) zapcore.Core {
	if core == nil {
		return nil
	}
	return &redactionCore{core: core}
}

func (c *redactionCore) Enabled(lvl zapcore.Level) bool {
	return c.core.Enabled(lvl)
}

func (c *redactionCore) With(fields []zapcore.Field) zapcore.Core {
	return &redactionCore{core: c.core.With(redactFields(fields))}
}

func (c *redactionCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return ce
	}
	entry.Message = redactString(entry.Message)
	downstream := c.core.Check(entry, nil)
	if downstream == nil {
		return ce
	}
	return ce.AddCore(entry, c)
}

func (c *redactionCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	entry.Message = redactString(entry.Message)
	return c.core.Write(entry, redactFields(fields))
}

func (c *redactionCore) Sync() error {
	return c.core.Sync()
}
