package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	logger, err := NewLogger(true)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	defer logger.Sync()
}

func TestNewLoggerFromConfig(t *testing.T) {
	config := DefaultLoggingConfig(true)
	logger, err := NewLoggerFromConfig(config)
	if err != nil {
		t.Fatalf("NewLoggerFromConfig failed: %v", err)
	}
	if logger == nil {
		t.Fatal("NewLoggerFromConfig returned nil")
	}
	defer logger.Sync()
}

func TestNewNopLogger(t *testing.T) {
	logger := NewNopLogger()
	if logger == nil {
		t.Fatal("NewNopLogger returned nil")
	}
}

func TestLoggerWithFields(t *testing.T) {
	logger, _ := NewLogger(true)
	defer logger.Sync()

	childLogger := logger.With(String("key", "value"))
	if childLogger == nil {
		t.Fatal("With returned nil")
	}
}

func TestLoggerTrace(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel - 1)
	logger := &Logger{
		Logger: zap.New(core),
		config: DefaultLoggingConfig(true),
	}
	defer logger.Sync()

	logger.Trace("trace message")

	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, "trace message", entry.Message)
	assert.Equal(t, zapcore.DebugLevel-1, entry.Level)
}

func TestQuickLogger(t *testing.T) {
	logger := QuickLogger()
	if logger == nil {
		t.Fatal("QuickLogger returned nil")
	}
	defer logger.Sync()
}

func TestProductionLogger(t *testing.T) {
	// Use a temp file for testing
	tmpFile := os.TempDir() + "/test.log"
	defer os.Remove(tmpFile)

	logger, err := ProductionLogger(tmpFile)
	if err != nil {
		t.Fatalf("ProductionLogger failed: %v", err)
	}
	if logger == nil {
		t.Fatal("ProductionLogger returned nil")
	}
	defer logger.Sync()
}

func TestBuilder(t *testing.T) {
	builder := NewBuilder()
	builder.Development()
	builder.Level(LevelDebug)
	builder.AddConsoleOutput(LevelDebug)

	logger, err := builder.Build()
	if err != nil {
		t.Fatalf("Builder.Build failed: %v", err)
	}
	if logger == nil {
		t.Fatal("Builder.Build returned nil")
	}
	defer logger.Sync()
}

func TestBuilderWithFile(t *testing.T) {
	tmpFile := os.TempDir() + "/builder_test.log"
	defer os.Remove(tmpFile)

	builder := NewBuilder()
	builder.Production()
	builder.AddFileOutput(tmpFile, LevelInfo, 10, 7, 3, false)

	logger, err := builder.Build()
	if err != nil {
		t.Fatalf("Builder with file failed: %v", err)
	}
	if logger == nil {
		t.Fatal("Builder with file returned nil")
	}
	defer logger.Sync()
}

func TestLoggerCloseFlushesAndClosesFileOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	logger, err := NewBuilder().Production().AddFileOutput(path, LevelInfo, 1, 1, 1, false).Build()
	require.NoError(t, err)

	logger.Info("persisted entry")
	require.NoError(t, logger.Close())
	require.NoError(t, logger.Close())

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "persisted entry")
}
