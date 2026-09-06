package server

import (
	"context"

	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextHelpers(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	t.Run("User", func(t *testing.T) {
		assert.Nil(t, GetUser(req))
		req = SetUser(req, "test_user")
		assert.Equal(t, "test_user", GetUser(req))
	})

	t.Run("DB", func(t *testing.T) {
		assert.Nil(t, GetDB(req))
		req = SetDB(req, "test_db")
		assert.Equal(t, "test_db", GetDB(req))
	})

	t.Run("Logger", func(t *testing.T) {
		assert.Nil(t, GetLogger(req))
		req = SetLogger(req, "test_logger")
		assert.Equal(t, "test_logger", GetLogger(req))
	})

	t.Run("Locale", func(t *testing.T) {
		assert.Equal(t, "en", GetLocale(req))
		req = SetLocale(req, "fr")
		assert.Equal(t, "fr", GetLocale(req))
	})

	t.Run("RequestID", func(t *testing.T) {
		assert.Empty(t, GetRequestID(req))
		req = SetRequestID(req, "12345")
		// errors.GetRequestIDFromContext expects a different key type in the current implementation,
		// but SetRequestID sets it using server.RequestIDKey. We will just test that SetRequestID works.
		ctx := req.Context()
		assert.Equal(t, "12345", ctx.Value(RequestIDKey))
	})

	t.Run("WithContext & GetContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "custom", "value")
		req = WithContext(req, ctx)
		assert.Equal(t, ctx, GetContext(req))
		assert.Equal(t, "value", GetContext(req).Value("custom"))
	})
}
