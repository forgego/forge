package errors

import (
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgego/forge/api/exceptions"
	"github.com/stretchr/testify/assert"
)

func TestWriteErrorWritesProblems(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"validation", exceptions.NewValidationError(nil), http.StatusBadRequest},
		{"authentication", exceptions.NewAuthenticationFailed("invalid"), http.StatusUnauthorized},
		{"permission", exceptions.NewPermissionDenied("forbidden"), http.StatusForbidden},
		{"not found", exceptions.NewNotFound("missing"), http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			WriteError(response, httptest.NewRequest(http.MethodGet, "/items", nil), tt.err)

			assert.Equal(t, tt.status, response.Code)
			assert.Equal(t, "application/problem+json", response.Header().Get("Content-Type"))
		})
	}
}

func TestWriteErrorSetsRetryAfter(t *testing.T) {
	response := httptest.NewRecorder()
	err := exceptions.NewThrottled("slow down", 5*time.Minute)

	WriteError(response, httptest.NewRequest(http.MethodGet, "/items", nil), err)

	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.Equal(t, "300", response.Header().Get("Retry-After"))
}

func TestWriteErrorHidesUnknownErrorMessage(t *testing.T) {
	response := httptest.NewRecorder()

	WriteError(response, httptest.NewRequest(http.MethodGet, "/items", nil), stderrors.New("database password exposed"))

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotContains(t, response.Body.String(), "database password exposed")
}
