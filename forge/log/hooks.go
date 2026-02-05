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
}

// NewHookCore creates a new hook core
func NewHookCore(core zapcore.Core, registry *HookRegistry) *HookCore {
	return &HookCore{
		Core:     core,
		registry: registry,
	}
}

// Write processes hooks before writing
func (c *HookCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	entry, fields, shouldLog := c.registry.ProcessHooks(entry, fields)
	if !shouldLog {
		return nil // Skip logging
	}
	return c.Core.Write(entry, fields)
}
