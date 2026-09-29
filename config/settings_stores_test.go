package config

import "testing"

// server.stores keeps its pre-#293 behavior (memory) unless an app opts in.
func TestLoadSettings_StoresDefaultsToMemory(t *testing.T) {
	if got := LoadSettings(NewConfig()).Server.Stores; got != "memory" {
		t.Fatalf("Stores default = %q, want memory", got)
	}
}

func TestLoadSettings_StoresFromEnv(t *testing.T) {
	t.Setenv("FORGE_SERVER_STORES", "database")
	if got := LoadSettings(NewConfig()).Server.Stores; got != "database" {
		t.Fatalf("Stores = %q, want database", got)
	}
}
