package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if cfg.Viper == nil {
		t.Error("Config.Viper is nil")
	}
}

func TestNewConfig_GeneratesSecrets(t *testing.T) {
	keys := []string{
		"security.secret_key",
		"security.csrf_secret_key",
		"security.session_secret",
	}

	first := NewConfig()
	seen := map[string]bool{}
	for _, key := range keys {
		v := first.GetString(key, "")
		if len(v) != 64 {
			t.Errorf("GetString(%q) = %q, want 64-char generated hex", key, v)
		}
		if v == "change-me-in-production" {
			t.Errorf("GetString(%q) still ships the placeholder secret", key)
		}
		if seen[v] {
			t.Errorf("duplicate generated secret for %q", key)
		}
		seen[v] = true
	}

	// A fresh instance must not reuse the same secrets.
	second := NewConfig()
	for _, key := range keys {
		if second.GetString(key, "") == first.GetString(key, "") {
			t.Errorf("secret for %q was reused across instances", key)
		}
	}
}

func TestNewConfig_OverridesPlaceholderSecrets(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", "change-me-in-production-with-random-string")
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", "change-me-in-production")
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", "secret")

	cfg := NewConfig()
	for _, key := range []string{
		"security.secret_key",
		"security.csrf_secret_key",
		"security.session_secret",
	} {
		v := cfg.GetString(key, "")
		if len(v) != 64 {
			t.Errorf("GetString(%q) = %q, expected 64-char generated hex secret", key, v)
		}
		if isPlaceholderSecret(v) {
			t.Errorf("GetString(%q) remained placeholder secret %q", key, v)
		}
	}
}

func TestConfigGeneratedSecretsReportsOnlyGeneratedKeyNamesSorted(t *testing.T) {
	t.Setenv("FORGE_SECURITY_SECRET_KEY", "this-is-an-explicit-strong-secret-key-value")
	t.Setenv("FORGE_SECURITY_CSRF_SECRET_KEY", "change-me-in-production")
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", "")

	cfg := NewConfig()
	want := []string{"security.csrf_secret_key", "security.session_secret"}
	if got := cfg.GeneratedSecrets(); !reflect.DeepEqual(got, want) {
		t.Fatalf("GeneratedSecrets() = %v, want %v", got, want)
	}

	got := cfg.GeneratedSecrets()
	got[0] = "mutated"
	if next := cfg.GeneratedSecrets(); !reflect.DeepEqual(next, want) {
		t.Fatalf("GeneratedSecrets returned mutable internal state: %v", next)
	}
}

func TestNewConfig_Defaults(t *testing.T) {
	cfg := NewConfig()

	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"app.name", "app.name", "forge"},
		{"app.env", "app.env", "development"},
		{"app.debug", "app.debug", true},
		{"app.version", "app.version", "0.1.0"},
		{"server.host", "server.host", "localhost"},
		{"server.port", "server.port", "8000"},
		{"server.read_timeout", "server.read_timeout", 30},
		{"server.write_timeout", "server.write_timeout", 30},
		{"database.driver", "database.driver", "postgres"},
		{"database.host", "database.host", "localhost"},
		{"database.port", "database.port", 5432},
		{"database.user", "database.user", "postgres"},
		{"database.password", "database.password", ""},
		{"database.name", "database.name", "forge"},
		{"database.sslmode", "database.sslmode", "disable"},
		{"database.max_open_conns", "database.max_open_conns", 25},
		{"database.max_idle_conns", "database.max_idle_conns", 10},
		{"admin.enabled", "admin.enabled", true},
		{"admin.path", "admin.path", "/admin"},
		{"admin.title", "admin.title", "forge Admin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.expected.(type) {
			case string:
				if result := cfg.GetString(tt.key, ""); result != v {
					t.Errorf("GetString(%q) = %q, want %q", tt.key, result, v)
				}
			case int:
				if result := cfg.GetInt(tt.key, 0); result != v {
					t.Errorf("GetInt(%q) = %d, want %d", tt.key, result, v)
				}
			case bool:
				if result := cfg.GetBool(tt.key, false); result != v {
					t.Errorf("GetBool(%q) = %v, want %v", tt.key, result, v)
				}
			}
		})
	}
}

func TestConfig_GetString(t *testing.T) {
	cfg := NewConfig()

	// Test existing key
	result := cfg.GetString("app.name", "default")
	if result != "forge" {
		t.Errorf("GetString('app.name') = %q, want 'forge'", result)
	}

	// Test non-existing key with default
	result = cfg.GetString("nonexistent.key", "default_value")
	if result != "default_value" {
		t.Errorf("GetString('nonexistent.key') = %q, want 'default_value'", result)
	}
}

