package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryHistoryManager_LogAndGet(t *testing.T) {
	ctx := context.Background()
	mgr := NewMemoryHistoryManager()

	// 1. Initially empty
	history, err := mgr.GetHistory(ctx, "products", "1")
	require.NoError(t, err)
	assert.Empty(t, history)

	// 2. Log actions
	entry1 := LogEntry{
		ModelName:   "products",
		ObjectID:    "1",
		ObjectRepr:  "Product #1",
		Action:      ActionAdd,
		ChangeStats: `{"name":"Product A"}`,
		UserID:      "user-1",
	}
	require.NoError(t, mgr.LogAction(ctx, entry1))

	entry2 := LogEntry{
		ModelName:   "products",
		ObjectID:    "1",
		ObjectRepr:  "Product #1",
		Action:      ActionChange,
		ChangeStats: `{"price":10.5}`,
		UserID:      "user-2",
	}
	require.NoError(t, mgr.LogAction(ctx, entry2))

	entry3 := LogEntry{
		ModelName:   "orders",
		ObjectID:    "99",
		ObjectRepr:  "Order #99",
		Action:      ActionAdd,
		ChangeStats: `{"total":50}`,
		UserID:      "user-1",
	}
	require.NoError(t, mgr.LogAction(ctx, entry3))

	// 3. Query for product 1 -> should return 2 entries, newest first (entry2, then entry1)
	prodHistory, err := mgr.GetHistory(ctx, "products", "1")
	require.NoError(t, err)
	require.Len(t, prodHistory, 2)
	assert.Equal(t, ActionChange, prodHistory[0].Action)
	assert.Equal(t, "user-2", prodHistory[0].UserID)
	assert.Equal(t, ActionAdd, prodHistory[1].Action)
	assert.Equal(t, "user-1", prodHistory[1].UserID)

	// 4. Query for order 99
	orderHistory, err := mgr.GetHistory(ctx, "orders", "99")
	require.NoError(t, err)
	require.Len(t, orderHistory, 1)
	assert.Equal(t, "Order #99", orderHistory[0].ObjectRepr)

	// 5. Query for non-existent object
	none, err := mgr.GetHistory(ctx, "products", "999")
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestMemoryHistoryManager_MaxSize(t *testing.T) {
	ctx := context.Background()
	mgr := &MemoryHistoryManager{
		entries: make([]LogEntry, 0),
		maxSize: 3,
	}

	for i := 1; i <= 5; i++ {
		require.NoError(t, mgr.LogAction(ctx, LogEntry{
			ModelName:  "items",
			ObjectID:   fmt.Sprintf("%d", i),
			ObjectRepr: fmt.Sprintf("Item %d", i),
			Action:     ActionAdd,
		}))
	}

	// Should only retain the last 3 items (5, 4, 3)
	all, err := mgr.GetHistory(ctx, "items", "")
	require.NoError(t, err)
	require.Len(t, all, 3)
	assert.Equal(t, "5", all[0].ObjectID)
	assert.Equal(t, "4", all[1].ObjectID)
	assert.Equal(t, "3", all[2].ObjectID)
}

func TestDefaultHistoryManager(t *testing.T) {
	ctx := context.Background()
	mgr := NewDefaultHistoryManager()

	entry := LogEntry{
		ModelName:  "posts",
		ObjectID:   "42",
		ObjectRepr: "Post 42",
		Action:     ActionAdd,
	}
	require.NoError(t, mgr.LogAction(ctx, entry))

	history, err := mgr.GetHistory(ctx, "posts", "42")
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, "Post 42", history[0].ObjectRepr)
}

type dummyUser struct {
	ID       int64
	Username string
}

func TestAdmin_ResolveUserID(t *testing.T) {
	admin := &Admin[any]{name: "test"}

	assert.Equal(t, "anonymous", admin.resolveUserID(nil))
	assert.Equal(t, int64(101), admin.resolveUserID(dummyUser{ID: 101, Username: "admin"}))
	assert.Equal(t, int64(102), admin.resolveUserID(&dummyUser{ID: 102, Username: "staff"}))
	assert.Equal(t, "plain_string_user", admin.resolveUserID("plain_string_user"))
}
