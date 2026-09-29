package config

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level   string
	Format  string
	Outputs []LoggingOutputConfig
}

// loggingOutputs reads the logging.outputs list. A missing or malformed list
// yields no outputs, which log.NewLoggerFromSettings treats as the console.
// LoadSettings cannot return an error, so a malformed list is reported by
// SettingsWarnings, which the server logs when it starts.
func loggingOutputs(cfg *Config) []LoggingOutputConfig {
	outputs, _ := decodeLoggingOutputs(cfg)
	return outputs
}

func decodeLoggingOutputs(cfg *Config) ([]LoggingOutputConfig, error) {
	if cfg == nil || cfg.Viper == nil || !cfg.Viper.IsSet("logging.outputs") {
		return nil, nil
	}
	var outputs []LoggingOutputConfig
	if err := cfg.Viper.UnmarshalKey("logging.outputs", &outputs); err != nil {
		return nil, err
	}
	return outputs, nil
}

// SettingsWarnings returns one warning per setting that LoadSettings could
// not read and replaced with its default. Like SecretWarnings, loading the
// config does not print them; the server logs them when it starts.
func (c *Config) SettingsWarnings() []string {
	var warnings []string
	if _, err := decodeLoggingOutputs(c); err != nil {
		warnings = append(warnings, "logging.outputs is malformed and was ignored, so logs go to the console: "+err.Error()+
			" (expected a list of outputs such as {type: file, enabled: true, path: logs/app.log})")
	}
	return warnings
}

// LoggingOutputConfig configures a logging output
type LoggingOutputConfig struct {
	Type    string
	Enabled bool
	Level   string
	Format  string
	Path    string // For file output
}

// ErrorSettings contains error handling configuration
type ErrorSettings struct {
	ProblemDetails ProblemDetailsSettings
	RequestID      RequestIDSettings
	Sanitization   SanitizationSettings
	Observability  ObservabilitySettings
	Idempotency    IdempotencySettings
	HTTP           HTTPSettings
}

// ProblemDetailsSettings configures RFC 7807 Problem Details
type ProblemDetailsSettings struct {
	TypeBaseURL            string
	IncludeStackTrace      bool
	IncludeInternalDetails bool
}

// RequestIDSettings configures request ID handling
type RequestIDSettings struct {
	HeaderName        string
	GenerateIfMissing bool
	IncludeInResponse bool
}

// SanitizationSettings configures error sanitization
type SanitizationSettings struct {
	HideDatabaseErrors bool
	HideStackTraces    bool
	RedactPII          bool
	PIIPatterns        []string
}

// ObservabilitySettings configures observability features
type ObservabilitySettings struct {
	MetricsEnabled bool
	TracingEnabled bool
	ErrorTracking  ErrorTrackingSettings
	Alerts         AlertsSettings
}

// ErrorTrackingSettings configures error tracking services
type ErrorTrackingSettings struct {
	Enabled bool
	Service string
	DSN     string
}

// AlertsSettings configures alerting
type AlertsSettings struct {
	Enabled    bool
	Thresholds AlertThresholds
}

// AlertThresholds defines alert thresholds
type AlertThresholds struct {
	ErrorRatePerMinute   int
	ErrorRatePerEndpoint int
}

// IdempotencySettings configures idempotency handling
type IdempotencySettings struct {
	Enabled         bool
	HeaderName      string
	CacheTTL        int
	StoreType       string
	MaxNestingDepth int
}

// HTTPSettings configures HTTP semantics
type HTTPSettings struct {
	IncludeRetryAfter      bool
	IncludeLinkHeader      bool
	ProblemJSONContentType bool
}
