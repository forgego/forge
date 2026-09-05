package admin

import (
	"context"

	"github.com/forgego/forge/admin/core"
)

// HistoryManager is a compatibility helper for legacy configs.
// It satisfies core.HistoryManager but does not persist history.
type HistoryManager struct {
	TrackFields []string
}

func (m *HistoryManager) LogAction(ctx context.Context, entry core.LogEntry) error {
	return nil
}

func (m *HistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	return []core.LogEntry{}, nil
}
