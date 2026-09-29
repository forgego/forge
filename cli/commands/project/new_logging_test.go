package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/templates"
	"github.com/forgego/forge/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The scaffold must build its logger from the logging.* settings instead of
// only app.debug, and its config must carry a logging section (#294).
func TestNewProjectLoggerReadsLoggingSettings(t *testing.T) {
	for _, tpl := range []templates.ProjectTemplate{templates.TemplateSimple, templates.TemplateAdvanced} {
		t.Run(string(tpl), func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "logcheck")
			require.NoError(t, createProjectStructure(projectPath, "logcheck", tpl, "sqlite", false))

			mainGo, err := os.ReadFile(filepath.Join(projectPath, "cmd", "server", "main.go"))
			require.NoError(t, err)
			assert.Contains(t, string(mainGo), "forgelog.NewLoggerFromSettings(settings.Logging)")
			assert.NotContains(t, string(mainGo), "NewLogger(settings.App.Debug)")

			cfg := config.NewConfig()
			cfg.SetConfigFile(filepath.Join(projectPath, "config", "config.yaml"))
			require.NoError(t, cfg.ReadInConfig())
			settings := config.LoadSettings(cfg)
			// The committed config carries production logging; the local
			// .env switches to debug console output.
			assert.Equal(t, "info", settings.Logging.Level)
			assert.Equal(t, "json", settings.Logging.Format)
			env := readDotEnv(t, filepath.Join(projectPath, ".env"))
			assert.Equal(t, "debug", env["FORGE_LOGGING_LEVEL"])
			assert.Equal(t, "console", env["FORGE_LOGGING_FORMAT"])
		})
	}
}
