package config

import (
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
		{"security.secret_key", "security.secret_key", "change-me-in-production"},
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
