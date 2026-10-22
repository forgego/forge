package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/forgego/forge/schema"
	"reflect"
	"time"
)

// History methods
func (a *Admin[T]) GetHistory(ctx context.Context, objectID string) ([]LogEntry, error) {
	if a.config.HistoryManager != nil {
		entries, err := a.config.HistoryManager.GetHistory(ctx, a.name, objectID)
		if err != nil {
			return nil, err
		}
		// Redact a copy: reading old history must not mutate the backend.
		visible := append([]LogEntry(nil), entries...)
		for i := range visible {
			visible[i].ChangeStats = a.visibleHistoryChanges(visible[i].ChangeStats)
		}
		return visible, nil
	}
	return []LogEntry{}, nil
}

func (a *Admin[T]) LogAction(ctx context.Context, user interface{}, objectID string, repr string, action ActionType, changes string) error {
	if a.config.HistoryManager != nil {
		entry := LogEntry{
			Timestamp:   time.Now(),
			ModelName:   a.name,
			ObjectID:    objectID,
			ObjectRepr:  repr,
			Action:      action,
			ChangeStats: a.visibleHistoryChanges(changes),
			UserID:      fmt.Sprintf("%v", a.resolveUserID(user)),
		}

		return a.config.HistoryManager.LogAction(ctx, entry)
	}
	return nil
}

// resolveUserID tries to extract a string identifier from the user object
func (a *Admin[T]) resolveUserID(user interface{}) interface{} {
	if user == nil {
		return "anonymous"
	}

	// The admin API's auth middleware stores the user as a map.
	if m, ok := user.(map[string]interface{}); ok {
		for _, key := range []string{"id", "username", "name", "email"} {
			if v, exists := m[key]; exists && v != nil && fmt.Sprintf("%v", v) != "" {
				return v
			}
		}
		return "anonymous"
	}

	val := reflect.ValueOf(user)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 1. Try common ID fields
	if val.Kind() == reflect.Struct {
		for _, name := range []string{"ID", "Id", "id", "UserID", "Username", "Email"} {
			f := val.FieldByName(name)
			if f.IsValid() {
				return f.Interface()
			}
		}
	}

	// 2. Fallback to string representation
	return fmt.Sprintf("%v", user)
}

func (a *Admin[T]) visibleHistoryChanges(changes string) string {
	hidden := schema.NonSerializableFieldNames(new(T))
	if len(hidden) == 0 || changes == "" {
		return changes
	}
	var fields map[string]json.RawMessage
	// A non-JSON legacy payload cannot be safely separated into field values.
	if json.Unmarshal([]byte(changes), &fields) != nil {
		return "{}"
	}
	for _, name := range hidden {
		delete(fields, name)
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return "{}"
	}
	return string(data)
}
