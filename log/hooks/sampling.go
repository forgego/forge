package hooks

import (
	"sync/atomic"

	"go.uber.org/zap/zapcore"
)

// SamplingHook implements log sampling
type SamplingHook struct {
	initial    int
	thereafter int
	count      atomic.Int64
}

// NewSamplingHook creates a new sampling hook
func NewSamplingHook(initial, thereafter int) *SamplingHook {
	return &SamplingHook{
		initial:    initial,
		thereafter: thereafter,
	}
}

// Process processes a log entry for sampling
func (h *SamplingHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	current := int(h.count.Add(1))

	// Always log errors and above
	if entry.Level >= zapcore.ErrorLevel {
		return entry, fields, true
	}

	// Sample based on count
	if current <= h.initial {
		return entry, fields, true
	}

	// Sample every Nth log after initial
	if h.thereafter > 0 && (current-h.initial)%h.thereafter == 0 {
		return entry, fields, true
	}

	return entry, fields, false
}
