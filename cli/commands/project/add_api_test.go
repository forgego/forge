package project

import (
	"go/parser"
	"go/token"
	"testing"

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
