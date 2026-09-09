package admin

import (
	"context"

	"github.com/forgego/forge/admin/core"
)

// HistoryManager is a compatibility helper for legacy configs.
// It satisfies core.HistoryManager and persists history using an in-memory manager.
type HistoryManager struct {
	TrackFields []string
	mem         *core.MemoryHistoryManager
}

func (m *HistoryManager) getMem() *core.MemoryHistoryManager {
	if m.mem == nil {
		m.mem = core.NewMemoryHistoryManager()
	}
	return m.mem
}

func (m *HistoryManager) LogAction(ctx context.Context, entry core.LogEntry) error {
	return m.getMem().LogAction(ctx, entry)
}

func (m *HistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	return m.getMem().GetHistory(ctx, modelName, objectID)
}
