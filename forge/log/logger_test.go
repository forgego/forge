package log

import (
	"os"
	"testing"
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
	logger, _ := NewLogger(true)
	defer logger.Sync()

	// Trace should not panic
	logger.Trace("trace message")
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
