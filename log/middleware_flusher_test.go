package log

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type plainWriter struct{ header http.Header }

func (p *plainWriter) Header() http.Header {
	if p.header == nil {
		p.header = make(http.Header)
	}
	return p.header
}
func (p *plainWriter) Write(b []byte) (int, error) { return len(b), nil }
func (p *plainWriter) WriteHeader(int)             {}

func TestMiddleware_UnsupportedWriterCapabilitiesAreGraceful(t *testing.T) {
	var flushReturned bool
	var hijackErr, pushErr error
	handler := Middleware(NewNopLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.(http.Flusher).Flush()
		flushReturned = true
		_, _, hijackErr = w.(http.Hijacker).Hijack()
		pushErr = w.(http.Pusher).Push("/asset", nil)
	}))
	handler.ServeHTTP(&plainWriter{}, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.True(t, flushReturned)
	assert.ErrorIs(t, hijackErr, http.ErrNotSupported)
	assert.ErrorIs(t, pushErr, http.ErrNotSupported)
}
