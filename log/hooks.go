package log

import (
	"go.uber.org/zap/zapcore"
)

// Hook is an interface for log hooks
type Hook interface {
	// Process processes a log entry before it's written
	Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool)
}

// HookRegistry maintains a registry of hooks
type HookRegistry struct {
	hooks []Hook
}

// NewHookRegistry creates a new hook registry
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make([]Hook, 0),
	}
}

// AddHook adds a hook to the registry
func (r *HookRegistry) AddHook(hook Hook) {
	r.hooks = append(r.hooks, hook)
}

// ProcessHooks processes all hooks for a log entry
func (r *HookRegistry) ProcessHooks(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	shouldLog := true
	for _, hook := range r.hooks {
		var ok bool
		entry, fields, ok = hook.Process(entry, fields)
		if !ok {
			shouldLog = false
			break
		}
	}
	return entry, fields, shouldLog
}

// HookCore wraps a core with hook processing
type HookCore struct {
	zapcore.Core
	registry *HookRegistry
	bound    []zapcore.Field
}

// NewHookCore creates a new hook core
func NewHookCore(core zapcore.Core, registry *HookRegistry) *HookCore {
	return &HookCore{
		Core:     core,
		registry: registry,
	}
}

// Check determines whether the entry should be logged
func (c *HookCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return ce
	}
	downstream := c.Core.Check(entry, nil)
	if downstream == nil {
		return ce
	}
	return ce.AddCore(entry, &hookedCheckedWrite{core: c, downstream: downstream})
}

// With adds structured context to the Core
func (c *HookCore) With(fields []zapcore.Field) zapcore.Core {
	bound := make([]zapcore.Field, len(c.bound)+len(fields))
	copy(bound, c.bound)
	copy(bound[len(c.bound):], fields)
	return &HookCore{
		Core:     c.Core.With(fields),
		registry: c.registry,
		bound:    bound,
	}
}

// Write processes hooks before writing
func (c *HookCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	downstream := c.Core.Check(entry, nil)
	if downstream == nil {
		return nil
	}
	hw := &hookedCheckedWrite{core: c, downstream: downstream}
	return hw.Write(entry, fields)
}

type hookedCheckedWrite struct {
	core       *HookCore
	downstream *zapcore.CheckedEntry
}

func (h *hookedCheckedWrite) Enabled(lvl zapcore.Level) bool {
	return h.core.Enabled(lvl)
}

func (h *hookedCheckedWrite) With(fields []zapcore.Field) zapcore.Core {
	return h.core.With(fields)
}

func (h *hookedCheckedWrite) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	return ce.AddCore(entry, h)
}

func (h *hookedCheckedWrite) Sync() error {
	return h.core.Sync()
}

func (h *hookedCheckedWrite) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	var allFields []zapcore.Field
	if len(h.core.bound) == 0 {
		allFields = fields
	} else if len(fields) == 0 {
		allFields = h.core.bound
	} else {
		allFields = make([]zapcore.Field, len(h.core.bound)+len(fields))
		copy(allFields, h.core.bound)
		copy(allFields[len(h.core.bound):], fields)
	}

	entry, processedFields, shouldLog := h.core.registry.ProcessHooks(entry, allFields)
	if !shouldLog {
		return nil // Skip logging
	}

	var outFields []zapcore.Field
	if len(processedFields) >= len(h.core.bound) {
		outFields = processedFields[len(h.core.bound):]
	}

	downstream := h.downstream
	if entry.Level != downstream.Entry.Level {
		// A changed level can select different outputs. Preserve the original
		// check otherwise, so samplers only count the entry once.
		downstream = h.core.Core.Check(entry, nil)
		if downstream == nil {
			return nil
		}
	}
	downstream.Entry = entry
	downstream.Write(outFields...)
	return nil
}
