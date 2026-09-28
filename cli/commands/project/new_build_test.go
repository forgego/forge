package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"testing"

	"github.com/forgego/forge/cli/templates"
	"github.com/forgego/forge/config"
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

// TestDockerScaffoldMatchesToolchainAndListensOnAllInterfaces checks the
// --docker files: the builder image must satisfy the generated go directive
// (official golang images set GOTOOLCHAIN=local, so an older image cannot
// build the project), and compose must bind the server beyond the container's
// loopback interface so the published port is reachable.
func TestDockerScaffoldMatchesToolchainAndListensOnAllInterfaces(t *testing.T) {
	data := templates.TemplateData{ProjectName: "dockercheck", ForgeVersion: "v0.0.0"}

	goMod, err := templates.RenderTemplate("go_mod.tmpl", data)
	require.NoError(t, err)
	goDirective := regexp.MustCompile(`(?m)^go (\d+)\.(\d+)`).FindSubmatch(goMod)
	require.NotNil(t, goDirective, "generated go.mod has no go directive")

	dockerfile, err := templates.RenderTemplate("dockerfile.tmpl", data)
	require.NoError(t, err)
	builder := regexp.MustCompile(`(?m)^FROM golang:(\d+)\.(\d+)`).FindSubmatch(dockerfile)
	require.NotNil(t, builder, "Dockerfile has no golang builder stage")

	atoi := func(b []byte) int {
		n, err := strconv.Atoi(string(b))
		require.NoError(t, err)
		return n
	}
	wantMajor, wantMinor := atoi(goDirective[1]), atoi(goDirective[2])
	gotMajor, gotMinor := atoi(builder[1]), atoi(builder[2])
	require.True(t, gotMajor > wantMajor || (gotMajor == wantMajor && gotMinor >= wantMinor),
		"Dockerfile builds with Go %d.%d but go.mod requires %d.%d", gotMajor, gotMinor, wantMajor, wantMinor)

	compose, err := templates.RenderTemplate("compose.tmpl", data)
	require.NoError(t, err)
	require.Contains(t, string(compose), "FORGE_SERVER_HOST=0.0.0.0")

	t.Setenv("FORGE_SERVER_HOST", "0.0.0.0")
	require.Equal(t, "0.0.0.0", config.LoadSettings(config.NewConfig()).Server.Host)
}
