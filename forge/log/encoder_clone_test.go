package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestConsoleEncoder_Clone(t *testing.T) {
	devConfig := DevelopmentConfig{
		Colored:    false,
		OneLine:    true,
		Caller:     false,
		Stacktrace: false,
	}
	enc := NewConsoleEncoder(devConfig)

	// Verify Clone copies all config fields
	cloned := enc.Clone()
	clonedConsole, ok := cloned.(*ConsoleEncoder)
	require.True(t, ok, "cloned encoder must be *ConsoleEncoder")
	assert.Equal(t, enc.colored, clonedConsole.colored)
	assert.Equal(t, enc.oneLine, clonedConsole.oneLine)
	assert.Equal(t, enc.caller, clonedConsole.caller)
	assert.Equal(t, enc.stacktrace, clonedConsole.stacktrace)

	// Test with logger: parent vs child With(...)
	parentBuf := &bytes.Buffer{}
	parentCore := zapcore.NewCore(enc, zapcore.AddSync(parentBuf), zapcore.DebugLevel)
	parentLogger := &Logger{
		Logger: zap.New(parentCore),
		config: DefaultLoggingConfig(true),
	}
	parentLogger.Info("hello")
	parentOutput := parentBuf.String()

	childBuf := &bytes.Buffer{}
	childCore := zapcore.NewCore(enc, zapcore.AddSync(childBuf), zapcore.DebugLevel)
	childLogger := (&Logger{
		Logger: zap.New(childCore),
		config: DefaultLoggingConfig(true),
	}).With(zap.String("k", "v"))
	childLogger.Info("hello")
	childOutput := childBuf.String()

	// Both parent and child must have the custom one-line format (timestamp LEVEL message\n, no tabs).
	// Default zap console encoder uses tab characters (\t) and JSON fields.
	customFormatPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} INFO hello\n$`)
	childFormatPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3} INFO hello \| k=v\n$`)
	assert.Regexp(t, customFormatPattern, parentOutput)
	assert.Regexp(t, childFormatPattern, childOutput)
	assert.NotContains(t, childOutput, "\t")
}

func TestProductionEncoder_Clone(t *testing.T) {
	prodConfig := ProductionConfig{
		Caller:     true,
		Stacktrace: false,
	}
	enc := NewProductionEncoder(prodConfig)

	// Verify Clone copies all config fields
	cloned := enc.Clone()
	clonedProd, ok := cloned.(*ProductionEncoder)
	require.True(t, ok, "cloned encoder must be *ProductionEncoder")
	assert.Equal(t, enc.caller, clonedProd.caller)
	assert.Equal(t, enc.stacktrace, clonedProd.stacktrace)

	// Test with logger: child With(...) retains TraceLowercaseLevelEncoder and caller field
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(enc, zapcore.AddSync(buf), zapcore.DebugLevel-1)
	logger := (&Logger{
		Logger: zap.New(core, zap.AddCaller()),
		config: DefaultLoggingConfig(false),
	}).With(zap.String("k", "v"))

	logger.Trace("prod trace")
	output := buf.String()
	// Custom format has lowercase "trace" and caller field
	assert.Contains(t, output, `"level":"trace"`)
	assert.Contains(t, output, `"caller":`)
	assert.Contains(t, output, `"k":"v"`)
}
