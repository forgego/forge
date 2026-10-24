package core

import (
	"context"
	"strings"
	"sync"
	"time"
)

// MemoryHistoryManager implements an in-memory thread-safe HistoryManager for tests.
type MemoryHistoryManager struct {
	mu      sync.RWMutex
	entries []LogEntry
	nextID  int64
	maxSize int
}

// NewMemoryHistoryManager creates a new thread-safe in-memory history manager for tests.
func NewMemoryHistoryManager() *MemoryHistoryManager {
	return &MemoryHistoryManager{
		entries: make([]LogEntry, 0),
		maxSize: 1000,
	}
}

func (m *MemoryHistoryManager) LogAction(ctx context.Context, entry LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	m.nextID++
	entry.ID = m.nextID

	m.entries = append(m.entries, entry)
	if m.maxSize > 0 && len(m.entries) > m.maxSize {
		m.entries = m.entries[len(m.entries)-m.maxSize:]
	}
	return nil
}

func (m *MemoryHistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]LogEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []LogEntry
	// Return in reverse chronological order (newest first)
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := m.entries[i]
		if modelName != "" && !strings.EqualFold(e.ModelName, modelName) {
			continue
		}
		if objectID != "" && e.ObjectID != objectID {
			continue
		}
		results = append(results, e)
	}

	if results == nil {
		results = []LogEntry{}
	}
	return results, nil
}

// DefaultHistoryManager implements a default in-memory history manager for tests.
type DefaultHistoryManager struct {
	mem *MemoryHistoryManager
}

func NewDefaultHistoryManager() *DefaultHistoryManager {
	return &DefaultHistoryManager{
		mem: NewMemoryHistoryManager(),
	}
}

func (m *DefaultHistoryManager) getMem() *MemoryHistoryManager {
	if m.mem == nil {
		m.mem = NewMemoryHistoryManager()
	}
	return m.mem
}

func (m *DefaultHistoryManager) LogAction(ctx context.Context, entry LogEntry) error {
	return m.getMem().LogAction(ctx, entry)
}

func (m *DefaultHistoryManager) GetHistory(ctx context.Context, modelName string, objectID string) ([]LogEntry, error) {
	return m.getMem().GetHistory(ctx, modelName, objectID)
}
