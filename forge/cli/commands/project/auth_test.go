package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/stretchr/testify/require"
)

func TestAuthCommandExecute_ScaffoldsAuthAPIWithLoginLogoutAndJWT(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.25.0\n"), 0644))
	require.NoError(t, os.Chdir(tmp))

	cmd := NewAuthCommand()
	err = cmd.Execute(&core.Context{}, nil)
	require.NoError(t, err)

	apiPath := filepath.Join(tmp, "app", "auth", "api.go")
	contentBytes, err := os.ReadFile(apiPath)
	require.NoError(t, err)
	content := string(contentBytes)

	require.Contains(t, content, `router.Post("/api/v1/auth/login", handleLogin)`)
	require.Contains(t, content, `router.Post("/api/v1/auth/logout", handleLogout)`)
	require.Contains(t, content, "func generateJWTToken(userID, username string) string")
	require.Contains(t, content, "hmac.New(sha256.New")
	require.Contains(t, content, "change-me-in-production")
	require.NotContains(t, strings.ToLower(content), "todo:")
}

func TestAuthCommandExecute_ReturnsErrorWhenAuthAppAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.Chdir(origWD)
	})

	require.NoError(t, os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/test\n\ngo 1.25.0\n"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "app", "auth"), 0755))
	require.NoError(t, os.Chdir(tmp))

	cmd := NewAuthCommand()
	err = cmd.Execute(&core.Context{}, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "auth app already exists")
}
