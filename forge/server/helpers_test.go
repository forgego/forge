package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHelpers(t *testing.T) {
	t.Run("GetJSON", func(t *testing.T) {
		body := []byte(`{"key": "value"}`)
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))

		var data map[string]string
		err := GetJSON(req, &data)

		assert.NoError(t, err)
		assert.Equal(t, "value", data["key"])
	})

	t.Run("GetJSON Error", func(t *testing.T) {
		body := []byte(`{"key": "value"`) // Invalid JSON
		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(body))

		var data map[string]string
		err := GetJSON(req, &data)

		assert.Error(t, err)
	})

	t.Run("GetQueryInt", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?page=2&invalid=abc", nil)

		assert.Equal(t, 2, GetQueryInt(req, "page", 1))
		assert.Equal(t, 1, GetQueryInt(req, "missing", 1))
		assert.Equal(t, 1, GetQueryInt(req, "invalid", 1))
	})

	t.Run("GetQueryString", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?sort=desc", nil)

		assert.Equal(t, "desc", GetQueryString(req, "sort", "asc"))
		assert.Equal(t, "asc", GetQueryString(req, "missing", "asc"))
	})

	t.Run("GetQueryBool", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/?active=true&inactive=false&invalid=yes", nil)

		assert.True(t, GetQueryBool(req, "active", false))
		assert.False(t, GetQueryBool(req, "inactive", true))
		assert.False(t, GetQueryBool(req, "missing", false))
		assert.False(t, GetQueryBool(req, "invalid", false))
	})

	t.Run("GetParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/123", nil)

		// Set up chi context with url parameters
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		assert.Equal(t, "123", GetParam(req, "id"))
	})

	t.Run("SendJSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"message": "success"}

		err := SendJSON(w, http.StatusCreated, data)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]string
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "success", resp["message"])
	})

	t.Run("SendSuccess", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"user": "test"}

		err := SendSuccess(w, http.StatusOK, data)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.True(t, resp["success"].(bool))
		respData := resp["data"].(map[string]interface{})
		assert.Equal(t, "test", respData["user"])
	})

	t.Run("SendError", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := SendError(w, http.StatusBadRequest, "invalid input")

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/problem+json", w.Header().Get("Content-Type"))

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)

		assert.Contains(t, resp["type"], "bad-request-error")
		assert.Equal(t, float64(http.StatusBadRequest), resp["status"]) // JSON numbers are parsed as float64
		assert.Equal(t, "invalid input", resp["detail"])
	})

	t.Run("SendError - Different Status Codes", func(t *testing.T) {
		codes := []int{
			http.StatusNotFound,
			http.StatusUnauthorized,
			http.StatusForbidden,
			http.StatusMethodNotAllowed,
			http.StatusNotAcceptable,
			http.StatusConflict,
			http.StatusUnsupportedMediaType,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
		}

		for _, code := range codes {
			w := httptest.NewRecorder()
			err := SendError(w, code, "error message")
			assert.NoError(t, err)
			assert.Equal(t, code, w.Code)
		}
	})
}
