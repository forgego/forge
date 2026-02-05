package exporters

import (
	"os"
	
	"go.uber.org/zap/zapcore"
)

// FileExporter creates a file output exporter with rotation
type FileExporter struct {
	writer zapcore.WriteSyncer
	level  zapcore.Level
	file   *os.File // Keep reference for closing
}

// FileConfig configures file output
type FileConfig struct {
	Path     string
	MaxSize  int // MB
	MaxAge   int // days
	MaxBackups int
	Compress bool
}

// NewFileExporter creates a new file exporter with rotation
// Note: This requires lumberjack.v2 to be added to go.mod
// For now, returns a basic file exporter without rotation
func NewFileExporter(config FileConfig, level zapcore.Level) *FileExporter {
	// TODO: Once lumberjack.v2 is added to go.mod, uncomment:
	// lj := &lumberjack.Logger{
	// 	Filename:   config.Path,
	// 	MaxSize:    config.MaxSize,
	// 	MaxAge:     config.MaxAge,
	// 	MaxBackups: config.MaxBackups,
	// 	Compress:   config.Compress,
	// }
	// return &FileExporter{
	// 	writer: lj,
	// 	level:  level,
	// }
	
	// Temporary: Use basic file (rotation will be added when dependency is available)
	file, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Return a no-op writer if file can't be opened
		return &FileExporter{
			writer: zapcore.AddSync(&noOpFile{}),
			level:  level,
			file:   nil,
		}
	}
	
	return &FileExporter{
		writer: zapcore.AddSync(file),
		level:  level,
		file:   file,
	}
}

// noOpFile is a placeholder for when file can't be opened
type noOpFile struct{}

func (f *noOpFile) Write(p []byte) (int, error) { return len(p), nil }
func (f *noOpFile) Sync() error                 { return nil }
func (f *noOpFile) Close() error                { return nil }

// GetWriter returns the write syncer
func (e *FileExporter) GetWriter() zapcore.WriteSyncer {
	return e.writer
}

// GetLevel returns the log level
func (e *FileExporter) GetLevel() zapcore.Level {
	return e.level
}

// Close closes the file exporter
func (e *FileExporter) Close() error {
	// Sync the writer
	if err := e.writer.Sync(); err != nil {
		return err
	}
	// Close the file if we have a reference
	if e.file != nil {
		return e.file.Close()
	}
	return nil
}
