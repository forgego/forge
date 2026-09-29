package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// New projects keep sessions, throttling counters and the admin's state in
// the database (#293), so they work with more than one instance.
func TestNewProjectUsesDatabaseStores(t *testing.T) {
	for _, database := range []string{"postgres", "sqlite"} {
		t.Run(database, func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "project")
			require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "config"), 0o755))
			require.NoError(t, createConfigFile(projectPath, database))

			v := viper.New()
			v.SetConfigFile(filepath.Join(projectPath, "config", "config.yaml"))
			require.NoError(t, v.ReadInConfig())
			require.Equal(t, "database", v.GetString("server.stores"))
		})
	}

	for _, tpl := range []templates.ProjectTemplate{templates.TemplateSimple, templates.TemplateAdvanced} {
		t.Run(string(tpl), func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "storescheck")
			require.NoError(t, createProjectStructure(projectPath, "storescheck", tpl, "postgres", false))
			mainGo, err := os.ReadFile(filepath.Join(projectPath, "cmd", "server", "main.go"))
			require.NoError(t, err)
			require.Contains(t, string(mainGo), "server.NewServer(cfg, settings, logger, server.WithDatabase(database))")
			require.Contains(t, string(mainGo), "adminSite.UseStores(context.Background(), settings.Server.Stores)")
		})
	}
}
