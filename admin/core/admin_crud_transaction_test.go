package core

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/forgego/forge/audit"
	"github.com/stretchr/testify/require"
)

func TestAdminCRUDHistoryFailureRollsBackModelWrites(t *testing.T) {
	ctx := context.Background()
	admin, manager, id, _ := setupWriteFieldsDB(t, &Config[writeFieldsTestItem]{HistoryManager: failedHistoryManager{}})

	_, err := admin.CreateObject(ctx, map[string]interface{}{"name": "created"})
	require.Error(t, err)
	items, err := manager.All(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)

	_, err = admin.UpdateObject(ctx, id, map[string]interface{}{"name": "changed"})
	require.Error(t, err)
	stored, err := manager.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "Original Name", stored.Name)

	err = admin.DeleteObject(ctx, id)
	require.Error(t, err)
	_, err = manager.Get(ctx, id)
	require.NoError(t, err)
}

func TestAdminCRUDHistorySuccessCommitsModelWrites(t *testing.T) {
	ctx := context.Background()
	history := NewMemoryHistoryManager()
	admin, manager, id, _ := setupWriteFieldsDB(t, &Config[writeFieldsTestItem]{HistoryManager: history})

	createdValue, err := admin.CreateObject(ctx, map[string]interface{}{"name": "created"})
	require.NoError(t, err)
	created := createdValue.(*writeFieldsTestItem)
	stored, err := manager.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "created", stored.Name)

	_, err = admin.UpdateObject(ctx, id, map[string]interface{}{"name": "changed"})
	require.NoError(t, err)
	stored, err = manager.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, "changed", stored.Name)

	require.NoError(t, admin.DeleteObject(ctx, id))
	_, err = manager.Get(ctx, id)
	require.Error(t, err)
	entries, err := history.GetHistory(ctx, admin.ModelName(), "1")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	entries, err = history.GetHistory(ctx, admin.ModelName(), "2")
	require.NoError(t, err)
	require.Len(t, entries, 1)
}

func TestAdminCRUDFailOpenCommitsAfterHistoryFailure(t *testing.T) {
	ctx := context.Background()
	admin, manager, _, _ := setupWriteFieldsDB(t, &Config[writeFieldsTestItem]{
		HistoryManager: failedHistoryManager{}, AuditPolicy: audit.FailOpen,
	})
	var output bytes.Buffer
	admin.Config().AuditLogger = slog.New(slog.NewTextHandler(&output, nil))
	counter := &audit.Counter{}
	admin.Config().AuditFailures = counter

	value, err := admin.CreateObject(ctx, map[string]interface{}{"name": "committed"})
	require.NoError(t, err)
	created := value.(*writeFieldsTestItem)
	stored, err := manager.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "committed", stored.Name)
	require.EqualValues(t, 1, counter.Count())
	require.Contains(t, output.String(), "audit write failed")
}
