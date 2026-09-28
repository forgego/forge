package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	codegen "github.com/forgego/forge/codegen"
	"github.com/stretchr/testify/require"
)

func TestAuthScaffoldBuildsAsExternalModule(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping external scaffold build in short mode")
	}

	forgePath, err := filepath.Abs("../../..")
	require.NoError(t, err)
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module example.com/authscaffold\n\ngo 1.23\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => "+forgePath+"\n"), 0644))

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(projectPath))
	require.NoError(t, NewAuthCommand().Execute(&core.Context{}, nil))

	// UserObjects is owned by the generated code, so generation must run
	// before the scaffold compiles.
	appPath := filepath.Join(projectPath, "app", "auth")
	require.NoError(t, codegen.NewGenerator(appPath, appPath).Generate())

	cmd := exec.Command("go", "build", "-mod=mod", "./...")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated auth scaffold did not build:\n%s", output)
}

func TestWireAuthAPIUsesDatabaseAndConfiguredSigningKey(t *testing.T) {
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module example.com/project\n\ngo 1.23\n"), 0644))
	mainCode := `package main

import (
	stdlog "log"
	"github.com/forgego/forge/server"
)

func configure() {
	srv.RegisterRoutes(func(router *server.Router) {
		if settings.Admin.Enabled {
			router.Mount(settings.Admin.Path, adminSite.Handler())
		}
	})
}
`
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(mainCode), 0644))

	require.NoError(t, wireAuthAPI(projectPath))
	updated, err := os.ReadFile(filepath.Join(projectPath, "main.go"))
	require.NoError(t, err)
	content := string(updated)
	require.Contains(t, content, `"example.com/project/app/auth"`)
	require.Contains(t, content, "auth.UserObjects.SetDB(database)")
	require.Contains(t, content, "auth.RegisterAuthAPI(router, []byte(settings.Security.SecretKey))")
}
