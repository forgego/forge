package log

import (
	"fmt"
	"time"
)

// Level represents a log level
type Level string

const (
	LevelTrace Level = "trace"
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelFatal Level = "fatal"
)

// Format represents a log format
type Format string

const (
	FormatJSON    Format = "json"
	FormatText    Format = "text"
	FormatConsole Format = "console" // Colored console output
)

// OutputType represents an output destination type
type OutputType string

const (
	OutputConsole OutputType = "console"
	OutputFile    OutputType = "file"
	OutputRemote  OutputType = "remote"
)

// LoggingConfig represents the complete logging configuration
type LoggingConfig struct {
	// Level is the default log level
	Level Level `yaml:"level" json:"level"`
	// Format is the default log format
	Format Format `yaml:"format" json:"format"`
	// Outputs is a list of output destinations
	Outputs []OutputConfig `yaml:"outputs" json:"outputs"`
	// Development settings
	Development DevelopmentConfig `yaml:"development" json:"development"`
	// Production settings
	Production ProductionConfig `yaml:"production" json:"production"`
}

// OutputConfig configures a log output destination
type OutputConfig struct {
	// Type is the output type (console, file, remote)
	Type OutputType `yaml:"type" json:"type"`
	// Enabled enables or disables this output
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Level is the minimum log level for this output
	Level Level `yaml:"level" json:"level"`
	// Format is the format for this output
	Format Format `yaml:"format,omitempty" json:"format,omitempty"`
	// File-specific settings
	File FileOutputConfig `yaml:"file,omitempty" json:"file,omitempty"`
	// Remote-specific settings
	Remote RemoteOutputConfig `yaml:"remote,omitempty" json:"remote,omitempty"`
}

// FileOutputConfig configures file output
type FileOutputConfig struct {
	// Path is the file path
	Path string `yaml:"path" json:"path"`
	// Rotation configures log rotation
	Rotation RotationConfig `yaml:"rotation" json:"rotation"`
}

// RotationConfig configures log file rotation
type RotationConfig struct {
	// MaxSize is the maximum file size in MB before rotation
	MaxSize int `yaml:"max_size" json:"max_size"`
	// MaxAge is the maximum age in days before deletion
	MaxAge int `yaml:"max_age" json:"max_age"`
	// MaxBackups is the maximum number of backup files to keep
	MaxBackups int `yaml:"max_backups" json:"max_backups"`
	// Compress enables compression of rotated files
	Compress bool `yaml:"compress" json:"compress"`
}

// RemoteOutputConfig configures remote output
type RemoteOutputConfig struct {
	// URL is the remote logging service URL
	URL string `yaml:"url" json:"url"`
	// Format is the format for remote output
	Format Format `yaml:"format" json:"format"`
	// Timeout is the request timeout
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

// DevelopmentConfig configures development mode settings
type DevelopmentConfig struct {
	// Colored enables colored output
	Colored bool `yaml:"colored" json:"colored"`
	// Caller enables caller information (file:line)
	Caller bool `yaml:"caller" json:"caller"`
	// Stacktrace enables stack traces for errors
	Stacktrace bool `yaml:"stacktrace" json:"stacktrace"`
	// OneLine enables one-line log format
	OneLine bool `yaml:"one_line" json:"one_line"`
}

// ProductionConfig configures production mode settings
type ProductionConfig struct {
	// Caller enables caller information
	Caller bool `yaml:"caller" json:"caller"`
	// Stacktrace enables stack traces for errors
	Stacktrace bool `yaml:"stacktrace" json:"stacktrace"`
	// Sampling configures log sampling
	Sampling SamplingConfig `yaml:"sampling" json:"sampling"`
}

// SamplingConfig configures log sampling
type SamplingConfig struct {
	// Enabled enables log sampling
	Enabled bool `yaml:"enabled" json:"enabled"`
	// Initial is the initial number of logs to sample
	Initial int `yaml:"initial" json:"initial"`
	// Thereafter is the sampling rate after initial logs
	Thereafter int `yaml:"thereafter" json:"thereafter"`
}

// DefaultLoggingConfig returns a default logging configuration
func DefaultLoggingConfig(development bool) *LoggingConfig {
	config := &LoggingConfig{
		Level:  LevelInfo,
		Format: FormatJSON,
		Outputs: []OutputConfig{
			{
				Type:    OutputConsole,
				Enabled: true,
				Level:   LevelDebug,
				Format:  FormatConsole,
			},
		},
		Development: DevelopmentConfig{
			Colored:    true,
			Caller:     true,
			Stacktrace: true,
			OneLine:    true,
		},
		Production: ProductionConfig{
			Caller:     false,
			Stacktrace: false,
			Sampling: SamplingConfig{
				Enabled:    true,
				Initial:    100,
				Thereafter: 100,
			},
		},
	}

	if development {
		config.Level = LevelDebug
		config.Format = FormatConsole
	} else {
		config.Level = LevelInfo
		config.Format = FormatJSON
		// Add file output in production
		config.Outputs = append(config.Outputs, OutputConfig{
			Type:    OutputFile,
			Enabled: true,
			Level:   LevelInfo,
			Format:  FormatJSON,
			File: FileOutputConfig{
				Path: "logs/app.log",
				Rotation: RotationConfig{
					MaxSize:    100, // MB
					MaxAge:     30,  // days
					MaxBackups: 10,
					Compress:   true,
				},
			},
		})
	}

	return config
}

// Validate validates the logging configuration
func (c *LoggingConfig) Validate() error {
	if c.Level == "" {
		return fmt.Errorf("log level is required")
	}

	if !isValidLevel(c.Level) {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	if c.Format == "" {
		return fmt.Errorf("log format is required")
	}

	if !isValidFormat(c.Format) {
		return fmt.Errorf("invalid log format: %s", c.Format)
	}

	if len(c.Outputs) == 0 {
		return fmt.Errorf("at least one output must be configured")
	}

	for i, output := range c.Outputs {
		if err := output.Validate(); err != nil {
			return fmt.Errorf("output %d: %w", i, err)
		}
	}

	return nil
}

// Validate validates an output configuration
func (o *OutputConfig) Validate() error {
	if o.Type == "" {
		return fmt.Errorf("output type is required")
	}

	if !isValidOutputType(o.Type) {
		return fmt.Errorf("invalid output type: %s", o.Type)
	}

	if o.Level == "" {
		o.Level = LevelInfo
	}

	if !isValidLevel(o.Level) {
		return fmt.Errorf("invalid output level: %s", o.Level)
	}

	if o.Type == OutputFile {
		if o.File.Path == "" {
			return fmt.Errorf("file path is required for file output")
		}
	}

	if o.Type == OutputRemote {
		if o.Remote.URL == "" {
			return fmt.Errorf("URL is required for remote output")
		}
	}

	return nil
}

// isValidLevel checks if a level is valid
func isValidLevel(level Level) bool {
	switch level {
	case LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
		return true
	default:
		return false
	}
}

// isValidFormat checks if a format is valid
func isValidFormat(format Format) bool {
	switch format {
	case FormatJSON, FormatText, FormatConsole:
		return true
	default:
		return false
	}
}

// isValidOutputType checks if an output type is valid
func isValidOutputType(outputType OutputType) bool {
	switch outputType {
	case OutputConsole, OutputFile, OutputRemote:
		return true
	default:
		return false
	}
}
