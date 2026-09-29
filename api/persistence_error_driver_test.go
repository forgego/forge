package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/forgego/forge/api/exceptions"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pins the REST API's statuses for driver errors, which it classifies with
// internal/dberrors like the admin API but maps to its own responses: data
// exceptions stay a 500 here, while the admin answers them with 400.
func TestPersistenceExceptionDriverErrorStatuses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int // 0: returned unchanged, rendered as a generic 500
		code   string
	}{
		{"unique", &pq.Error{Code: "23505", Detail: "Key (slug)=(secret)"}, http.StatusConflict, "conflict"},
		{"foreign key", &pq.Error{Code: "23503"}, http.StatusBadRequest, "invalid_request"},
		{"not null", &pq.Error{Code: "23502", Column: "name"}, http.StatusBadRequest, "invalid_request"},
		{"check", &pq.Error{Code: "23514"}, http.StatusBadRequest, "invalid_request"},
		{"data exception", &pq.Error{Code: "22001", Message: "value too long secret"}, 0, ""},
		{"undefined column", &pq.Error{Code: "42703"}, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("insert failed: %w", tc.err)
			got := persistenceException(wrapped)
			if tc.status == 0 {
				assert.Same(t, wrapped, got)
				return
			}
			apiErr, ok := got.(*exceptions.APIException)
			require.True(t, ok, "got %T", got)
			assert.Equal(t, tc.status, apiErr.Status)
			assert.Equal(t, tc.code, apiErr.Code)
			assert.NotContains(t, apiErr.Message, "secret")
		})
	}
}
