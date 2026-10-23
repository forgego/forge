package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

func TestProjectTemplateSecurity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	projectDir := filepath.Join(t.TempDir(), "fresh-project")
	stdout, _, err := testhelpers.RunCLI(ctx, t.TempDir(), nil, []string{
		"new", "fresh-project", "--path", projectDir,
		"--template", "simple", "--database", "sqlite", "--docker=false",
	}, 30*time.Second)
	require.NoError(t, err, "forge new output: %s", stdout)
	requireNoWeakTemplateValues(t, projectDir)
	generatedSecrets := readGeneratedSecrets(t, filepath.Join(projectDir, ".env"))

	// The scaffold's ignored .env is for local development. Remove it to model
	// production startup without any configured security keys.
	require.NoError(t, os.Remove(filepath.Join(projectDir, ".env")))
	require.NoError(t, patchGeneratedProject(projectDir))

	binName := "server"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(projectDir, "bin", binName)
	build := exec.CommandContext(ctx, "go", "build", "-o", binPath, "./cmd/server")
	build.Dir = projectDir
	build.Env = append(os.Environ(), "GOTOOLCHAIN=go1.26.10", "GOFLAGS=-buildvcs=false")
	buildOutput, err := build.CombinedOutput()
	require.NoError(t, err, "build generated server: %s", buildOutput)

	cmd := exec.CommandContext(ctx, binPath)
	cmd.Dir = projectDir
	cmd.Env = append(os.Environ(),
		"FORGE_APP_ENV=production",
		"FORGE_APP_DEBUG=false",
		"FORGE_SECURITY_SECRET_KEY=",
		"FORGE_SECURITY_CSRF_SECRET_KEY=",
		"FORGE_SECURITY_SESSION_SECRET=",
	)
	output, err := cmd.CombinedOutput()
	require.Error(t, err, "production server must reject missing security keys")
	message := string(output)
	require.Contains(t, message, "production requires explicit security.secret_key")
	require.NotContains(t, message, "Starting server on", "validation must fail before listening")
	for _, secret := range generatedSecrets {
		require.NotContains(t, message, secret, "startup error must not disclose a generated secret")
	}
	for _, weak := range weakTemplateValues {
		require.NotContains(t, strings.ToLower(message), weak, "startup error must not disclose weak or secret values")
	}
}

var weakTemplateValues = []string{
	"changeme", "change-me", "insecure", "default-secret", "your-secret-key",
	"your-256-bit-secret", "your-secret-key-here", "supersecret",
}

func readGeneratedSecrets(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var secrets []string
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if ok && strings.HasPrefix(key, "FORGE_SECURITY_") {
			secrets = append(secrets, value)
		}
	}
	require.Len(t, secrets, 3)
	return secrets
}

func requireNoWeakTemplateValues(t *testing.T, root string) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := strings.ToLower(string(data))
		for _, weak := range weakTemplateValues {
			require.NotContains(t, content, weak, "generated file %s contains an insecure placeholder", strings.TrimPrefix(path, root+string(filepath.Separator)))
		}
		require.NotRegexp(t, regexp.MustCompile(`\bsecret\b`), content, "generated file %s contains the default secret value", strings.TrimPrefix(path, root+string(filepath.Separator)))
		return nil
	})
	require.NoError(t, err)
}
