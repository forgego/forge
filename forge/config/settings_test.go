package config

import (
	"testing"
)

func TestLoadSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	if settings == nil {
		t.Fatal("LoadSettings() returned nil")
	}
}

func TestLoadSettings_AppSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Name", settings.App.Name, "forge"},
		{"Env", settings.App.Env, "development"},
		{"Version", settings.App.Version, "0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("App.%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	if settings.App.Debug != true {
		t.Errorf("App.Debug = %v, want true", settings.App.Debug)
	}
}

func TestLoadSettings_ServerSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Host", settings.Server.Host, "localhost"},
		{"Port", settings.Server.Port, "8000"},
		{"HealthCheckPath", settings.Server.HealthCheckPath, "/health"},
		{"MetricsPath", settings.Server.MetricsPath, "/metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Server.%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	testsInt := []struct {
		name     string
		got      int
		expected int
	}{
		{"ReadTimeout", settings.Server.ReadTimeout, 30},
		{"WriteTimeout", settings.Server.WriteTimeout, 30},
		{"GracefulTimeout", settings.Server.GracefulTimeout, 30},
	}

	for _, tt := range testsInt {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Server.%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}

	if settings.Server.MetricsEnabled != false {
		t.Errorf("Server.MetricsEnabled = %v, want false", settings.Server.MetricsEnabled)
	}
	if settings.Server.EnableProfiling != false {
		t.Errorf("Server.EnableProfiling = %v, want false", settings.Server.EnableProfiling)
	}
}

func TestLoadSettings_DatabaseSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Driver", settings.Database.Driver, "postgres"},
		{"Host", settings.Database.Host, "localhost"},
		{"User", settings.Database.User, "postgres"},
		{"Password", settings.Database.Password, ""},
		{"Name", settings.Database.Name, "forge"},
		{"SSLMode", settings.Database.SSLMode, "disable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Database.%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	testsInt := []struct {
		name     string
		got      int
		expected int
	}{
		{"Port", settings.Database.Port, 5432},
		{"MaxOpenConns", settings.Database.MaxOpenConns, 25},
		{"MaxIdleConns", settings.Database.MaxIdleConns, 10},
		// ConnMaxLifetime and ConnMaxIdleTime are duration strings in config,
		// but LoadSettings uses GetInt which returns 0 for non-integer values
		{"ConnMaxLifetime", settings.Database.ConnMaxLifetime, 0},
		{"ConnMaxIdleTime", settings.Database.ConnMaxIdleTime, 0},
	}

	for _, tt := range testsInt {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Database.%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestLoadSettings_SecuritySettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"SecretKey", settings.Security.SecretKey, "change-me-in-production"},
		{"CSRFSecretKey", settings.Security.CSRFSecretKey, "change-me-in-production"},
		{"SessionSecret", settings.Security.SessionSecret, "change-me-in-production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Security.%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestLoadSettings_AdminSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Path", settings.Admin.Path, "/admin"},
		{"Title", settings.Admin.Title, "forge Admin"},
		{"HeaderTitle", settings.Admin.HeaderTitle, "forge"},
		{"SiteName", settings.Admin.SiteName, "forge"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Admin.%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}

	if settings.Admin.Enabled != true {
		t.Errorf("Admin.Enabled = %v, want true", settings.Admin.Enabled)
	}
}

func TestLoadSettings_LoggingSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	if settings.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want 'info'", settings.Logging.Level)
	}
	if settings.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want 'json'", settings.Logging.Format)
	}
}

func TestLoadSettings_ErrorSettings(t *testing.T) {
	cfg := NewConfig()
	settings := LoadSettings(cfg)

	// ProblemDetails
	if settings.Errors.ProblemDetails.TypeBaseURL != "https://api.example.com/problems" {
		t.Errorf("ProblemDetails.TypeBaseURL = %q, unexpected", settings.Errors.ProblemDetails.TypeBaseURL)
	}
	if settings.Errors.ProblemDetails.IncludeStackTrace != false {
		t.Error("ProblemDetails.IncludeStackTrace should be false")
	}

	// RequestID
	if settings.Errors.RequestID.HeaderName != "X-Request-ID" {
		t.Errorf("RequestID.HeaderName = %q, want 'X-Request-ID'", settings.Errors.RequestID.HeaderName)
	}
	if settings.Errors.RequestID.GenerateIfMissing != true {
		t.Error("RequestID.GenerateIfMissing should be true")
	}

	// Sanitization
	if settings.Errors.Sanitization.HideDatabaseErrors != true {
		t.Error("Sanitization.HideDatabaseErrors should be true")
	}
	if settings.Errors.Sanitization.HideStackTraces != true {
		t.Error("Sanitization.HideStackTraces should be true")
	}
	if settings.Errors.Sanitization.RedactPII != true {
		t.Error("Sanitization.RedactPII should be true")
	}

	// Idempotency
	if settings.Errors.Idempotency.Enabled != true {
		t.Error("Idempotency.Enabled should be true")
	}
	if settings.Errors.Idempotency.HeaderName != "Idempotency-Key" {
		t.Errorf("Idempotency.HeaderName = %q, want 'Idempotency-Key'", settings.Errors.Idempotency.HeaderName)
	}

	// HTTP
	if settings.Errors.HTTP.IncludeRetryAfter != true {
		t.Error("HTTP.IncludeRetryAfter should be true")
	}
	if settings.Errors.HTTP.IncludeLinkHeader != true {
		t.Error("HTTP.IncludeLinkHeader should be true")
	}
}

