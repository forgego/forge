package core

import (
	"context"
	"errors"
	"testing"

	apicore "github.com/forgego/forge/api/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteAction_ErrActionNotFound(t *testing.T) {
	cfg := &Config[writeFieldsTestItem]{
		Actions: []Action[writeFieldsTestItem]{
			NewAction("existing", "Existing Action", func(context.Context, []*writeFieldsTestItem) error {
				return nil
			}),
		},
	}
	admin, _, id, _ := setupWriteFieldsDB(t, cfg)
	ctx := apicore.WithUser(context.Background(), map[string]string{"role": "admin"})

	res, err := admin.ExecuteAction(ctx, "non_existent_action", []any{id}, nil)
	assert.Nil(t, res)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrActionNotFound), "expected ErrActionNotFound, got %v", err)
	assert.Contains(t, err.Error(), "action not found: non_existent_action")
}

func TestExecuteAction_FillsSucceeded(t *testing.T) {
	var executed []*writeFieldsTestItem
	cfg := &Config[writeFieldsTestItem]{
		Actions: []Action[writeFieldsTestItem]{
			NewAction("publish", "Publish", func(ctx context.Context, items []*writeFieldsTestItem) error {
				executed = items
				return nil
			}),
		},
	}
	admin, manager, id1, _ := setupWriteFieldsDB(t, cfg)
	ctx := apicore.WithUser(context.Background(), map[string]string{"role": "admin"})

	// Create a second item
	item2 := &writeFieldsTestItem{Name: "Item 2"}
	err := manager.Create(ctx, item2)
	require.NoError(t, err)
	id2 := item2.ID

	t.Run("all succeeded", func(t *testing.T) {
		res, err := admin.ExecuteAction(ctx, "publish", []any{id1, id2}, nil)
		require.NoError(t, err)
		resp, ok := res.(*BulkActionResponse)
		require.True(t, ok)
		assert.True(t, resp.Success)
		assert.Equal(t, 2, resp.Affected)
		assert.Equal(t, []int64{id1, id2}, resp.Succeeded)
		assert.Empty(t, resp.Errors)
		assert.Len(t, executed, 2)
	})

	t.Run("mixed result with invalid and not_found ids", func(t *testing.T) {
		res, err := admin.ExecuteAction(ctx, "publish", []any{id1, int64(999999), "invalid"}, nil)
		require.NoError(t, err)
		resp, ok := res.(*BulkActionResponse)
		require.True(t, ok)
		assert.False(t, resp.Success)
		assert.Equal(t, 1, resp.Affected)
		assert.Equal(t, []int64{id1}, resp.Succeeded)
		require.Len(t, resp.Errors, 2)
		assert.Equal(t, "not_found", resp.Errors[0].Code)
		assert.Equal(t, "invalid_id", resp.Errors[1].Code)
	})

	t.Run("handler error does not fill succeeded", func(t *testing.T) {
		handlerErr := errors.New("handler failed")
		failCfg := &Config[writeFieldsTestItem]{
			Actions: []Action[writeFieldsTestItem]{
				NewAction("fail_action", "Fail Action", func(ctx context.Context, items []*writeFieldsTestItem) error {
					return handlerErr
				}),
			},
		}
		failAdmin, _, failID, _ := setupWriteFieldsDB(t, failCfg)
		res, err := failAdmin.ExecuteAction(ctx, "fail_action", []any{failID}, nil)
		assert.Nil(t, res)
		require.Error(t, err)
		assert.True(t, errors.Is(err, handlerErr))
	})
}
