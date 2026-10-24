package core

import (
	"context"
	"time"

	"github.com/forgego/forge/db"
)

// LogEntry represents a single change in the admin system
type LogEntry struct {
	ID          int64      `json:"id"`
	Timestamp   time.Time  `json:"timestamp"`
	UserID      string     `json:"user_id"`
	UserName    string     `json:"user_name"`
	ModelName   string     `json:"model_name"`
	ObjectID    string     `json:"object_id"`
	ObjectRepr  string     `json:"object_repr"`
	Action      ActionType `json:"action"`
	ChangeStats string     `json:"change_stats"` // JSON of changed fields
}

type ActionType string

const (
	ActionAdd    ActionType = "add"
	ActionChange ActionType = "change"
	ActionDelete ActionType = "delete"
)

// HistoryManager handles audit logging
type HistoryManager interface {
	LogAction(ctx context.Context, entry LogEntry) error
	GetHistory(ctx context.Context, modelName string, objectID string) ([]LogEntry, error)
}

// TransactionalHistoryManager is optionally implemented by history stores
// that can write using the admin mutation's database transaction.
type TransactionalHistoryManager interface {
	LogActionWithTx(ctx context.Context, tx *db.Tx, entry LogEntry) error
}

// HistoryBackend is implemented by history managers that keep their
// entries in another HistoryManager once one is set, such as the
// compatibility admin.HistoryManager. UseSiteHistory sets it.
type HistoryBackend interface {
	SetHistoryBackend(HistoryManager)
}

// UseSiteHistory makes the site's shared history manager this admin's
// history, unless the config chose its own. A config that left
// HistoryManager unset switches from the in-memory default to hm; a config
// whose manager implements HistoryBackend keeps it and has it write to hm.
func (a *Admin[T]) UseSiteHistory(hm HistoryManager) {
	if hm == nil {
		return
	}
	if a.defaultHistory {
		a.config.HistoryManager = hm
		return
	}
	if backend, ok := a.config.HistoryManager.(HistoryBackend); ok {
		backend.SetHistoryBackend(hm)
	}
}
