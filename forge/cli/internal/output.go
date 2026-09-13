package internal

import (
	"fmt"
	"os"

	"github.com/forgego/forge/log"
)

var defaultLogger *log.Logger

// init initializes the default logger for CLI output
func init() {
	// Create a simple console logger for CLI
	config := log.DefaultLoggingConfig(true) // Development mode for CLI
	config.Outputs = []log.OutputConfig{
		{
			Type:    log.OutputConsole,
			Enabled: true,
			Level:   log.LevelInfo,
			Format:  log.FormatConsole,
		},
	}

	var err error
	defaultLogger, err = log.NewLoggerFromConfig(config)
	if err != nil {
		// Fallback to basic logger if config fails
		defaultLogger, _ = log.NewLogger(true)
	}
}

// PrintSuccess prints a success message with a checkmark
func PrintSuccess(message string) {
	if defaultLogger != nil {
		defaultLogger.Info("✓ " + message)
	} else {
		fmt.Fprintf(os.Stdout, "✓ %s\n", message)
	}
}

// PrintError prints an error message
func PrintError(message string) {
	if defaultLogger != nil {
		defaultLogger.Error("❌ " + message)
	} else {
		fmt.Fprintf(os.Stderr, "❌ %s\n", message)
	}
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	if defaultLogger != nil {
		defaultLogger.Warn("⚠️  " + message)
	} else {
		fmt.Fprintf(os.Stderr, "⚠️  %s\n", message)
	}
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	if defaultLogger != nil {
		defaultLogger.Info("ℹ️  " + message)
	} else {
		fmt.Fprintf(os.Stdout, "ℹ️  %s\n", message)
	}
}

// SetLogger sets a custom logger for CLI output
func SetLogger(logger *log.Logger) {
	defaultLogger = logger
}

// GetLogger returns the current CLI logger
func GetLogger() *log.Logger {
	return defaultLogger
}
