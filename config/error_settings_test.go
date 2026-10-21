package config

import (
	"github.com/spf13/viper"
	"testing"
)

func errorSettingsConfig() *Config { return &Config{Viper: viper.New()} }

func TestStackTraceHidingSettingCanBeConfigured(t *testing.T) {
	cfg := errorSettingsConfig()
	if !LoadSettings(cfg).Errors.Sanitization.HideStackTraces {
		t.Fatal("stack traces should be hidden by default")
	}
	cfg.Set("errors.sanitization.hide_stack_traces", false)
	if LoadSettings(cfg).Errors.Sanitization.HideStackTraces {
		t.Fatal("explicit false should be retained")
	}
}
