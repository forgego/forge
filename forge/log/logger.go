// Package log provides logging functionality
package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with enhanced functionality
type Logger struct {
	*zap.Logger
	config *LoggingConfig
}

// NewLogger creates a new logger with framework defaults
// This is kept for backward compatibility but uses the new system internally
func NewLogger(development bool) (*Logger, error) {
	config := DefaultLoggingConfig(development)
	return NewLoggerFromConfig(config)
}

// NewLoggerFromConfig creates a new logger from a configuration
func NewLoggerFromConfig(config *LoggingConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLoggingConfig(false)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logging configuration: %w", err)
	}

	// Build cores for each output
	cores := make([]zapcore.Core, 0)
	development := isDevelopmentMode(config)

	for _, output := range config.Outputs {
		if !output.Enabled {
			continue
		}

		// Get encoder for this output
		format := output.Format
		if format == "" {
			format = config.Format
		}
		encoder := getZapFormat(format, config, development)

		// Get level for this output
		level := getZapLevel(output.Level)

		// Create core based on output type
		var core zapcore.Core
		switch output.Type {
		case OutputConsole:
			core = createConsoleCore(encoder, level)
		case OutputFile:
			core = createFileCore(encoder, level, output.File)
		case OutputRemote:
			core = createRemoteCore(encoder, level, output.Remote)
		default:
			continue
		}

		cores = append(cores, core)
	}

	// If no cores were created, create a default console core
	if len(cores) == 0 {
		encoder := getZapFormat(config.Format, config, development)
		level := getZapLevel(config.Level)
		cores = append(cores, createConsoleCore(encoder, level))
	}

	// Combine cores
	combinedCore := zapcore.NewTee(cores...)

	// Apply sampling in production
	if !development && config.Production.Sampling.Enabled {
		combinedCore = zapcore.NewSamplerWithOptions(
			combinedCore,
			time.Second,
			config.Production.Sampling.Initial,
			config.Production.Sampling.Thereafter,
		)
	}

	// Create logger
	zapLogger := zap.New(
		combinedCore,
		zap.AddCaller(),
		zap.AddStacktrace(getStacktraceLevel(config, development)),
	)

	return &Logger{
		Logger: zapLogger,
		config: config,
	}, nil
}

// createConsoleCore creates a console core
func createConsoleCore(encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	writer := zapcore.AddSync(os.Stderr)
	return zapcore.NewCore(encoder, writer, level)
}

// createFileCore creates a file core with rotation
func createFileCore(encoder zapcore.Encoder, level zapcore.Level, fileConfig FileOutputConfig) zapcore.Core {
	// Import exporters package inline to avoid circular dependency issues
	// This will be resolved when the package structure is finalized
	fileExp := newFileExporter(fileConfig, level)
	writer := fileExp.GetWriter()
	return zapcore.NewCore(encoder, writer, level)
}

// createRemoteCore creates a remote core
func createRemoteCore(encoder zapcore.Encoder, level zapcore.Level, remoteConfig RemoteOutputConfig) zapcore.Core {
	remoteExp := newRemoteExporter(remoteConfig, level)
	writer := remoteExp.GetWriter()
	return zapcore.NewCore(encoder, writer, level)
}

// fileExporter wraps file exporter functionality
type fileExporter struct {
	writer zapcore.WriteSyncer
	level  zapcore.Level
}

func newFileExporter(config FileOutputConfig, level zapcore.Level) *fileExporter {
	// Note: lumberjack import will be added to go.mod
	// For now, we'll use a simple file writer
	// The full implementation with rotation will be in the exporters package
	file, err := os.OpenFile(config.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Fallback to stderr if file can't be opened
		return &fileExporter{
			writer: zapcore.AddSync(os.Stderr),
			level:  level,
		}
	}
	return &fileExporter{
		writer: zapcore.AddSync(file),
		level:  level,
	}
}

func (e *fileExporter) GetWriter() zapcore.WriteSyncer {
	return e.writer
}

// remoteExporter wraps remote exporter functionality
type remoteExporter struct {
	writer zapcore.WriteSyncer
	level  zapcore.Level
}

func newRemoteExporter(config RemoteOutputConfig, level zapcore.Level) *remoteExporter {
	// Remote exporter implementation
	// For now, return a no-op writer
	return &remoteExporter{
		writer: zapcore.AddSync(&noOpWriter{}),
		level:  level,
	}
}

func (e *remoteExporter) GetWriter() zapcore.WriteSyncer {
	return e.writer
}

// noOpWriter is a no-op writer for remote exporter placeholder
type noOpWriter struct{}

func (w *noOpWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *noOpWriter) Sync() error {
	return nil
}

// getStacktraceLevel returns the stacktrace level based on configuration
func getStacktraceLevel(config *LoggingConfig, development bool) zapcore.Level {
	if development {
		if config.Development.Stacktrace {
			return zapcore.ErrorLevel
		}
		return zapcore.PanicLevel
	}
	if config.Production.Stacktrace {
		return zapcore.ErrorLevel
	}
	return zapcore.PanicLevel
}

// isDevelopmentMode determines if we're in development mode
func isDevelopmentMode(config *LoggingConfig) bool {
	// Check if any console output with colored format is enabled
	for _, output := range config.Outputs {
		if output.Type == OutputConsole && output.Enabled {
			if output.Format == FormatConsole || config.Format == FormatConsole {
				return true
			}
		}
	}
	return config.Format == FormatConsole
}

// NewNopLogger creates a no-op logger for testing
func NewNopLogger() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
		config: DefaultLoggingConfig(false),
	}
}

// String creates a zap.String field
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates a zap.Int field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Error creates a zap.Error field
func Error(err error) zap.Field {
	return zap.Error(err)
}

// With creates a child logger with fields
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// Trace logs at trace level (if supported)
func (l *Logger) Trace(msg string, fields ...zapcore.Field) {
	// TRACE is implemented as DebugLevel - 1
	// We need to use a custom level
	if ce := l.Logger.Check(zapcore.DebugLevel-1, msg); ce != nil {
		ce.Write(fields...)
	}
}

// GetConfig returns the logging configuration
func (l *Logger) GetConfig() *LoggingConfig {
	return l.config
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
