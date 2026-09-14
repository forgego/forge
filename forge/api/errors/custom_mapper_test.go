package errors

import (
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSentinelConflict = stderrors.New("item already exists in inventory")

func TestCustomMapper_MapsSentinelTo409Problem(t *testing.T) {
	cfg := DefaultHandlerConfig()
	cfg.CustomMapper = func(err error, r *http.Request) *Problem {
		if stderrors.Is(err, errSentinelConflict) {
			return &Problem{
				Type:     "https://api.example.com/problems/conflict",
				Status:   http.StatusConflict,
				Title:    "Conflict",
				Detail:   err.Error(),
				Instance: r.URL.Path,
			}
		}
		return nil
	}

	writer := NewWriter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/items", nil)
	writer(rec, req, errSentinelConflict)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Body.String(), `"status":409`)
	assert.Contains(t, rec.Body.String(), `"title":"Conflict"`)
	assert.Contains(t, rec.Body.String(), "item already exists in inventory")
}

func TestCustomMapper_NilFallsBackToDefault500(t *testing.T) {
	cfg := DefaultHandlerConfig()
	cfg.CustomMapper = func(err error, r *http.Request) *Problem {
		if stderrors.Is(err, errSentinelConflict) {
			return &Problem{
				Type:   "https://api.example.com/problems/conflict",
				Status: http.StatusConflict,
				Title:  "Conflict",
			}
		}
		return nil
	}

	writer := NewWriter(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	otherErr := stderrors.New("unexpected database error")
	writer(rec, req, otherErr)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
	require.NotContains(t, rec.Body.String(), "unexpected database error")
}
