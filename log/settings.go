package log

import (
	"strings"

	"github.com/forgego/forge/config"
)

// ConfigFromSettings converts the logging.* settings loaded by
// config.LoadSettings into a LoggingConfig. The level and format apply to
// every output that does not set its own. With no logging.outputs configured,
// logs go to the console. The console format selects the development encoder
// (colored, one line, caller); json and text select the production ones.
func ConfigFromSettings(settings config.LoggingSettings) *LoggingConfig {
	format := Format(strings.ToLower(strings.TrimSpace(settings.Format)))
	cfg := DefaultLoggingConfig(format == FormatConsole)
	if level := Level(strings.ToLower(strings.TrimSpace(settings.Level))); level != "" {
		cfg.Level = level
	}
	if format != "" {
		cfg.Format = format
	}
	cfg.Outputs = nil
	for _, output := range settings.Outputs {
		out := OutputConfig{
			Type:    OutputType(strings.ToLower(strings.TrimSpace(output.Type))),
			Enabled: output.Enabled,
			Level:   Level(strings.ToLower(strings.TrimSpace(output.Level))),
			Format:  Format(strings.ToLower(strings.TrimSpace(output.Format))),
		}
		if out.Level == "" {
			out.Level = cfg.Level
		}
		if out.Type == OutputFile {
			out.File = FileOutputConfig{
				Path: output.Path,
				Rotation: RotationConfig{
					MaxSize:    100, // MB
					MaxAge:     30,  // days
					MaxBackups: 10,
					Compress:   true,
				},
			}
		}
		cfg.Outputs = append(cfg.Outputs, out)
	}
	if len(cfg.Outputs) == 0 {
		cfg.Outputs = []OutputConfig{{Type: OutputConsole, Enabled: true, Level: cfg.Level}}
	}
	return cfg
}

// NewLoggerFromSettings builds a logger from the logging.* settings, the way
// generated projects create theirs. An invalid level, format or output is an
// error rather than being silently replaced by a default.
func NewLoggerFromSettings(settings config.LoggingSettings) (*Logger, error) {
	return NewLoggerFromConfig(ConfigFromSettings(settings))
}
