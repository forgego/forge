package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// chdirToTempProject switches the working directory to a fresh temp dir so
// NewConfig picks up the local .env through its config search path.
func chdirToTempProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(dir))
	return dir
}

func TestDotEnvValueIsVisibleThroughConfig(t *testing.T) {
	dir := chdirToTempProject(t)
	require.NoError(t, os.Unsetenv("FORGE_APP_NAME"))
	t.Cleanup(func() { _ = os.Unsetenv("FORGE_APP_NAME") })
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"),
		[]byte("FORGE_APP_NAME=dotenv-app-name\n"), 0644))

	cfg := NewConfig()
	require.Equal(t, "dotenv-app-name", cfg.GetString("app.name", ""))
}

func TestDotEnvQuotedValueIsUnquoted(t *testing.T) {
	dir := chdirToTempProject(t)
	require.NoError(t, os.Unsetenv("FORGE_APP_NAME"))
	t.Cleanup(func() { _ = os.Unsetenv("FORGE_APP_NAME") })
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"),
		[]byte("# comment line\n\nFORGE_APP_NAME=\"quoted app name\"\n"), 0644))

	cfg := NewConfig()
	require.Equal(t, "quoted app name", cfg.GetString("app.name", ""))
}

func TestExistingEnvironmentWinsOverDotEnv(t *testing.T) {
	dir := chdirToTempProject(t)
	t.Setenv("FORGE_APP_NAME", "from-environment")
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"),
		[]byte("FORGE_APP_NAME=from-dotenv-file\n"), 0644))

	cfg := NewConfig()
	require.Equal(t, "from-environment", cfg.GetString("app.name", ""))
}

func TestMalformedDotEnvLineIsIgnored(t *testing.T) {
	dir := chdirToTempProject(t)
	require.NoError(t, os.Unsetenv("FORGE_APP_NAME"))
	t.Cleanup(func() { _ = os.Unsetenv("FORGE_APP_NAME") })
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".env"),
		[]byte("THIS LINE HAS NO EQUALS SIGN\nFORGE_APP_NAME=after-malformed\n"), 0644))

	cfg := NewConfig()
	require.NotNil(t, cfg)
	require.Equal(t, "after-malformed", cfg.GetString("app.name", ""))
}
