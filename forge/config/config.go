package config

import (
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
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 300)
	v.SetDefault("database.conn_max_idle_time", 600)
	v.SetDefault("security.secret_key", "change-me-in-production")
	v.SetDefault("security.csrf_secret_key", "change-me-in-production")
	v.SetDefault("security.session_secret", "change-me-in-production")
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
