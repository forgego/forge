package project

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A dashed resource name must yield a valid Go identifier for the Register
// function while the URL segment keeps the dash (#294).
func TestGenerateAPICodeDashedResourceName(t *testing.T) {
	code, err := generateAPICode("blog", "Post", "blog-posts", "blog-posts", false)
	require.NoError(t, err)
	assert.Contains(t, code, "func RegisterBlogPostsAPI(router *httplib.Router) {")
	assert.Contains(t, code, `apiRouter.Register("blog-posts", viewset)`)

	src := "package blog\n" + code
	_, err = parser.ParseFile(token.NewFileSet(), "api.go", src, 0)
	require.NoError(t, err, "generated code must parse:\n%s", src)

	_, err = generateAPICode("blog", "Post", "2fa", "2fa", false)
	require.Error(t, err, "a name that cannot start an identifier is rejected")
}

func TestAddAPIRejectsDuplicateDeclarations(t *testing.T) {
	dir := t.TempDir()
	previous, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(previous) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.26\n"), 0600))
	appDir := filepath.Join(dir, "app", "blog")
	require.NoError(t, os.MkdirAll(appDir, 0755))
	apiPath := filepath.Join(appDir, "api.go")
	require.NoError(t, os.WriteFile(apiPath, []byte("package blog\n"), 0600))
	require.NoError(t, os.Chdir(dir))
	cmd := NewAddAPICommand().Definition()
	require.NoError(t, cmd.Flags().Set("app", "blog"))
	require.NoError(t, cmd.Flags().Set("model", "Post"))
	add := NewAddAPICommand()
	require.NoError(t, add.Execute(&core.Context{Cmd: cmd}, []string{"posts"}))
	first, err := os.ReadFile(apiPath)
	require.NoError(t, err)
	err = add.Execute(&core.Context{Cmd: cmd}, []string{"posts"})
	require.ErrorContains(t, err, "RegisterPostsAPI")
	second, err := os.ReadFile(apiPath)
	require.NoError(t, err)
	require.Equal(t, first, second)
}
