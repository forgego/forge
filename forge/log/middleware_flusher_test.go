package log

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type plainWriter struct {
	header http.Header
}

func (p *plainWriter) Header() http.Header {
	if p.header == nil {
		p.header = make(http.Header)
	}
	return p.header
}

func (p *plainWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (p *plainWriter) WriteHeader(statusCode int) {}

// Finding 6: responseWriter must implement http.Flusher only if underlying writer does.
func TestMiddleware_FlusherConditional(t *testing.T) {
	logger := NewNopLogger()
	mw := Middleware(logger)

	// 1. Wrapping a writer without Flush reports _, ok := w.(http.Flusher) false
	var innerIsFlusher bool
	handlerNoFlush := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, innerIsFlusher = w.(http.Flusher)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handlerNoFlush.ServeHTTP(&plainWriter{}, req)
	assert.False(t, innerIsFlusher, "wrapping a writer without Flush reports _, ok := w.(http.Flusher) false")

	// 2. Wrapping httptest.NewRecorder (has Flush) reports true and Flush reaches it
	var recorderIsFlusher bool
	handlerWithFlush := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, ok := w.(http.Flusher)
		recorderIsFlusher = ok
		if ok {
			f.Flush()
		}
	}))

	rec := httptest.NewRecorder()
	handlerWithFlush.ServeHTTP(rec, req)
	assert.True(t, recorderIsFlusher, "wrapping httptest.NewRecorder reports true")
	assert.True(t, rec.Flushed, "Flush reaches underlying recorder")
}
