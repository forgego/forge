package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type passThroughHook struct{}

func (h passThroughHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	return entry, fields, true
}

type dropSecretKeyHook struct{}

func (h dropSecretKeyHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	for _, f := range fields {
		if f.Key == "secret" {
			return entry, fields, false
		}
	}
	return entry, fields, true
}

type fnHook struct {
	fn func(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool)
}

func (h fnHook) Process(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
	return h.fn(entry, fields)
}

// Finding 1: HookCore.Check must consult wrapped core's Check so sampling works underneath.
func TestHookCore_CheckConsultsWrappedCore_Sampling(t *testing.T) {
	obsCore, recorded := observer.New(zapcore.InfoLevel)
	// Initial=1, Thereafter=0 means only the 1st entry is logged; all subsequent identical entries are dropped.
	samplerCore := zapcore.NewSamplerWithOptions(obsCore, time.Minute, 1, 0)
	registry := NewHookRegistry()
	registry.AddHook(passThroughHook{})
	hookCore := NewHookCore(samplerCore, registry)
	logger := zap.New(hookCore)

	for i := 0; i < 5; i++ {
		logger.Info("same info message")
	}

	assert.Equal(t, 1, recorded.Len(), "expected exactly 1 write due to sampling underneath HookCore")
}

// Finding 2: HookCore.With must pass bound fields to hooks, and not duplicate them in output.
func TestHookCore_With_BoundFieldsVisibleToHook(t *testing.T) {
	obsCore, recorded := observer.New(zapcore.InfoLevel)
	registry := NewHookRegistry()
	registry.AddHook(dropSecretKeyHook{})
	hookCore := NewHookCore(obsCore, registry)
	logger := zap.New(hookCore)

	// logger.With secret field should be dropped by hook
	logger.With(zap.String("secret", "x")).Info("m")
	assert.Equal(t, 0, recorded.Len(), "hook should have seen bound 'secret' field and dropped the entry")

	// logger without secret field should write
	logger.Info("m")
	assert.Equal(t, 1, recorded.Len(), "entry without 'secret' field should be written")
}

func TestHookCore_With_OutputDoesNotDuplicateFields(t *testing.T) {
	obsCore, recorded := observer.New(zapcore.InfoLevel)
	registry := NewHookRegistry()
	registry.AddHook(passThroughHook{})
	hookCore := NewHookCore(obsCore, registry)
	logger := zap.New(hookCore)

	logger.With(zap.String("bound_key", "bound_val")).Info("m", zap.String("call_key", "call_val"))
	require.Equal(t, 1, recorded.Len())

	entryFields := recorded.All()[0].Context
	boundCount := 0
	callCount := 0
	for _, f := range entryFields {
		if f.Key == "bound_key" {
			boundCount++
		}
		if f.Key == "call_key" {
			callCount++
		}
	}
	assert.Equal(t, 1, boundCount, "bound field must appear exactly once")
	assert.Equal(t, 1, callCount, "call field must appear exactly once")
}

// Finding 3: NewLoggerFromConfig must pass config.Hooks to newLoggerFromConfig.
func TestNewLoggerFromConfig_InvokesConfigHooks(t *testing.T) {
	invoked := false
	hook := fnHook{
		fn: func(entry zapcore.Entry, fields []zapcore.Field) (zapcore.Entry, []zapcore.Field, bool) {
			invoked = true
			return entry, fields, true
		},
	}

	cfg := DefaultLoggingConfig(false)
	cfg.Hooks = []Hook{hook}
	logger, err := NewLoggerFromConfig(cfg)
	require.NoError(t, err)

	logger.Info("trigger hook")
	assert.True(t, invoked, "hook in LoggingConfig.Hooks must be invoked")
}

func TestHookCore_Check_PerCoreLevelFilteringInTee(t *testing.T) {
	infoCore, infoLogs := observer.New(zapcore.InfoLevel)
	errorCore, errorLogs := observer.New(zapcore.ErrorLevel)

	tee := zapcore.NewTee(infoCore, errorCore)
	registry := NewHookRegistry()
	registry.AddHook(passThroughHook{})
	hookCore := NewHookCore(tee, registry)
	logger := zap.New(hookCore)

	logger.Info("info message")

	assert.Equal(t, 1, infoLogs.Len(), "info entry must reach info core")
	assert.Equal(t, 0, errorLogs.Len(), "info entry must NOT reach error core")
}
