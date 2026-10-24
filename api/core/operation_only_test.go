package core_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckOperationOnlyFields(t *testing.T) {
	data := map[string]interface{}{
		"status": "approved",
		"notes":  "test notes",
	}

	err := core.CheckOperationOnlyFields(data, []string{"status"})
	require.NotNil(t, err)
	assert.Equal(t, "status", err.Field)
	assert.Equal(t, "operation_only_field", err.Code())
	assert.Equal(t, http.StatusForbidden, err.StatusCode())

	errNil := core.CheckOperationOnlyFields(data, []string{"other_field"})
	assert.Nil(t, errNil)
}

func TestEnforceOperationOnlyFields_Allowed(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/1", nil)

	allowed := core.EnforceOperationOnlyFields(rec, req, map[string]interface{}{"notes": "allowed"}, []string{"status"})
	assert.True(t, allowed)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEnforceOperationOnlyFields_Rejected(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/1", nil)

	allowed := core.EnforceOperationOnlyFields(rec, req, map[string]interface{}{"status": "approved"}, []string{"status"})
	assert.False(t, allowed)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "operation_only_field", body["code"])
	errObj, ok := body["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "operation_only_field", errObj["code"])
}
