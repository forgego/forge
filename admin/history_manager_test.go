package admin

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/stores"
	"github.com/forgego/forge/stores/adminstore"
	"github.com/forgego/forge/stores/storestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryManagerConcurrentAccess(t *testing.T) {
	database := storestest.OpenMemory(t)
	shared, err := stores.New(database)
	require.NoError(t, err)

	m := &HistoryManager{}
	m.SetHistoryBackend(adminstore.NewHistory(shared))
	ctx := context.Background()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				_ = m.LogAction(ctx, core.LogEntry{
					Timestamp: time.Now(),
					ModelName: "User",
					ObjectID:  fmt.Sprintf("%d", idx),
					Action:    core.ActionAdd,
				})
			} else {
				_, _ = m.GetHistory(ctx, "User", fmt.Sprintf("%d", idx))
			}
		}(i)
	}

	wg.Wait()
}

func TestNewHistoryManager(t *testing.T) {
	database := storestest.OpenMemory(t)
	shared, err := stores.New(database)
	require.NoError(t, err)

	m := NewHistoryManager("title", "price")
	assert.Equal(t, []string{"title", "price"}, m.TrackFields)
	m.SetHistoryBackend(adminstore.NewHistory(shared))

	ctx := context.Background()
	err = m.LogAction(ctx, core.LogEntry{
		Timestamp: time.Now(),
		ModelName: "Product",
		ObjectID:  "1",
		Action:    core.ActionAdd,
	})
	require.NoError(t, err)

	history, err := m.GetHistory(ctx, "Product", "1")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestHistoryManager_NoBackendReturnsErrNoHistoryBackend(t *testing.T) {
	m := NewHistoryManager("title")
	ctx := context.Background()
	entry := core.LogEntry{
		ModelName: "Product",
		ObjectID:  "1",
		Action:    core.ActionAdd,
	}
	err := m.LogAction(ctx, entry)
	require.ErrorIs(t, err, ErrNoHistoryBackend)
	assert.Equal(t, "admin history needs a database: call Site.SetDB", err.Error())

	err = m.LogActionWithTx(ctx, nil, entry)
	require.ErrorIs(t, err, ErrNoHistoryBackend)

	history, err := m.GetHistory(ctx, "Product", "1")
	require.ErrorIs(t, err, ErrNoHistoryBackend)
	assert.Nil(t, history)
}
