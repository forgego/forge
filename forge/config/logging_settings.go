package config

// LoggingSettings contains logging configuration
type LoggingSettings struct {
	Level   string
	Format  string
	Outputs []LoggingOutputConfig
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
	Sanitization  SanitizationSettings
	Observability ObservabilitySettings
	Idempotency   IdempotencySettings
	HTTP          HTTPSettings
}

// ProblemDetailsSettings configures RFC 7807 Problem Details
type ProblemDetailsSettings struct {
	TypeBaseURL          string
	IncludeStackTrace    bool
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
	Enabled   bool
	Thresholds AlertThresholds
}

// AlertThresholds defines alert thresholds
type AlertThresholds struct {
	ErrorRatePerMinute    int
	ErrorRatePerEndpoint  int
}

// IdempotencySettings configures idempotency handling
type IdempotencySettings struct {
	Enabled      bool
	HeaderName   string
	CacheTTL     int
	StoreType    string
	MaxNestingDepth int
}

// HTTPSettings configures HTTP semantics
type HTTPSettings struct {
	IncludeRetryAfter        bool
	IncludeLinkHeader        bool
	ProblemJSONContentType   bool
}
