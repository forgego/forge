package deleteerrors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestProtection_PostgresRestrictViolationOnDelete(t *testing.T) {
	err := fmt.Errorf("delete failed: %w", &pq.Error{
		Code:   "23001",
		Detail: `Key (id)=(1) is still referenced from table "categories".`,
	})

	status, code, message, ok := Protection(err)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "in_use", code)
	assert.Equal(t, "Other records still refer to this record, so it cannot be deleted.", message)
}

func TestProtection_PostgresForeignKeyViolationOnDelete(t *testing.T) {
	err := fmt.Errorf("delete failed: %w", &pq.Error{
		Code:   "23503",
		Detail: `Key (id)=(1) is still referenced from table "categories".`,
	})

	status, code, message, ok := Protection(err)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "in_use", code)
	assert.Equal(t, "Other records still refer to this record, so it cannot be deleted.", message)
}
