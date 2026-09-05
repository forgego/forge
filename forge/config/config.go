package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config wraps viper for configuration management
type Config struct {
	*viper.Viper
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")

	// Set defaults
	v.SetDefault("app.name", "forge")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8000")
	v.SetDefault("server.read_timeout", 30)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "forge")
	v.SetDefault("database.sslmode", "disable")
	// Database connection pool defaults
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("database.conn_max_idle_time", "2m")
	v.SetDefault("security.secret_key", "change-me-in-production")
	v.SetDefault("security.csrf_secret_key", "change-me-in-production")
	v.SetDefault("security.session_secret", "change-me-in-production")
	v.SetDefault("security.csrf_exempt_paths", []string{})
	v.SetDefault("admin.enabled", true)
	v.SetDefault("admin.path", "/admin")
	v.SetDefault("admin.title", "forge Admin")
	v.SetDefault("admin.header_title", "forge")
	v.SetDefault("admin.site_name", "forge")

	// Read config file (ignore errors if file doesn't exist)
	_ = v.ReadInConfig()

	return &Config{Viper: v}
}

// GetString gets a string value with a default
func (c *Config) GetString(key string, defaultValue string) string {
	if c.Viper.IsSet(key) {
		return c.Viper.GetString(key)
	}
	return defaultValue
}

// GetInt gets an int value with a default
func (c *Config) GetInt(key string, defaultValue int) int {
	if c.Viper.IsSet(key) {
		return c.Viper.GetInt(key)
	}
	return defaultValue
}

// GetBool gets a bool value with a default
func (c *Config) GetBool(key string, defaultValue bool) bool {
	if c.Viper.IsSet(key) {
		return c.Viper.GetBool(key)
	}
	return defaultValue
}

// GetInt64 gets an int64 value with a default
func (c *Config) GetInt64(key string, defaultValue int64) int64 {
	if c.Viper.IsSet(key) {
		return c.Viper.GetInt64(key)
	}
	return defaultValue
}

// GetDriver returns the database driver name
func (c *Config) GetDriver() string {
	return c.GetString("database.driver", "postgres")
}

// GetStringSlice gets a string slice value with a default.
func (c *Config) GetStringSlice(key string, defaultValue []string) []string {
	if c.Viper.IsSet(key) {
		return c.Viper.GetStringSlice(key)
	}
	return defaultValue
}

// GetDuration gets a duration value with a default
func (c *Config) GetDuration(key string, defaultValue time.Duration) time.Duration {
	if c.Viper.IsSet(key) {
		return c.Viper.GetDuration(key)
	}
	return defaultValue
}

// PoolConfig represents database connection pool configuration.
// This is a copy to avoid import cycles with the db package.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// GetPoolConfig returns the database connection pool configuration.
// This provides a convenient way to get all pool settings at once.
func (c *Config) GetPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    c.GetInt("database.max_open_conns", 25),
		MaxIdleConns:    c.GetInt("database.max_idle_conns", 10),
		ConnMaxLifetime: c.GetDuration("database.conn_max_lifetime", 5*time.Minute),
		ConnMaxIdleTime: c.GetDuration("database.conn_max_idle_time", 2*time.Minute),
	}
}
