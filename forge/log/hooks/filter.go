package hooks

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// FilterHook filters logs based on patterns
type FilterHook struct {
	excludeLevels []zapcore.Level
	excludeKeys   []string
	excludeValues []string
}

// NewFilterHook creates a new filter hook
func NewFilterHook() *FilterHook {
	return &FilterHook{
		excludeLevels: make([]zapcore.Level, 0),
		excludeKeys:   make([]string, 0),
		excludeValues: make([]string, 0),
	}
}

// ExcludeLevel excludes logs at a specific level
func (h *FilterHook) ExcludeLevel(level zapcore.Level) *FilterHook {
	h.excludeLevels = append(h.excludeLevels, level)
	return h
}

// ExcludeKey excludes logs with a specific key
func (h *FilterHook) ExcludeKey(key string) *FilterHook {
	h.excludeKeys = append(h.excludeKeys, key)
	return h
}

// ExcludeValue excludes logs with a specific value
func (h *FilterHook) ExcludeValue(value string) *FilterHook {
	h.excludeValues = append(h.excludeValues, value)
	return h
}

// Process processes a log entry for filtering
func (h *FilterHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	// Check level exclusion
	for _, level := range h.excludeLevels {
		if entry.Level == level {
			return entry, fields, false
		}
	}

	// Check key exclusion
	for _, field := range fields {
		for _, excludeKey := range h.excludeKeys {
			if field.Key == excludeKey {
				return entry, fields, false
			}
		}
	}

	// Check value exclusion
	for _, field := range fields {
		fieldValue := field.String
		for _, excludeValue := range h.excludeValues {
			if strings.Contains(fieldValue, excludeValue) {
				return entry, fields, false
			}
		}
	}

	return entry, fields, true
}
