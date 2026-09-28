package exporters

import (
	"go.uber.org/zap/zapcore"
)

// MultiExporter combines multiple exporters
type MultiExporter struct {
	exporters []Exporter
	level     zapcore.Level
}

// Exporter is the interface for log exporters
type Exporter interface {
	GetWriter() zapcore.WriteSyncer
	GetLevel() zapcore.Level
}

// NewMultiExporter creates a new multi-exporter
func NewMultiExporter(exporters []Exporter, level zapcore.Level) *MultiExporter {
	return &MultiExporter{
		exporters: exporters,
		level:     level,
	}
}

// GetWriters returns all write syncers
func (e *MultiExporter) GetWriters() []zapcore.WriteSyncer {
	writers := make([]zapcore.WriteSyncer, len(e.exporters))
	for i, exp := range e.exporters {
		writers[i] = exp.GetWriter()
	}
	return writers
}

// GetLevel returns the log level
func (e *MultiExporter) GetLevel() zapcore.Level {
	return e.level
}
