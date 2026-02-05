package log

import (
	"fmt"
	"time"
)

// Builder provides a fluent interface for building loggers
type Builder struct {
	config *LoggingConfig
}

// NewBuilder creates a new logger builder
func NewBuilder() *Builder {
	return &Builder{
		config: DefaultLoggingConfig(false),
	}
}

// Development sets development mode
func (b *Builder) Development() *Builder {
	b.config = DefaultLoggingConfig(true)
	return b
}

// Production sets production mode
func (b *Builder) Production() *Builder {
	b.config = DefaultLoggingConfig(false)
	return b
}

// Level sets the default log level
func (b *Builder) Level(level Level) *Builder {
	b.config.Level = level
	return b
}

// Format sets the default log format
func (b *Builder) Format(format Format) *Builder {
	b.config.Format = format
	return b
}

// AddConsoleOutput adds a console output
func (b *Builder) AddConsoleOutput(level Level) *Builder {
	b.config.Outputs = append(b.config.Outputs, OutputConfig{
		Type:    OutputConsole,
		Enabled: true,
		Level:   level,
		Format:  FormatConsole,
	})
	return b
}

// AddFileOutput adds a file output with rotation
func (b *Builder) AddFileOutput(path string, level Level, maxSize, maxAge, maxBackups int, compress bool) *Builder {
	b.config.Outputs = append(b.config.Outputs, OutputConfig{
		Type:    OutputFile,
		Enabled: true,
		Level:   level,
		Format:  FormatJSON,
		File: FileOutputConfig{
			Path: path,
			Rotation: RotationConfig{
				MaxSize:    maxSize,
				MaxAge:     maxAge,
				MaxBackups: maxBackups,
				Compress:   compress,
			},
		},
	})
	return b
}

// AddRemoteOutput adds a remote output
func (b *Builder) AddRemoteOutput(url string, level Level, timeout int) *Builder {
	b.config.Outputs = append(b.config.Outputs, OutputConfig{
		Type:    OutputRemote,
		Enabled: true,
		Level:   level,
		Format:  FormatJSON,
		Remote: RemoteOutputConfig{
			URL:     url,
			Format:  FormatJSON,
			Timeout: time.Duration(timeout) * time.Second,
		},
	})
	return b
}

// Colored enables colored output (dev mode)
func (b *Builder) Colored(enabled bool) *Builder {
	b.config.Development.Colored = enabled
	return b
}

// Caller enables caller information
func (b *Builder) Caller(enabled bool) *Builder {
	b.config.Development.Caller = enabled
	b.config.Production.Caller = enabled
	return b
}

// Stacktrace enables stack traces
func (b *Builder) Stacktrace(enabled bool) *Builder {
	b.config.Development.Stacktrace = enabled
	b.config.Production.Stacktrace = enabled
	return b
}

// OneLine enables one-line format (dev mode)
func (b *Builder) OneLine(enabled bool) *Builder {
	b.config.Development.OneLine = enabled
	return b
}

// Sampling enables log sampling (prod mode)
func (b *Builder) Sampling(initial, thereafter int) *Builder {
	b.config.Production.Sampling.Enabled = true
	b.config.Production.Sampling.Initial = initial
	b.config.Production.Sampling.Thereafter = thereafter
	return b
}

// Build creates the logger from the builder configuration
func (b *Builder) Build() (*Logger, error) {
	return NewLoggerFromConfig(b.config)
}

// MustBuild creates the logger and panics on error
func (b *Builder) MustBuild() *Logger {
	logger, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}
	return logger
}

// QuickLogger creates a quick logger for development
func QuickLogger() *Logger {
	builder := NewBuilder().Development()
	builder.AddConsoleOutput(LevelDebug)
	logger, err := builder.Build()
	if err != nil {
		// Fallback to basic logger
		logger, _ = NewLogger(true)
	}
	return logger
}

// ProductionLogger creates a production-ready logger
func ProductionLogger(logPath string) (*Logger, error) {
	builder := NewBuilder().Production()
	builder.AddConsoleOutput(LevelInfo)
	if logPath != "" {
		builder.AddFileOutput(logPath, LevelInfo, 100, 30, 10, true)
	}
	builder.Sampling(100, 100)
	return builder.Build()
}
