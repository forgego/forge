package exporters

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap/zapcore"
)

// FileExporter creates a file output exporter with rotation
type FileExporter struct {
	writer zapcore.WriteSyncer
	level  zapcore.Level
	closer func() error
}

// FileConfig configures file output
type FileConfig struct {
	Path       string
	MaxSize    int // MB
	MaxAge     int // days
	MaxBackups int
	Compress   bool
}

// NewFileExporter creates a new file exporter with in-process size-based rotation.
func NewFileExporter(config FileConfig, level zapcore.Level) *FileExporter {
	writer, err := newRotatingFileWriter(config)
	if err != nil {
		// Return a no-op writer if file can't be opened
		return &FileExporter{
			writer: zapcore.AddSync(&noOpFile{}),
			level:  level,
			closer: func() error { return nil },
		}
	}

	return &FileExporter{
		writer: writer,
		level:  level,
		closer: writer.Close,
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
	return e.closer()
}

type rotatingFileWriter struct {
	mu           sync.Mutex
	path         string
	maxSizeBytes int64
	maxBackups   int
	file         *os.File
}

func newRotatingFileWriter(config FileConfig) (*rotatingFileWriter, error) {
	if config.Path == "" {
		return nil, fmt.Errorf("file path is required")
	}

	if err := os.MkdirAll(filepath.Dir(config.Path), 0755); err != nil {
		return nil, err
	}

	writer := &rotatingFileWriter{
		path:       config.Path,
		maxBackups: config.MaxBackups,
	}
	if config.MaxSize > 0 {
		writer.maxSizeBytes = int64(config.MaxSize) * 1024 * 1024
	}

	if err := writer.openFile(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureCapacity(int64(len(p))); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

func (w *rotatingFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingFileWriter) ensureCapacity(incomingBytes int64) error {
	if w.file == nil {
		if err := w.openFile(); err != nil {
			return err
		}
	}

	if w.maxSizeBytes <= 0 {
		return nil
	}

	info, err := w.file.Stat()
	if err != nil {
		return err
	}

	if info.Size()+incomingBytes <= w.maxSizeBytes {
		return nil
	}

	return w.rotate()
}

func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		_ = w.file.Sync()
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
	}

	if w.maxBackups > 0 {
		for i := w.maxBackups - 1; i >= 1; i-- {
			src := fmt.Sprintf("%s.%d", w.path, i)
			dst := fmt.Sprintf("%s.%d", w.path, i+1)

			_ = os.Remove(dst)
			_ = os.Rename(src, dst)
		}

		firstBackup := fmt.Sprintf("%s.1", w.path)
		_ = os.Remove(firstBackup)
		_ = os.Rename(w.path, firstBackup)
	} else {
		_ = os.Remove(w.path)
	}

	return w.openFile()
}

func (w *rotatingFileWriter) openFile() error {
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = file
	return nil
}

