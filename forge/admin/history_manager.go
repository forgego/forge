package admin

import (
	"context"
	"sync"

	"github.com/forgego/forge/admin/core"
)

// HistoryManager is a compatibility helper for legacy configs.
// It satisfies core.HistoryManager and persists history using an in-memory manager.
type HistoryManager struct {
	TrackFields []string
	once        sync.Once
	mem         *core.MemoryHistoryManager
}

// NewHistoryManager creates a new HistoryManager with an initialized memory store.
func NewHistoryManager(trackFields ...string) *HistoryManager {
	m := &HistoryManager{
		TrackFields: trackFields,
		mem:         core.NewMemoryHistoryManager(),
	}
	m.once.Do(func() {})
	return m
}

func (m *HistoryManager) getMem() *core.MemoryHistoryManager {
	m.once.Do(func() {
		if m.mem == nil {
			m.mem = core.NewMemoryHistoryManager()
		}
	})
	return m.mem
}

func (m *HistoryManager) LogAction(ctx context.Context, entry core.LogEntry) error {
	return m.getMem().LogAction(ctx, entry)
}

func (m *HistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	return m.getMem().GetHistory(ctx, modelName, objectID)
}