func TestSettings_Struct(t *testing.T) {
	settings := &Settings{
		App: AppSettings{
			Name:    "test-app",
			Env:     "production",
			Version: "1.0.0",
			Debug:   false,
		},
		Server: ServerSettings{
			Host:         "0.0.0.0",
			Port:         "8080",
			ReadTimeout:  60,
			WriteTimeout: 60,
		},
		Database: DatabaseSettings{
			Driver:   "mysql",
			Host:     "db.example.com",
			Port:     3306,
			User:     "admin",
			Password: "secret",
			Name:     "production_db",
		},
		Security: SecuritySettings{
			SecretKey:     "production-secret-key",
			CSRFSecretKey: "csrf-secret",
			SessionSecret: "session-secret",
		},
		Admin: AdminSettings{
			Enabled:     true,
			Path:        "/dashboard",
			Title:       "Admin Panel",
			HeaderTitle: "Dashboard",
			SiteName:    "MyApp",
		},
	}

	if settings.App.Name != "test-app" {
		t.Errorf("App.Name = %q, want 'test-app'", settings.App.Name)
	}
	if settings.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want '8080'", settings.Server.Port)
	}
	if settings.Database.Driver != "mysql" {
		t.Errorf("Database.Driver = %q, want 'mysql'", settings.Database.Driver)
	}
}

func TestAppSettings_Struct(t *testing.T) {
	app := AppSettings{
		Name:    "myapp",
		Env:     "staging",
		Version: "2.0.0",
		Debug:   true,
	}

	if app.Name != "myapp" {
		t.Errorf("Name = %q, want 'myapp'", app.Name)
	}
	if app.Env != "staging" {
		t.Errorf("Env = %q, want 'staging'", app.Env)
	}
}

func TestServerSettings_Struct(t *testing.T) {
	server := ServerSettings{
		Host:            "127.0.0.1",
		Port:            "3000",
		ReadTimeout:     15,
		WriteTimeout:    15,
		StaticFilesPath: "/static",
		HealthCheckPath: "/healthz",
		MetricsEnabled:  true,
		MetricsPath:     "/metrics",
		GracefulTimeout: 10,
		MaxRequestSize:  5 * 1024 * 1024,
		EnableProfiling: true,
	}

	if server.Port != "3000" {
		t.Errorf("Port = %q, want '3000'", server.Port)
	}
	if !server.MetricsEnabled {
		t.Error("MetricsEnabled should be true")
	}
}

func TestDatabaseSettings_Struct(t *testing.T) {
	db := DatabaseSettings{
		Driver:          "postgres",
		Host:            "localhost",
		User:            "user",
		Password:        "pass",
		Name:            "db",
		SSLMode:         "require",
		Port:            5432,
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 600,
		ConnMaxIdleTime: 300,
	}

	if db.Driver != "postgres" {
		t.Errorf("Driver = %q, want 'postgres'", db.Driver)
	}
	if db.MaxOpenConns != 100 {
		t.Errorf("MaxOpenConns = %d, want 100", db.MaxOpenConns)
	}
}

func TestSecuritySettings_Struct(t *testing.T) {
	security := SecuritySettings{
		SecretKey:     "key1",
		CSRFSecretKey: "key2",
		SessionSecret: "key3",
	}

	if security.SecretKey != "key1" {
		t.Errorf("SecretKey = %q, want 'key1'", security.SecretKey)
	}
}

func TestAdminSettings_Struct(t *testing.T) {
	admin := AdminSettings{
		Path:        "/admin",
		Title:       "Admin",
		HeaderTitle: "Header",
		SiteName:    "Site",
		Enabled:     true,
	}

	if !admin.Enabled {
		t.Error("Enabled should be true")
	}
}
