package log

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogRedaction(t *testing.T) {
	tests := []struct {
		name          string
		logFunc       func(logger *Logger)
		secretSubstr  []string // strings that MUST NOT appear anywhere in output or observer
		wantSubstr    []string // strings that MUST appear in output or observer
		checkObserver func(t *testing.T, entry observer.LoggedEntry)
	}{
		// Case 1: Sensitive field names (any case) -> value masked
		{
			name: "sensitive field password lower case",
			logFunc: func(l *Logger) {
				l.Info("login attempt", zap.String("password", "hunter2"))
			},
			secretSubstr: []string{"hunter2"},
			wantSubstr:   []string{"password", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["password"]
				require.True(t, ok, "password field should exist")
				assert.NotContains(t, val, "hunter2")
			},
		},
		{
			name: "sensitive field PASSWORD upper case",
			logFunc: func(l *Logger) {
				l.Info("login attempt", zap.String("PASSWORD", "hunter2_caps"))
			},
			secretSubstr: []string{"hunter2_caps"},
			wantSubstr:   []string{"PASSWORD", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["PASSWORD"]
				require.True(t, ok, "PASSWORD field should exist")
				assert.NotContains(t, val, "hunter2_caps")
			},
		},
		{
			name: "sensitive field passwd any case",
			logFunc: func(l *Logger) {
				l.Info("auth check", zap.String("Passwd", "my_passwd_val"))
			},
			secretSubstr: []string{"my_passwd_val"},
			wantSubstr:   []string{"Passwd", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["Passwd"]
				require.True(t, ok, "Passwd field should exist")
				assert.NotContains(t, val, "my_passwd_val")
			},
		},
		{
			name: "sensitive field secret any case",
			logFunc: func(l *Logger) {
				l.Info("vault fetch", zap.String("SECRET", "top_secret_token_val"))
			},
			secretSubstr: []string{"top_secret_token_val"},
			wantSubstr:   []string{"SECRET", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["SECRET"]
				require.True(t, ok, "SECRET field should exist")
				assert.NotContains(t, val, "top_secret_token_val")
			},
		},
		{
			name: "sensitive field token any case",
			logFunc: func(l *Logger) {
				l.Info("auth token", zap.String("Token", "token_value_abc_123"))
			},
			secretSubstr: []string{"token_value_abc_123"},
			wantSubstr:   []string{"Token", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["Token"]
				require.True(t, ok, "Token field should exist")
				assert.NotContains(t, val, "token_value_abc_123")
			},
		},
		{
			name: "sensitive field access_token any case",
			logFunc: func(l *Logger) {
				l.Info("oauth", zap.String("access_token", "access_tok_xyz"))
			},
			secretSubstr: []string{"access_tok_xyz"},
			wantSubstr:   []string{"access_token", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["access_token"]
				require.True(t, ok, "access_token field should exist")
				assert.NotContains(t, val, "access_tok_xyz")
			},
		},
		{
			name: "sensitive field refresh_token any case",
			logFunc: func(l *Logger) {
				l.Info("oauth refresh", zap.String("REFRESH_TOKEN", "refresh_tok_999"))
			},
			secretSubstr: []string{"refresh_tok_999"},
			wantSubstr:   []string{"REFRESH_TOKEN", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["REFRESH_TOKEN"]
				require.True(t, ok, "REFRESH_TOKEN field should exist")
				assert.NotContains(t, val, "refresh_tok_999")
			},
		},
		{
			name: "sensitive field api_key any case",
			logFunc: func(l *Logger) {
				l.Info("external call", zap.String("API_KEY", "ak_live_1234567890"))
			},
			secretSubstr: []string{"ak_live_1234567890"},
			wantSubstr:   []string{"API_KEY", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["API_KEY"]
				require.True(t, ok, "API_KEY field should exist")
				assert.NotContains(t, val, "ak_live_1234567890")
			},
		},
		{
			name: "sensitive field authorization any case",
			logFunc: func(l *Logger) {
				l.Info("headers", zap.String("Authorization", "Bearer sensitive_token_data"))
			},
			secretSubstr: []string{"sensitive_token_data"},
			wantSubstr:   []string{"Authorization", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["Authorization"]
				require.True(t, ok, "Authorization field should exist")
				assert.NotContains(t, val, "sensitive_token_data")
			},
		},

		// Case 2: DSN like postgres://user:hunter2@host/db -> password part masked, host kept
		{
			name: "DSN as field value",
			logFunc: func(l *Logger) {
				l.Info("db connect", zap.String("dsn", "postgres://user:hunter2@host/db"))
			},
			secretSubstr: []string{"hunter2"},
			wantSubstr:   []string{"postgres://user:", "host/db"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["dsn"].(string)
				require.True(t, ok)
				assert.NotContains(t, val, "hunter2")
				assert.Contains(t, val, "host/db")
			},
		},
		{
			name: "DSN in error message",
			logFunc: func(l *Logger) {
				l.Error("query failed", zap.Error(errors.New("dial postgres://user:hunter2@host/db: connection refused")))
			},
			secretSubstr: []string{"hunter2"},
			wantSubstr:   []string{"dial postgres://user:", "host/db", "connection refused"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val := entry.ContextMap()["error"]
				strVal := ""
				if err, ok := val.(error); ok {
					strVal = err.Error()
				} else if s, ok := val.(string); ok {
					strVal = s
				}
				assert.NotContains(t, strVal, "hunter2")
				assert.Contains(t, strVal, "host/db")
			},
		},
		{
			name: "DSN in log message",
			logFunc: func(l *Logger) {
				l.Info("connecting to postgres://user:hunter2@host/db now")
			},
			secretSubstr: []string{"hunter2"},
			wantSubstr:   []string{"connecting to postgres://user:", "host/db", "now"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				assert.NotContains(t, entry.Message, "hunter2")
				assert.Contains(t, entry.Message, "host/db")
			},
		},

		// Case 3: URL query ?token=abc -> masked
		{
			name: "URL query ?token=abc as field value",
			logFunc: func(l *Logger) {
				l.Info("request", zap.String("query", "?token=abc"))
			},
			secretSubstr: []string{"abc"},
			wantSubstr:   []string{"token", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["query"].(string)
				require.True(t, ok)
				assert.NotContains(t, val, "abc")
				assert.Contains(t, val, "REDACTED")
			},
		},
		{
			name: "URL with query in full URL field value",
			logFunc: func(l *Logger) {
				l.Info("outbound call", zap.String("url", "https://api.example.com/v1?token=abc&page=2"))
			},
			secretSubstr: []string{"token=abc"},
			wantSubstr:   []string{"https://api.example.com/v1", "page=2", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				val, ok := entry.ContextMap()["url"].(string)
				require.True(t, ok)
				assert.NotContains(t, val, "token=abc")
				assert.Contains(t, val, "page=2")
				assert.Contains(t, val, "REDACTED")
			},
		},
		{
			name: "URL query ?token=abc in log message",
			logFunc: func(l *Logger) {
				l.Info("fetching /download?token=abc completed")
			},
			secretSubstr: []string{"token=abc"},
			wantSubstr:   []string{"/download", "REDACTED"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				assert.NotContains(t, entry.Message, "token=abc")
				assert.Contains(t, entry.Message, "REDACTED")
			},
		},

		// Case 4: Non-sensitive fields (user, path, status) unchanged
		{
			name: "non-sensitive fields user path status unchanged",
			logFunc: func(l *Logger) {
				l.Info("http request",
					zap.String("user", "alice"),
					zap.String("path", "/api/v1/users"),
					zap.Int("status", 200),
				)
			},
			secretSubstr: nil,
			wantSubstr:   []string{"user", "alice", "path", "/api/v1/users", "status", "200"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				m := entry.ContextMap()
				assert.Equal(t, "alice", m["user"])
				assert.Equal(t, "/api/v1/users", m["path"])
				assert.Equal(t, int64(200), m["status"])
			},
		},
		{
			name: "bound fields with With() redacted appropriately",
			logFunc: func(l *Logger) {
				child := l.With(
					zap.String("password", "bound_secret"),
					zap.String("user", "bob"),
				)
				child.Info("bound check", zap.Int("status", 201))
			},
			secretSubstr: []string{"bound_secret"},
			wantSubstr:   []string{"password", "REDACTED", "user", "bob", "status", "201"},
			checkObserver: func(t *testing.T, entry observer.LoggedEntry) {
				m := entry.ContextMap()
				assert.Equal(t, "bob", m["user"])
				assert.Equal(t, int64(201), m["status"])
				assert.NotContains(t, m["password"], "bound_secret")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obsCore, logs := observer.New(zapcore.DebugLevel)
			var prodBuf, consoleBuf bytes.Buffer
			prodCore := zapcore.NewCore(
				NewProductionEncoder(ProductionConfig{}),
				zapcore.AddSync(&prodBuf),
				zapcore.DebugLevel,
			)
			consoleCore := zapcore.NewCore(
				NewConsoleEncoder(DevelopmentConfig{OneLine: true}),
				zapcore.AddSync(&consoleBuf),
				zapcore.DebugLevel,
			)
			tee := zapcore.NewTee(obsCore, prodCore, consoleCore)

			logger := NewLoggerWithCore(tee)
			tt.logFunc(logger)

			require.Equal(t, 1, logs.Len(), "expected 1 logged entry")
			entry := logs.All()[0]

			for _, bufOutput := range []string{prodBuf.String(), consoleBuf.String()} {
				for _, secret := range tt.secretSubstr {
					assert.NotContains(t, bufOutput, secret, "buffer output should not contain secret %q", secret)
				}
				for _, want := range tt.wantSubstr {
					assert.True(t, strings.Contains(bufOutput, want),
						"expected output to contain %q, bufOutput: %s", want, bufOutput)
				}
			}

			for _, secret := range tt.secretSubstr {
				assert.NotContains(t, entry.Message, secret, "entry message should not contain secret %q", secret)
			}

			if tt.checkObserver != nil {
				tt.checkObserver(t, entry)
			}
		})
	}
}

func TestLogRedaction_BuilderNormalLogger(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "redaction.log")
	logger, err := NewBuilder().
		Production().
		AddFileOutput(logPath, LevelDebug, 10, 1, 1, false).
		Build()
	require.NoError(t, err)

	logger.Info("connecting",
		zap.String("password", "super_secret_file_pass"),
		zap.String("dsn", "postgres://user:hunter2@host/db"),
		zap.String("query", "?token=abc_file_tok"),
		zap.String("user", "charlie"),
		zap.Int("status", 200),
	)

	err = logger.Close()
	require.NoError(t, err)

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	data := string(content)

	assert.NotContains(t, data, "super_secret_file_pass")
	assert.NotContains(t, data, "hunter2")
	assert.NotContains(t, data, "abc_file_tok")

	assert.Contains(t, data, "postgres://user:")
	assert.Contains(t, data, "host/db")
	assert.Contains(t, data, "token")
	assert.Contains(t, data, "REDACTED")
	assert.Contains(t, data, "charlie")
	assert.Contains(t, data, "200")
}