func TestConfig_GetInt(t *testing.T) {
	cfg := NewConfig()

	// Test existing key
	result := cfg.GetInt("server.read_timeout", 0)
	if result != 30 {
		t.Errorf("GetInt('server.read_timeout') = %d, want 30", result)
	}

	// Test non-existing key with default
	result = cfg.GetInt("nonexistent.key", 999)
	if result != 999 {
		t.Errorf("GetInt('nonexistent.key') = %d, want 999", result)
	}
}

func TestConfig_GetBool(t *testing.T) {
	cfg := NewConfig()

	// Test existing key
	result := cfg.GetBool("app.debug", false)
	if result != true {
		t.Errorf("GetBool('app.debug') = %v, want true", result)
	}

	// Test non-existing key with default
	result = cfg.GetBool("nonexistent.key", true)
	if result != true {
		t.Errorf("GetBool('nonexistent.key') = %v, want true", result)
	}
}

func TestConfig_GetInt64(t *testing.T) {
	cfg := NewConfig()

	// Test non-existing key with default
	result := cfg.GetInt64("nonexistent.key", 12345)
	if result != 12345 {
		t.Errorf("GetInt64('nonexistent.key') = %d, want 12345", result)
	}
}

func TestConfig_GetDuration(t *testing.T) {
	cfg := NewConfig()

	// Test non-existing key with default
	defaultDuration := 5 * time.Minute
	result := cfg.GetDuration("nonexistent.key", defaultDuration)
	if result != defaultDuration {
		t.Errorf("GetDuration('nonexistent.key') = %v, want %v", result, defaultDuration)
	}
}

func TestConfig_GetDriver(t *testing.T) {
	cfg := NewConfig()
	result := cfg.GetDriver()
	if result != "postgres" {
		t.Errorf("GetDriver() = %q, want 'postgres'", result)
	}
}

func TestConfig_GetPoolConfig(t *testing.T) {
	cfg := NewConfig()
	poolConfig := cfg.GetPoolConfig()

	if poolConfig.MaxOpenConns != 25 {
		t.Errorf("PoolConfig.MaxOpenConns = %d, want 25", poolConfig.MaxOpenConns)
	}
	if poolConfig.MaxIdleConns != 10 {
		t.Errorf("PoolConfig.MaxIdleConns = %d, want 10", poolConfig.MaxIdleConns)
	}
	if poolConfig.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("PoolConfig.ConnMaxLifetime = %v, want 5m", poolConfig.ConnMaxLifetime)
	}
	if poolConfig.ConnMaxIdleTime != 2*time.Minute {
		t.Errorf("PoolConfig.ConnMaxIdleTime = %v, want 2m", poolConfig.ConnMaxIdleTime)
	}
}

func TestConfig_Set(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("custom.key", "custom_value")
	result := cfg.GetString("custom.key", "")
	if result != "custom_value" {
		t.Errorf("GetString('custom.key') = %q, want 'custom_value'", result)
	}
}

func TestConfig_IsSet(t *testing.T) {
	cfg := NewConfig()

	// Default values are set
	if !cfg.IsSet("app.name") {
		t.Error("IsSet('app.name') should be true for default value")
	}

	// Non-existing key
	if cfg.IsSet("nonexistent.key") {
		t.Error("IsSet('nonexistent.key') should be false")
	}
}

func TestPoolConfig_Struct(t *testing.T) {
	poolConfig := PoolConfig{
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	if poolConfig.MaxOpenConns != 50 {
		t.Errorf("MaxOpenConns = %d, want 50", poolConfig.MaxOpenConns)
	}
	if poolConfig.MaxIdleConns != 25 {
		t.Errorf("MaxIdleConns = %d, want 25", poolConfig.MaxIdleConns)
	}
	if poolConfig.ConnMaxLifetime != 10*time.Minute {
		t.Errorf("ConnMaxLifetime = %v, want 10m", poolConfig.ConnMaxLifetime)
	}
	if poolConfig.ConnMaxIdleTime != 5*time.Minute {
		t.Errorf("ConnMaxIdleTime = %v, want 5m", poolConfig.ConnMaxIdleTime)
	}
}

func TestNewConfig_EnvironmentOverrides(t *testing.T) {
	t.Setenv("FORGE_SERVER_HOST", "0.0.0.0")
	t.Setenv("FORGE_SERVER_PORT", "8020")
	t.Setenv("FORGE_DATABASE_DRIVER", "sqlite")
	secret := strings.Repeat("a", 64)
	t.Setenv("FORGE_SECURITY_SESSION_SECRET", secret)
	cfg := NewConfig()
	if got := cfg.GetString("server.host", ""); got != "0.0.0.0" {
		t.Errorf("host = %q", got)
	}
	if got := cfg.GetInt("server.port", 0); got != 8020 {
		t.Errorf("port = %d", got)
	}
	if got := cfg.GetDriver(); got != "sqlite" {
		t.Errorf("driver = %q", got)
	}
	if got := cfg.GetString("security.session_secret", ""); got != secret {
		t.Error("configured session secret was replaced")
	}
}
