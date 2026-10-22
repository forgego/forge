package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Dotted configuration keys have the documented FORGE_ environment form.
func TestEnvironmentOverridesDottedConfigurationKeys(t *testing.T) {
	keys := NewConfig().AllKeys()
	if len(keys) < 20 {
		t.Fatalf("expected default configuration keys, got %d", len(keys))
	}
	for i, key := range keys {
		value := strings.Repeat("x", 64) + string(rune('a'+i))
		t.Setenv("FORGE_"+strings.ToUpper(strings.ReplaceAll(key, ".", "_")), value)
	}
	// The documented form must also work through typed accessors used by the DB layer.
	t.Setenv("FORGE_DATABASE_MAX_OPEN_CONNS", "7")
	t.Setenv("FORGE_DATABASE_CONN_MAX_LIFETIME", "13m")
	cfg := NewConfig()
	for i, key := range keys {
		if key == "database.max_open_conns" || key == "database.conn_max_lifetime" {
			continue
		}
		want := strings.Repeat("x", 64) + string(rune('a'+i))
		if got := cfg.GetString(key, ""); got != want {
			t.Errorf("%s = %q, want environment override", key, got)
		}
	}
	if pool := cfg.GetPoolConfig(); pool.MaxOpenConns != 7 || pool.ConnMaxLifetime != 13*time.Minute {
		t.Fatalf("environment pool config = %+v", pool)
	}
}

// The same config file can be found from the project and its cmd directory.
func TestConfigSearchesCurrentAndAdjacentConfigDirectories(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{root, filepath.Join(root, "config"), filepath.Join(root, "cmd")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "config", "config.yaml"), []byte("lookup_probe: adjacent-test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	for _, dir := range []string{root, filepath.Join(root, "cmd")} {
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		if got := NewConfig().GetString("lookup_probe", ""); got != "adjacent-test" {
			t.Fatalf("from %s: lookup_probe = %q", dir, got)
		}
	}
}

// DATABASE_URL has no config key or connection-string effect.
func TestDatabaseURLIsNotAConfiguredSetting(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://ignored/elsewhere")
	t.Setenv("FORGE_DATABASE_HOST", "localhost")
	cfg := NewConfig()
	if cfg.IsSet("database_url") || cfg.IsSet("database.url") {
		t.Fatal("unexpected DATABASE_URL setting")
	}
	if got := cfg.GetString("database.host", ""); got != "localhost" {
		t.Fatalf("database host = %q", got)
	}
}
