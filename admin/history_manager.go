package admin

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/forgego/forge/admin/core"
	"github.com/forgego/forge/db"
)

// ErrNoHistoryBackend is returned when a HistoryManager has no database backend configured.
var ErrNoHistoryBackend = errors.New("admin history needs a database: call Site.SetDB")

// HistoryManager is a compatibility helper for legacy configs.
// It satisfies core.HistoryManager and forwards to the site's shared
// history once the site uses database stores.
type HistoryManager struct {
	TrackFields []string
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
	return nil
}

// NewHistoryManager creates a new HistoryManager.
func NewHistoryManager(trackFields ...string) *HistoryManager {
	return &HistoryManager{
		TrackFields: trackFields,
	}
}

func (m *HistoryManager) LogAction(ctx context.Context, entry core.LogEntry) error {
	if t := m.target(); t != nil {
		return t.LogAction(ctx, entry)
	}
	return ErrNoHistoryBackend
}

// LogActionWithTx forwards to a transaction-aware backend when available.
// Other backends retain the legacy LogAction behavior.
func (m *HistoryManager) LogActionWithTx(ctx context.Context, tx *db.Tx, entry core.LogEntry) error {
	target := m.target()
	if target == nil {
		return ErrNoHistoryBackend
	}
	if transactional, ok := target.(core.TransactionalHistoryManager); ok {
		return transactional.LogActionWithTx(ctx, tx, entry)
	}
	return target.LogAction(ctx, entry)
}

func (m *HistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]core.LogEntry, error) {
	if t := m.target(); t != nil {
		return t.GetHistory(ctx, modelName, objectID)
	}
	return nil, ErrNoHistoryBackend
}
