package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/stretchr/testify/require"
)

const unwirableMain = `package main

func main() {}
`

const wirableMain = `package main

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
	var _ = stdlog.Println
}
`

// TestAuthCommandCleansUpPartialScaffoldWhenWiringFails ensures a wiring
// failure leaves no partial app directory behind, so the user can fix main.go
// and retry without deleting anything by hand.
func TestAuthCommandCleansUpPartialScaffoldWhenWiringFails(t *testing.T) {
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module example.com/wiretest\n\ngo 1.23\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(unwirableMain), 0644))

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(projectPath))

	err = NewAuthCommand().Execute(&core.Context{}, nil)
	require.Error(t, err, "expected wiring failure with an unwirable main.go")

	_, statErr := os.Stat(filepath.Join(projectPath, "app", "auth"))
	require.True(t, os.IsNotExist(statErr), "partial auth app directory was left behind after wiring failure")

	// After fixing main.go a second run must succeed.
	require.NoError(t, os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(wirableMain), 0644))
	require.NoError(t, NewAuthCommand().Execute(&core.Context{}, nil))
	require.FileExists(t, filepath.Join(projectPath, "app", "auth", "models.go"))
	require.FileExists(t, filepath.Join(projectPath, "app", "auth", "api.go"))
}
