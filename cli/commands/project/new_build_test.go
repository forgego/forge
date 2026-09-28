package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/forgego/forge/cli/templates"
	"github.com/stretchr/testify/require"
)

// repoRoot returns the Forge module root, two directories above this package.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// TestNewProjectCompiles scaffolds each template and compiles it against this
// checkout, so a template that references missing files or APIs fails here
// rather than on a user's machine.
func TestNewProjectCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a generated project; skipped in -short mode")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go toolchain not on PATH")
	}
	root := repoRoot(t)
	goSum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	require.NoError(t, err)

	for _, tpl := range []templates.ProjectTemplate{templates.TemplateSimple, templates.TemplateAdvanced} {
		t.Run(string(tpl), func(t *testing.T) {
			projectPath := filepath.Join(t.TempDir(), "buildcheck")
			require.NoError(t, createProjectStructure(projectPath, "buildcheck", tpl, "postgres", true))
			require.NoError(t, os.WriteFile(filepath.Join(projectPath, "go.sum"), goSum, 0o644))

			run := func(args ...string) {
				cmd := exec.Command(goBin, args...)
				cmd.Dir = projectPath
				// Resolve Forge from this checkout and every other module from the
				// local module cache, which the root module's tests already need.
				cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off", "GOWORK=off")
				out, err := cmd.CombinedOutput()
				require.NoError(t, err, "go %v failed:\n%s", args, out)
			}
			run("mod", "edit", "-replace", "github.com/forgego/forge="+root)
			run("build", "./...")
			run("vet", "./...")
		})
	}
}
