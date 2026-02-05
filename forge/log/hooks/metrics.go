package hooks

import (
	"go.uber.org/zap/zapcore"
)

// MetricsHook collects metrics from log entries
type MetricsHook struct {
	// Metrics collection will be implemented when metrics system is added
	// For now, this is a placeholder
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{}
}

// Process processes a log entry for metrics collection
func (h *MetricsHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	// TODO: Collect metrics based on log level, message, etc.
	// This will be implemented when the metrics system is integrated
	return entry, fields, true
}
