package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestRootCommandScaffoldsWithoutPrompts runs `forge new` with every choice
// given as a flag, then `forge add app`, the way scripts and CI invoke them.
func TestRootCommandScaffoldsWithoutPrompts(t *testing.T) {
	root := BuildRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	workdir := t.TempDir()
	t.Chdir(workdir)

	// --docker=false must skip the Docker prompt; with no terminal the prompt fails.
	root.SetArgs([]string{"new", "myapp", "--template", "simple", "--database", "sqlite", "--docker=false"})
	require.NoError(t, root.Execute())
	require.NoFileExists(t, filepath.Join(workdir, "myapp", "Dockerfile"))

	// `add app` must run its handler, not print help and exit 0.
	t.Chdir(filepath.Join(workdir, "myapp"))
	root.SetArgs([]string{"add", "app", "blog"})
	require.NoError(t, root.Execute())
	for _, name := range []string{"models.go", "admin.go", "api.go"} {
		_, err := os.Stat(filepath.Join(workdir, "myapp", "app", "blog", name))
		require.NoError(t, err)
	}
}
