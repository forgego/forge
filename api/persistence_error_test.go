package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/forgego/forge/api/exceptions"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteException_PostgresRestrictViolationIsBadRequest(t *testing.T) {
	err := fmt.Errorf("delete failed: %w", &pq.Error{
		Code:    "23001",
		Message: `update or delete on table "categories" violates foreign key constraint "categories_parent_id_fkey" on table "categories"`,
		Detail:  `Key (id)=(1) is still referenced from table "categories".`,
	})

	exc := deleteException(err)
	require.Error(t, exc)

	var apiExc *exceptions.APIException
	require.True(t, errors.As(exc, &apiExc))
	assert.Equal(t, http.StatusBadRequest, apiExc.Status)
	assert.Equal(t, "in_use", apiExc.Code)
	assert.Equal(t, "Other records still refer to this record, so it cannot be deleted.", apiExc.Message)
	assert.NotContains(t, apiExc.Message, "categories_parent_id_fkey")
}

func TestDeleteException_PostgresForeignKeyViolationIsBadRequest(t *testing.T) {
	err := fmt.Errorf("delete failed: %w", &pq.Error{
		Code:    "23503",
		Message: `update or delete on table "categories" violates foreign key constraint "categories_parent_id_fkey" on table "categories"`,
		Detail:  `Key (id)=(1) is still referenced from table "categories".`,
	})

	exc := deleteException(err)
	require.Error(t, exc)

	var apiExc *exceptions.APIException
	require.True(t, errors.As(exc, &apiExc))
	assert.Equal(t, http.StatusBadRequest, apiExc.Status)
	assert.Equal(t, "in_use", apiExc.Code)
	assert.Equal(t, "Other records still refer to this record, so it cannot be deleted.", apiExc.Message)
	assert.NotContains(t, apiExc.Message, "categories_parent_id_fkey")
}
