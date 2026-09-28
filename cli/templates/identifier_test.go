package templates

import (
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoIdentifier(t *testing.T) {
	for name, want := range map[string]string{
		"blog-posts":     "BlogPosts",
		"order_items":    "OrderItems",
		"examples":       "Examples",
		"users.v2":       "UsersV2",
		"--weird--name-": "WeirdName",
		"already Camel":  "AlreadyCamel",
		"café-menus":     "CaféMenus",
	} {
		got, err := GoIdentifier(name)
		require.NoError(t, err, name)
		assert.Equal(t, want, got, name)
		assert.True(t, token.IsIdentifier(got) && token.IsExported(got), "%q -> %q", name, got)
	}
	for _, name := range []string{"", "---", "2fa-codes"} {
		_, err := GoIdentifier(name)
		assert.Error(t, err, name)
	}
}

func TestAPITemplateConvertsResourceNameToIdentifier(t *testing.T) {
	out, err := RenderTemplate("api.go.tmpl", TemplateData{
		AppName: "blog", ModelName: "Post", ResourceName: "blog-posts",
	})
	require.NoError(t, err)
	code := string(out)
	assert.Contains(t, code, "func RegisterBlogPostsAPI(router *server.Router)")
	assert.Contains(t, code, `apiRouter.Register("blog-posts", viewset)`)
	assert.False(t, strings.Contains(code, "Registerblog-posts"))
}
