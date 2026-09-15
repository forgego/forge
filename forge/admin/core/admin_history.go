package core

import (
	"context"
	"fmt"
	"reflect"
	"time"
)

// History methods
func (a *Admin[T]) GetHistory(ctx context.Context, objectID string) ([]LogEntry, error) {
	if a.config.HistoryManager != nil {
		return a.config.HistoryManager.GetHistory(ctx, a.name, objectID)
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
			ChangeStats: changes,
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
