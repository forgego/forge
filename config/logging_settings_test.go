package config

import (
	"strings"
	"testing"
)

func TestSettingsWarnings_MalformedLoggingOutputs(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{"scalar instead of list", "console"},
		{"list of scalars", []interface{}{"console", "file"}},
		{"wrong field type", []map[string]interface{}{{"type": "file", "enabled": map[string]interface{}{"x": 1}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.Set("logging.outputs", tt.value)

			settings := LoadSettings(cfg)
			if len(settings.Logging.Outputs) != 0 {
				t.Fatalf("Logging.Outputs = %+v, want none for a malformed list", settings.Logging.Outputs)
			}
			warnings := cfg.SettingsWarnings()
			if len(warnings) != 1 || !strings.Contains(warnings[0], "logging.outputs is malformed") {
				t.Fatalf("SettingsWarnings() = %q, want one logging.outputs warning", warnings)
			}
		})
	}
}

func TestSettingsWarnings_ValidOrMissingLoggingOutputs(t *testing.T) {
	cfg := NewConfig()
	if warnings := cfg.SettingsWarnings(); len(warnings) != 0 {
		t.Fatalf("SettingsWarnings() without logging.outputs = %q, want none", warnings)
	}

	cfg.Set("logging.outputs", []map[string]interface{}{
		{"type": "console", "enabled": true},
		{"type": "file", "enabled": true, "path": "logs/app.log"},
	})
	if warnings := cfg.SettingsWarnings(); len(warnings) != 0 {
		t.Fatalf("SettingsWarnings() with a valid list = %q, want none", warnings)
	}

	var nilCfg *Config
	if warnings := nilCfg.SettingsWarnings(); len(warnings) != 0 {
		t.Fatalf("nil Config SettingsWarnings() = %q, want none", warnings)
	}
}
