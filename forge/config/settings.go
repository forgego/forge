package config

// Settings represents framework settings structure
type Settings struct {
	Admin    AdminSettings
	App      AppSettings
	Security SecuritySettings
	Server   ServerSettings
	Database DatabaseSettings
	Logging  LoggingSettings
	Errors   ErrorSettings
}

// AppSettings contains application-level settings
type AppSettings struct {
	Name    string
	Env     string
	Version string
	Debug   bool
}

// ServerSettings contains HTTP server settings
type ServerSettings struct {
	Host            string
	Port            string
	ReadTimeout     int    // seconds
	WriteTimeout    int    // seconds
	StaticFilesPath string // path to static files directory
	HealthCheckPath string // path for health check endpoint
	MetricsEnabled  bool   // enable metrics endpoint
	MetricsPath     string // path for metrics endpoint
	GracefulTimeout int    // graceful shutdown timeout in seconds
	MaxRequestSize  int64  // maximum request body size in bytes
	EnableProfiling bool   // enable profiling (dev mode only)
}

// DatabaseSettings contains database connection settings
type DatabaseSettings struct {
	Driver          string
	Host            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	Port            int
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
	ConnMaxIdleTime int
}

// SecuritySettings contains security-related settings
type SecuritySettings struct {
	SecretKey     string
	CSRFSecretKey string
	SessionSecret string
}

// AdminSettings contains admin interface settings
type AdminSettings struct {
	Path        string
	Title       string
	HeaderTitle string
	SiteName    string
	Enabled     bool
}

// LoadSettings loads settings from config
func LoadSettings(cfg *Config) *Settings {
	return &Settings{
		App: AppSettings{
			Name:    cfg.GetString("app.name", "forge"),
			Env:     cfg.GetString("app.env", "development"),
			Debug:   cfg.GetBool("app.debug", true),
			Version: cfg.GetString("app.version", "0.1.0"),
		},
		Server: ServerSettings{
			Host:            cfg.GetString("server.host", "localhost"),
			Port:            cfg.GetString("server.port", "8000"),
			ReadTimeout:     cfg.GetInt("server.read_timeout", 30),
			WriteTimeout:    cfg.GetInt("server.write_timeout", 30),
			StaticFilesPath: cfg.GetString("server.static_files_path", ""),
			HealthCheckPath: cfg.GetString("server.health_check_path", "/health"),
			MetricsEnabled:  cfg.GetBool("server.metrics_enabled", false),
			MetricsPath:     cfg.GetString("server.metrics_path", "/metrics"),
			GracefulTimeout: cfg.GetInt("server.graceful_timeout", 30),
			MaxRequestSize:  cfg.GetInt64("server.max_request_size", 10*1024*1024), // 10MB default
			EnableProfiling: cfg.GetBool("server.enable_profiling", false),
		},
		Database: DatabaseSettings{
			Driver:          cfg.GetString("database.driver", "postgres"),
			Host:            cfg.GetString("database.host", "localhost"),
			Port:            cfg.GetInt("database.port", 5432),
			User:            cfg.GetString("database.user", "postgres"),
			Password:        cfg.GetString("database.password", ""),
			Name:            cfg.GetString("database.name", "forge"),
			SSLMode:         cfg.GetString("database.sslmode", "disable"),
			MaxOpenConns:    cfg.GetInt("database.max_open_conns", 25),
			MaxIdleConns:    cfg.GetInt("database.max_idle_conns", 5),
			ConnMaxLifetime: cfg.GetInt("database.conn_max_lifetime", 300),
			ConnMaxIdleTime: cfg.GetInt("database.conn_max_idle_time", 600),
		},
		Security: SecuritySettings{
			SecretKey:     cfg.GetString("security.secret_key", "change-me-in-production"),
			CSRFSecretKey: cfg.GetString("security.csrf_secret_key", "change-me-in-production"),
			SessionSecret: cfg.GetString("security.session_secret", "change-me-in-production"),
		},
		Admin: AdminSettings{
			Enabled:     cfg.GetBool("admin.enabled", true),
			Path:        cfg.GetString("admin.path", "/admin"),
			Title:       cfg.GetString("admin.title", "forge Admin"),
			HeaderTitle: cfg.GetString("admin.header_title", "forge"),
			SiteName:    cfg.GetString("admin.site_name", "forge"),
		},
		Logging: LoggingSettings{
			Level:  cfg.GetString("logging.level", "info"),
			Format: cfg.GetString("logging.format", "json"),
			// Outputs will be loaded from config if needed
		},
		Errors: ErrorSettings{
			ProblemDetails: ProblemDetailsSettings{
				TypeBaseURL:           cfg.GetString("errors.problem_details.type_base_url", "https://api.example.com/problems"),
				IncludeStackTrace:     cfg.GetBool("errors.problem_details.include_stack_trace", false),
				IncludeInternalDetails: cfg.GetBool("errors.problem_details.include_internal_details", false),
			},
			RequestID: RequestIDSettings{
				HeaderName:        cfg.GetString("errors.request_id.header_name", "X-Request-ID"),
				GenerateIfMissing: cfg.GetBool("errors.request_id.generate_if_missing", true),
				IncludeInResponse: cfg.GetBool("errors.request_id.include_in_response", true),
			},
			Sanitization: SanitizationSettings{
				HideDatabaseErrors: cfg.GetBool("errors.sanitization.hide_database_errors", true),
				HideStackTraces:    cfg.GetBool("errors.sanitization.hide_stack_traces", true),
				RedactPII:          cfg.GetBool("errors.sanitization.redact_pii", true),
			},
			Idempotency: IdempotencySettings{
				Enabled:        cfg.GetBool("errors.idempotency.enabled", true),
				HeaderName:     cfg.GetString("errors.idempotency.header_name", "Idempotency-Key"),
				CacheTTL:       cfg.GetInt("errors.idempotency.cache_ttl", 3600),
				StoreType:      cfg.GetString("errors.idempotency.store_type", "memory"),
				MaxNestingDepth: cfg.GetInt("errors.idempotency.max_nesting_depth", 10),
			},
			HTTP: HTTPSettings{
				IncludeRetryAfter:      cfg.GetBool("errors.http.include_retry_after", true),
				IncludeLinkHeader:      cfg.GetBool("errors.http.include_link_header", true),
				ProblemJSONContentType: cfg.GetBool("errors.http.problem_json_content_type", true),
			},
		},
	}
}
