package log

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRedactQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantSub []string
		wontSub []string
		exact   string
	}{
		{
			name:  "empty query",
			input: "",
			exact: "",
		},
		{
			name:  "invalid query unparseable",
			input: "%zz",
			exact: "[unparseable]",
		},
		{
			name:    "api_key redacted but page preserved",
			input:   "api_key=abc&page=2",
			wantSub: []string{"api_key=REDACTED", "page=2"},
			wontSub: []string{"abc"},
		},
		{
			name:    "Token case insensitive redacted",
			input:   "Token=x",
			wantSub: []string{"Token=REDACTED"},
			wontSub: []string{"=x"},
		},
		{
			name:    "mixed case and duplicate sensitive names redacted",
			input:   "ToKeN=first&token=second&API_KEY=third&page=2",
			wantSub: []string{"ToKeN=REDACTED", "token=REDACTED", "API_KEY=REDACTED", "page=2"},
			wontSub: []string{"first", "second", "third"},
		},
		{
			name:  "all sensitive keys redacted",
			input: "api_key=1&apikey=2&key=3&token=4&access_token=5&refresh_token=6&password=7&secret=8&signature=9&session=10&session_key=11&other=visible",
			wantSub: []string{
				"api_key=REDACTED",
				"apikey=REDACTED",
				"key=REDACTED",
				"token=REDACTED",
				"access_token=REDACTED",
				"refresh_token=REDACTED",
				"password=REDACTED",
				"secret=REDACTED",
				"signature=REDACTED",
				"session=REDACTED",
				"session_key=REDACTED",
				"other=visible",
			},
			wontSub: []string{
				"=1&", "=2&", "=3&", "=4&", "=5&", "=6&", "=7&", "=8&", "=9&", "=10&", "=11&",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redactQuery(tt.input)
			if tt.exact != "" || tt.input == "" {
				assert.Equal(t, tt.exact, got)
			}
			for _, want := range tt.wantSub {
				assert.True(t, strings.Contains(got, want), "expected %q to contain %q", got, want)
			}
			for _, wont := range tt.wontSub {
				assert.False(t, strings.Contains(got, wont), "expected %q NOT to contain %q", got, wont)
			}
		})
	}
}

func TestMiddleware_RedactsQueryInLog(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)
	logger := &Logger{Logger: zap.New(core)}

	mw := Middleware(logger)
	req := httptest.NewRequest(http.MethodGet, "/api/resource?api_key=superSecretKey123&page=3", nil)
	rec := httptest.NewRecorder()

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	logOutput := buf.String()
	assert.Contains(t, logOutput, "api_key=REDACTED")
	assert.Contains(t, logOutput, "page=3")
	assert.NotContains(t, logOutput, "superSecretKey123")
}

func TestMiddleware_DoesNotLogSensitiveHeaders(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := &Logger{Logger: zap.New(core)}
	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Authorization", "Bearer authorization-secret")
	req.Header.Add("Cookie", "session=cookie-secret")
	req.Header.Set("Set-Cookie", "session=set-cookie-secret")
	Middleware(logger)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(httptest.NewRecorder(), req)
	for _, secret := range []string{"authorization-secret", "cookie-secret", "set-cookie-secret"} {
		assert.NotContains(t, buf.String(), secret)
	}
}

func TestMiddleware_RecordsStatusAndByteCount(t *testing.T) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(buf), zapcore.DebugLevel)
	logger := &Logger{Logger: zap.New(core)}
	Middleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Contains(t, buf.String(), `"status":201`)
	assert.Contains(t, buf.String(), `"bytes":5`)
}

type flusherHijackerRecorder struct {
	*httptest.ResponseRecorder
}

type capabilityWriter struct {
	*httptest.ResponseRecorder
	flushed  bool
	hijacked bool
	pushed   string
}

func (w *capabilityWriter) Flush() { w.flushed = true }
func (w *capabilityWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, nil
}
func (w *capabilityWriter) Push(target string, _ *http.PushOptions) error {
	w.pushed = target
	return nil
}

func (f *flusherHijackerRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijacker not supported: %w", http.ErrNotSupported)
}

func TestMiddleware_ResponseWriterFlusherAndHijacker(t *testing.T) {
	logger := NewNopLogger()
	mw := Middleware(logger)

	var typeAssertFlushed bool
	var responseControllerFlushed bool
	var directHijackErr error
	var rcHijackErr error

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok, "responseWriter must implement http.Flusher")
		flusher.Flush()
		typeAssertFlushed = true

		rc := http.NewResponseController(w)
		err := rc.Flush()
		require.NoError(t, err, "ResponseController.Flush must succeed")
		responseControllerFlushed = true

		hijacker, ok := w.(http.Hijacker)
		require.True(t, ok, "responseWriter must implement http.Hijacker")
		_, _, directHijackErr = hijacker.Hijack()

		_, _, rcHijackErr = rc.Hijack()
	}))

	rec := &flusherHijackerRecorder{ResponseRecorder: httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	handler.ServeHTTP(rec, req)

	assert.True(t, typeAssertFlushed)
	assert.True(t, responseControllerFlushed)
	assert.True(t, rec.Flushed)
	assert.Error(t, directHijackErr)
	assert.Error(t, rcHijackErr)
}

func TestMiddleware_ForwardsWriterCapabilities(t *testing.T) {
	underlying := &capabilityWriter{ResponseRecorder: httptest.NewRecorder()}
	Middleware(NewNopLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.(http.Flusher).Flush()
		_, _, err := w.(http.Hijacker).Hijack()
		require.NoError(t, err)
		require.NoError(t, w.(http.Pusher).Push("/asset", nil))
		require.Same(t, underlying, w.(interface{ Unwrap() http.ResponseWriter }).Unwrap())
	})).ServeHTTP(underlying, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.True(t, underlying.flushed)
	assert.True(t, underlying.hijacked)
	assert.Equal(t, "/asset", underlying.pushed)
}
