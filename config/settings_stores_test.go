package config

import (
	"strings"
	"testing"
)

// server.stores is deprecated and ignored: framework stores always use the database.
func TestSettingsWarnings_DeprecatedServerStores(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("server.stores", "memory")
	warnings := cfg.SettingsWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "server.stores is deprecated and ignored") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("SettingsWarnings() = %q, want deprecation warning for server.stores", warnings)
	}
}

func TestSettingsWarnings_DeprecatedServerStoresFromEnv(t *testing.T) {
	t.Setenv("FORGE_SERVER_STORES", "database")
	cfg := NewConfig()
	warnings := cfg.SettingsWarnings()
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "server.stores is deprecated and ignored") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("SettingsWarnings() = %q, want deprecation warning for server.stores", warnings)
	}
}
