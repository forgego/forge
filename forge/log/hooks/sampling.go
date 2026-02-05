package hooks

import (
	"go.uber.org/zap/zapcore"
)

// SamplingHook implements log sampling
type SamplingHook struct {
	initial   int
	thereafter int
	count     int
}

// NewSamplingHook creates a new sampling hook
func NewSamplingHook(initial, thereafter int) *SamplingHook {
	return &SamplingHook{
		initial:    initial,
		thereafter: thereafter,
		count:      0,
	}
}

// Process processes a log entry for sampling
func (h *SamplingHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	h.count++

	// Always log errors and above
	if entry.Level >= zapcore.ErrorLevel {
		return entry, fields, true
	}

	// Sample based on count
	if h.count <= h.initial {
		return entry, fields, true
	}

	// Sample every Nth log after initial
	if (h.count-h.initial)%h.thereafter == 0 {
		return entry, fields, true
	}

	return entry, fields, false
}
