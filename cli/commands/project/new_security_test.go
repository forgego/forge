package project

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

// readDotEnv parses KEY=VALUE lines from a .env file.
func readDotEnv(t *testing.T, path string) map[string]string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	values := map[string]string{}
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		require.True(t, ok, "malformed .env line: %q", line)
		values[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return values
}

func requireStrongSecret(t *testing.T, key, value string) {
	t.Helper()
	require.Len(t, value, 64)
	_, err := hex.DecodeString(value)
	require.NoError(t, err)
	require.False(t, strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "change-me"))
	require.NotEqual(t, "secret", strings.ToLower(strings.TrimSpace(value)))
	require.NotEqual(t, "default", strings.ToLower(strings.TrimSpace(value)))
}

func TestCreateConfigFileGeneratesIndependentSecretsPerProject(t *testing.T) {
	keys := []string{"FORGE_SECURITY_SECRET_KEY", "FORGE_SECURITY_SESSION_SECRET", "FORGE_SECURITY_CSRF_SECRET_KEY"}
	valuesByProject := make([]map[string]string, 0, 2)

	for i := 0; i < 2; i++ {
		projectPath := filepath.Join(t.TempDir(), "project")
		require.NoError(t, os.MkdirAll(filepath.Join(projectPath, "config"), 0755))
		require.NoError(t, createConfigFile(projectPath, "sqlite"))

		values := readDotEnv(t, filepath.Join(projectPath, ".env"))
		for _, key := range keys {
			requireStrongSecret(t, key, values[key])
		}
		require.NotEqual(t, values[keys[0]], values[keys[1]])
		require.NotEqual(t, values[keys[0]], values[keys[2]])
		require.NotEqual(t, values[keys[1]], values[keys[2]])
		valuesByProject = append(valuesByProject, values)
	}

	for _, key := range keys {
		require.NotEqual(t, valuesByProject[0][key], valuesByProject[1][key])
	}
}

func TestCreateConfigFileKeepsSecretsOutOfCommittedConfig(t *testing.T) {
	// Exercise the full `forge new` project creation path.
	projectPath := filepath.Join(t.TempDir(), "project")
	require.NoError(t, createProjectStructure(projectPath, "project", templates.TemplateSimple, "sqlite", false))

	envValues := readDotEnv(t, filepath.Join(projectPath, ".env"))
	generated := []string{
		envValues["FORGE_SECURITY_SECRET_KEY"],
		envValues["FORGE_SECURITY_SESSION_SECRET"],
		envValues["FORGE_SECURITY_CSRF_SECRET_KEY"],
	}
	for _, value := range generated {
		requireStrongSecret(t, "generated", value)
	}

	configBytes, err := os.ReadFile(filepath.Join(projectPath, "config", "config.yaml"))
	require.NoError(t, err)
	configContent := string(configBytes)
	for _, value := range generated {
		require.NotContains(t, configContent, value, "config.yaml must not carry secret values")
	}
	// The committed config references the secrets through the environment.
	require.Contains(t, configContent, "FORGE_SECURITY_SECRET_KEY")
	require.Contains(t, configContent, "FORGE_SECURITY_SESSION_SECRET")
	require.Contains(t, configContent, "FORGE_SECURITY_CSRF_SECRET_KEY")

	// The committed config parses and carries no secret values.
	cfg := viper.New()
	cfg.SetConfigFile(filepath.Join(projectPath, "config", "config.yaml"))
	require.NoError(t, cfg.ReadInConfig())
	for _, key := range []string{"security.secret_key", "security.session_secret", "security.csrf_secret_key"} {
		require.Empty(t, cfg.GetString(key), "committed config must not define %s", key)
	}

	// The local secrets file is excluded from version control.
	gitignoreBytes, err := os.ReadFile(filepath.Join(projectPath, ".gitignore"))
	require.NoError(t, err)
	require.Contains(t, string(gitignoreBytes), ".env")
}
