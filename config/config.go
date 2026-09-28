package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config wraps viper for configuration management
type Config struct {
	*viper.Viper
	generatedSecrets []string
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	// Load local .env secrets (e.g. written by `forge new`) so they are
	// visible through the config instead of being replaced by ephemeral
	// generated values.
	loadDotEnv()
	v := viper.New()
	v.SetEnvPrefix("FORGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	// Explicit bindings for security secrets so generated projects can
	// reference them through the environment (e.g. FORGE_SECURITY_SECRET_KEY).
	_ = v.BindEnv("security.secret_key", "FORGE_SECURITY_SECRET_KEY")
	_ = v.BindEnv("security.csrf_secret_key", "FORGE_SECURITY_CSRF_SECRET_KEY")
	_ = v.BindEnv("security.session_secret", "FORGE_SECURITY_SESSION_SECRET")
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
	v.SetDefault("security.csrf_exempt_paths", []string{})
	v.SetDefault("admin.enabled", true)
	v.SetDefault("admin.path", "/admin")
	v.SetDefault("admin.title", "forge Admin")
	v.SetDefault("admin.header_title", "forge")
	v.SetDefault("admin.site_name", "forge")

	// Read config file (ignore errors if file doesn't exist)
	_ = v.ReadInConfig()

	c := &Config{Viper: v}
	c.ensureSecrets()
	return c
}

// IsPlaceholderSecret detects unconfigured or insecure placeholder secrets.
// It is exported so production validation (forge/server) reuses the same
// definition instead of duplicating it.
func IsPlaceholderSecret(val string) bool {
	s := strings.TrimSpace(val)
	if s == "" {
		return true
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "change-me") || lower == "secret" || lower == "default"
}

// isPlaceholderSecret detects unconfigured or insecure placeholder secrets.
func isPlaceholderSecret(val string) bool {
	return IsPlaceholderSecret(val)
}

// loadDotEnv loads a `.env` file from the config search path if present,
// setting only variables not already in the environment.
func loadDotEnv() {
	for _, dir := range []string{".", "./config", "../config"} {
		path := filepath.Join(dir, ".env")
		if st, err := os.Stat(path); err != nil || st.IsDir() {
			continue
		}
		loadDotEnvFile(path)
	}
}

// loadDotEnvFile parses KEY=VALUE lines from path. Comments and blank lines
// are skipped, single- and double-quoted values are unquoted, and malformed
// lines without `=` are ignored rather than fatal. Variables already present
// in the environment are never overwritten.
func loadDotEnvFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(rest)
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			first, last := value[0], value[len(value)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

// randRead is the randomness source for secret generation. It is a variable
// (rather than a direct crypto/rand call) so tests can force generation
// failures and verify fail-closed behavior.
var randRead = rand.Read

// ensureSecrets generates random secrets for any secret key that was not
// explicitly configured or was set to an insecure placeholder. Shipping
// predictable default secrets means every deployment shares the same signing keys.
func (c *Config) ensureSecrets() {
	for _, key := range []string{
		"security.secret_key",
		"security.csrf_secret_key",
		"security.session_secret",
	} {
		val := c.Viper.GetString(key)
		if c.Viper.IsSet(key) && !isPlaceholderSecret(val) {
			continue
		}
		// Mark the key as generated BEFORE attempting generation so a
		// randomness failure still blocks production startup (fail closed)
		// instead of letting it start with a missing or placeholder secret.
		c.generatedSecrets = append(c.generatedSecrets, key)
		var buf [32]byte
		if _, err := randRead(buf[:]); err != nil {
			log.Printf("forge/config: WARNING: could not generate random value for %s: %v", key, err)
			continue
		}
		c.Viper.Set(key, hex.EncodeToString(buf[:]))
		if isPlaceholderSecret(val) && val != "" {
			log.Printf("forge/config: WARNING: %s is set to an insecure placeholder; overriding with a generated ephemeral value (set it explicitly for production)", key)
		} else {
			log.Printf("forge/config: WARNING: %s is not configured; using a generated ephemeral value (set it explicitly for production)", key)
		}
	}
}

// GeneratedSecrets returns the names of security settings whose values were
// generated because they were missing or insecure placeholders.
func (c *Config) GeneratedSecrets() []string {
	if c == nil || len(c.generatedSecrets) == 0 {
		return nil
	}
	keys := append([]string(nil), c.generatedSecrets...)
	sort.Strings(keys)
	return keys
}

// Set assigns a configuration value. When a secret key is explicitly set to
// a non-placeholder value, it is no longer reported as generated so
// production validation does not rely on stale provenance after
// post-construction overrides.
func (c *Config) Set(key string, value interface{}) {
	c.Viper.Set(key, value)
	switch key {
	case "security.secret_key", "security.csrf_secret_key", "security.session_secret":
		if s, ok := value.(string); ok && !IsPlaceholderSecret(s) {
			kept := c.generatedSecrets[:0]
			for _, k := range c.generatedSecrets {
				if k != key {
					kept = append(kept, k)
				}
			}
			for i := len(kept); i < len(c.generatedSecrets); i++ {
				c.generatedSecrets[i] = ""
			}
			c.generatedSecrets = kept
		}
	}
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
