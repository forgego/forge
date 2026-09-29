package admin

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/forgego/forge/admin/core"
)

// HistoryManager is a compatibility helper for legacy configs.
// It satisfies core.HistoryManager and keeps history in memory, or in the
// site's shared history once the site uses database stores
// (Site.UseDatabaseStores).
type HistoryManager struct {
	TrackFields []string
	once        sync.Once
	mem         *core.MemoryHistoryManager
	backend     atomic.Pointer[core.HistoryManager]
}

// SetHistoryBackend implements core.HistoryBackend: entries go to hm from
// now on.
func (m *HistoryManager) SetHistoryBackend(hm core.HistoryManager) {
	if hm != nil {
		m.backend.Store(&hm)
	}
}

func (m *HistoryManager) target() core.HistoryManager {
	if hm := m.backend.Load(); hm != nil {
		return *hm
	}
	return m.getMem()
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
	return m.target().LogAction(ctx, entry)
}

func (m *HistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	return m.target().GetHistory(ctx, modelName, objectID)
}
