package admin

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/forgego/forge/admin/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryManagerConcurrentAccess(t *testing.T) {
	m := &HistoryManager{}
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
	m := NewHistoryManager("title", "price")
	assert.Equal(t, []string{"title", "price"}, m.TrackFields)

	ctx := context.Background()
	err := m.LogAction(ctx, core.LogEntry{
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
