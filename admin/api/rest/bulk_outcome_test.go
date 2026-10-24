package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAction_RejectsTooManyIDs(t *testing.T) {
	ids := make([]interface{}, core.MaxBulkActionIDs+1)
	for i := range ids {
		ids[i] = i + 1
	}

	body, err := json.Marshal(map[string]interface{}{"ids": ids})
	require.NoError(t, err)

	called := false
	admin := &mockAdmin{
		executeActionFn: func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
			called = true
			return nil, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/archive", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "archive"))

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, called, "ExecuteAction should not be called when IDs exceed limit")

	var errResp core.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
	assert.Equal(t, "too_many_ids", errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, fmt.Sprintf("%d", core.MaxBulkActionIDs))
}

func TestHandleAction_UnknownActionReturnsNotFound(t *testing.T) {
	admin := &mockAdmin{
		executeActionFn: func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("%w: %s", core.ErrActionNotFound, actionName)
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/unknown_act", bytes.NewBufferString(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "unknown_act"))

	require.Equal(t, http.StatusNotFound, rec.Code)

	var errResp core.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errResp))
	assert.Equal(t, "action_not_found", errResp.Error.Code)
	assert.Contains(t, errResp.Error.Message, "unknown_act")
}

func TestHandleAction_MixedResultReturnsMultiStatus(t *testing.T) {
	admin := &mockAdmin{
		executeActionFn: func(actionName string, ids []interface{}, params map[string]interface{}) (interface{}, error) {
			return &core.BulkActionResponse{
				Success:  false,
				Affected: 1,
				Message:  "Executed archive on 1 objects; 1 skipped",
				Errors: []core.BulkActionError{
					{ID: 2, Code: "permission_denied", Message: "permission denied"},
				},
				Succeeded: []int64{1},
			}, nil
		},
	}
	router := NewRouter(core.NewRegistry())

	req := httptest.NewRequest(http.MethodPost, "/api/products/action/archive", bytes.NewBufferString(`{"ids":[1,2]}`))
	rec := httptest.NewRecorder()

	router.handleAction(admin)(rec, withURLParam(req, "action", "archive"))

	require.Equal(t, http.StatusMultiStatus, rec.Code)

	var resp core.BulkActionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.False(t, resp.Success)
	assert.Equal(t, 1, resp.Affected)
	assert.Equal(t, []int64{1}, resp.Succeeded)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, int64(2), resp.Errors[0].ID)
	assert.Equal(t, "permission_denied", resp.Errors[0].Code)
}
