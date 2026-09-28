package exporters

import (
	"os"

	"go.uber.org/zap/zapcore"
)

// ConsoleExporter creates a console output exporter
type ConsoleExporter struct {
	writer zapcore.WriteSyncer
	level  zapcore.Level
}

// NewConsoleExporter creates a new console exporter
func NewConsoleExporter(level zapcore.Level) *ConsoleExporter {
	return &ConsoleExporter{
		writer: zapcore.AddSync(os.Stderr),
		level:  level,
	}
}

// GetWriter returns the write syncer
func (e *ConsoleExporter) GetWriter() zapcore.WriteSyncer {
	return e.writer
}

// GetLevel returns the log level
func (e *ConsoleExporter) GetLevel() zapcore.Level {
	return e.level
}
