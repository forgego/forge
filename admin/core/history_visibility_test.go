package core

import (
	"context"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type privateHistoryModel struct {
	schema.BaseSchema
	Name   string
	Secret string `db:"private_value" json:"credential"`
}

func (privateHistoryModel) Fields() []schema.Field {
	return []schema.Field{schema.StringField("name"), schema.StringField("private_value", schema.WriteOnly())}
}
func TestHistoryNeverStoresWriteOnlyChanges(t *testing.T) {
	history := NewMemoryHistoryManager()
	admin, err := NewAdmin[privateHistoryModel](privateHistoryModel{}, nil, &Config[privateHistoryModel]{HistoryManager: history})
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, admin.LogAction(ctx, "operator", "7", "item", ActionChange, `{"name":"public","credential":"synthetic-private","private_value":"synthetic-private","Secret":"synthetic-private"}`))
	entries, err := history.GetHistory(ctx, admin.ModelName(), "7")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.NotContains(t, entries[0].ChangeStats, "synthetic-private")
	assert.Contains(t, entries[0].ChangeStats, "public")
	// Old stored entries must also be filtered on reads without mutating storage.
	require.NoError(t, history.LogAction(ctx, LogEntry{ModelName: admin.ModelName(), ObjectID: "7", ChangeStats: `{"private_value":"legacy-private","name":"old"}`}))
	entries, err = admin.GetHistory(ctx, "7")
	require.NoError(t, err)
	require.Len(t, entries, 2)
	for _, entry := range entries {
		assert.NotContains(t, entry.ChangeStats, "private")
	}
	raw, err := history.GetHistory(ctx, admin.ModelName(), "7")
	require.NoError(t, err)
	assert.Contains(t, raw[0].ChangeStats, "legacy-private", "read filtering must not rewrite stored audit rows")
}
