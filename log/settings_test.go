package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFromSettingsAppliesLevelAndFormat(t *testing.T) {
	cfg := ConfigFromSettings(config.LoggingSettings{Level: "WARN", Format: "json"})
	assert.Equal(t, LevelWarn, cfg.Level)
	assert.Equal(t, FormatJSON, cfg.Format)
	require.Len(t, cfg.Outputs, 1)
	assert.Equal(t, OutputConsole, cfg.Outputs[0].Type)
	assert.Equal(t, LevelWarn, cfg.Outputs[0].Level, "an output without a level uses logging.level")
	assert.False(t, isDevelopmentMode(cfg))

	dev := ConfigFromSettings(config.LoggingSettings{Level: "debug", Format: "console"})
	assert.True(t, isDevelopmentMode(dev), "console format selects the development encoder")
}

func TestNewLoggerFromSettingsHonorsLevelAndFileOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	logger, err := NewLoggerFromSettings(config.LoggingSettings{
		Level:  "warn",
		Format: "json",
		Outputs: []config.LoggingOutputConfig{
			{Type: "file", Enabled: true, Path: path},
		},
	})
	require.NoError(t, err)
	logger.Info("info-is-below-the-configured-level")
	logger.Warn("warn-is-logged")
	require.NoError(t, logger.Close())

	written, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(written), "warn-is-logged")
	assert.NotContains(t, string(written), "info-is-below-the-configured-level")
}

func TestNewLoggerFromSettingsRejectsInvalidValues(t *testing.T) {
	_, err := NewLoggerFromSettings(config.LoggingSettings{Level: "loud", Format: "json"})
	require.Error(t, err)
	_, err = NewLoggerFromSettings(config.LoggingSettings{Level: "info", Format: "xml"})
	require.Error(t, err)
	_, err = NewLoggerFromSettings(config.LoggingSettings{Level: "info", Format: "json",
		Outputs: []config.LoggingOutputConfig{{Type: "file", Enabled: true}}})
	require.Error(t, err, "a file output needs a path")
}
