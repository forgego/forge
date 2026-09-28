// Package log provides logging functionality
package log

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	forgeerrors "github.com/forgego/forge/errors"
	logexporters "github.com/forgego/forge/log/exporters"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with enhanced functionality
type Logger struct {
	*zap.Logger
	config    *LoggingConfig
	resources *loggerResources
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
		return newLoggerFromConfig(nil)
	}
	return newLoggerFromConfig(config, config.Hooks...)
}

func newLoggerFromConfig(config *LoggingConfig, hooks ...Hook) (*Logger, error) {
	if config == nil {
		config = DefaultLoggingConfig(false)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logging configuration: %w", err)
	}

	// Build cores for each output
	cores := make([]zapcore.Core, 0)
	closers := make([]io.Closer, 0)
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
			var closer io.Closer
			core, closer = createFileCore(encoder, level, output.File)
			closers = append(closers, closer)
		case OutputRemote:
			return nil, forgeerrors.NewNotImplementedError("log remote output")
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

	if len(hooks) > 0 {
		registry := NewHookRegistry()
		for _, hook := range hooks {
			registry.AddHook(hook)
		}
		combinedCore = NewHookCore(combinedCore, registry)
	}

	// Create logger
	zapLogger := zap.New(
		combinedCore,
		zap.AddCaller(),
		zap.AddStacktrace(getStacktraceLevel(config, development)),
	)

	return &Logger{
		Logger:    zapLogger,
		config:    config,
		resources: &loggerResources{closers: closers},
	}, nil
}

// createConsoleCore creates a console core
func createConsoleCore(encoder zapcore.Encoder, level zapcore.Level) zapcore.Core {
	writer := zapcore.AddSync(os.Stderr)
	return zapcore.NewCore(encoder, writer, level)
}

// createFileCore creates a file core with rotation
func createFileCore(encoder zapcore.Encoder, level zapcore.Level, fileConfig FileOutputConfig) (zapcore.Core, io.Closer) {
	fileExp := logexporters.NewFileExporter(logexporters.FileConfig{
		Path:       fileConfig.Path,
		MaxSize:    fileConfig.Rotation.MaxSize,
		MaxAge:     fileConfig.Rotation.MaxAge,
		MaxBackups: fileConfig.Rotation.MaxBackups,
		Compress:   fileConfig.Rotation.Compress,
	}, level)
	writer := fileExp.GetWriter()
	return zapcore.NewCore(encoder, writer, level), fileExp
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
		Logger:    zap.NewNop(),
		config:    DefaultLoggingConfig(false),
		resources: &loggerResources{},
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
		Logger:    l.Logger.With(fields...),
		config:    l.config,
		resources: l.resources,
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

type loggerResources struct {
	closers []io.Closer
	once    sync.Once
	err     error
}

// Close flushes and closes outputs opened by the logger.
func (l *Logger) Close() error {
	if l.resources == nil {
		return ignoreSyncError(l.Logger.Sync())
	}
	l.resources.once.Do(func() {
		l.resources.err = ignoreSyncError(l.Logger.Sync())
		for _, closer := range l.resources.closers {
			if err := closer.Close(); err != nil {
				l.resources.err = multierr.Append(l.resources.err, fmt.Errorf("close log output: %w", err))
			}
		}
	})
	return l.resources.err
}

func ignoreSyncError(err error) error {
	if err == nil {
		return nil
	}
	var rest []error
	for _, part := range multierr.Errors(err) {
		if errors.Is(part, syscall.EINVAL) {
			continue
		}
		rest = append(rest, part)
	}
	return multierr.Combine(rest...)
}
