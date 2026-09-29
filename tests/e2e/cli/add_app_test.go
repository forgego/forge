package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/tests/testhelpers"
)

// TestCLIAddAppExampleCompiles follows the quickstart: the example model that
// `forge add app --example` writes, and the code `forge generate` writes for
// it, must compile against the framework.
func TestCLIAddAppExampleCompiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	workdir, cleanup := testhelpers.TempWorkdir(t, "forge-e2e-add-app-")
	defer cleanup()

	projectDir := filepath.Join(workdir, "myapp")
	stdout, _, err := testhelpers.RunCLI(ctx, workdir, nil, []string{
		"new", "myapp",
		"--path", projectDir,
		"--template", "simple",
		"--database", "sqlite",
		"--docker=false",
	}, 30*time.Second)
	require.NoError(t, err, "forge new output: %s", stdout)
	require.NoError(t, patchGeneratedProject(projectDir))

	stdout, _, err = testhelpers.RunCLI(ctx, projectDir, nil, []string{"add", "app", "blog", "--example"}, 30*time.Second)
	require.NoError(t, err, "forge add app output: %s", stdout)

	stdout, _, err = testhelpers.RunCLI(ctx, projectDir, nil, []string{
		"generate",
	}, 60*time.Second)
	require.NoError(t, err, "forge generate output: %s", stdout)

	// The model the ORM and the admin work with must carry the generated
	// columns, not only the schema methods.
	check := "package blog\n\nvar _ = Example{}.Name\nvar _ = Example{}.IsActive\nvar _ = ExampleObjects\n"
	require.NoError(t, os.WriteFile(filepath.Join(projectDir, "app", "blog", "example_fields_check.go"), []byte(check), 0o644))

	// The app package is new since patchGeneratedProject ran, so tidy again
	// before building it.
	for _, args := range [][]string{{"mod", "tidy"}, {"build", "./app/..."}} {
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = projectDir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "go %v: %s", args, out)
	}
}
