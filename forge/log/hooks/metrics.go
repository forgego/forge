package hooks

import (
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// MetricsHook collects metrics from log entries
type MetricsHook struct {
	total   uint64
	dropped uint64
	trace   uint64
	debug   uint64
	info    uint64
	warn    uint64
	error   uint64
	dpanic  uint64
	panic   uint64
	fatal   uint64
	unknown uint64
}

// MetricsSnapshot is a read-only view of collected log counters.
type MetricsSnapshot struct {
	Total   uint64
	Dropped uint64
	ByLevel map[zapcore.Level]uint64
	Unknown uint64
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{}
}

// Process processes a log entry for metrics collection
func (h *MetricsHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	atomic.AddUint64(&h.total, 1)

	switch entry.Level {
	case zapcore.DebugLevel - 1:
		atomic.AddUint64(&h.trace, 1)
	case zapcore.DebugLevel:
		atomic.AddUint64(&h.debug, 1)
	case zapcore.InfoLevel:
		atomic.AddUint64(&h.info, 1)
	case zapcore.WarnLevel:
		atomic.AddUint64(&h.warn, 1)
	case zapcore.ErrorLevel:
		atomic.AddUint64(&h.error, 1)
	case zapcore.DPanicLevel:
		atomic.AddUint64(&h.dpanic, 1)
	case zapcore.PanicLevel:
		atomic.AddUint64(&h.panic, 1)
	case zapcore.FatalLevel:
		atomic.AddUint64(&h.fatal, 1)
	default:
		atomic.AddUint64(&h.unknown, 1)
	}

	return entry, fields, true
}

// MarkDropped increments dropped-log counter for logs filtered by other hooks/cores.
func (h *MetricsHook) MarkDropped() {
	atomic.AddUint64(&h.dropped, 1)
}

// Snapshot returns current metrics counters.
func (h *MetricsHook) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Total:   atomic.LoadUint64(&h.total),
		Dropped: atomic.LoadUint64(&h.dropped),
		ByLevel: map[zapcore.Level]uint64{
			zapcore.DebugLevel - 1: atomic.LoadUint64(&h.trace),
			zapcore.DebugLevel:     atomic.LoadUint64(&h.debug),
			zapcore.InfoLevel:      atomic.LoadUint64(&h.info),
			zapcore.WarnLevel:      atomic.LoadUint64(&h.warn),
			zapcore.ErrorLevel:     atomic.LoadUint64(&h.error),
			zapcore.DPanicLevel:    atomic.LoadUint64(&h.dpanic),
			zapcore.PanicLevel:     atomic.LoadUint64(&h.panic),
			zapcore.FatalLevel:     atomic.LoadUint64(&h.fatal),
		},
		Unknown: atomic.LoadUint64(&h.unknown),
	}
}
