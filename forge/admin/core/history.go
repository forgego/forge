package core

import (
	"context"
	"time"
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

// DefaultHistoryManager implements a basic memory/null history (to be extended)
type DefaultHistoryManager struct{}

func (m *DefaultHistoryManager) LogAction(ctx context.Context, entry LogEntry) error {
	// Default: No-op or log to console
	return nil
}

func (m *DefaultHistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]LogEntry, error) {
	return []LogEntry{}, nil
}
