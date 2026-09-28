package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGeneratedAPIRegistersAndRejectsUnknownFields compiles generated API
// code in its own module and serves it: the generated viewset must pass the
// router's configuration check with the generated orm.Manager, and it must
// reject an unknown request key before touching the database.
func TestGeneratedAPIRegistersAndRejectsUnknownFields(t *testing.T) {
	if testing.Short() {
		t.Skip("builds a generated module")
	}
	forgePath, err := filepath.Abs("..")
	require.NoError(t, err)
	moduleDir := t.TempDir()
	pkgDir := filepath.Join(moduleDir, "catalog")
	require.NoError(t, os.MkdirAll(pkgDir, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(
		"module example.com/apiruntime\n\ngo 1.23\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => "+forgePath+"\n"), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "models.go"), []byte(`package catalog

import "github.com/forgego/forge/schema"

type Product struct {
	schema.BaseSchema
	ID        int64  `+"`json:\"id\" db:\"id\"`"+`
	Name      string `+"`json:\"name\" db:\"name\"`"+`
	CreatedAt string `+"`json:\"created_at\" db:\"created_at\"`"+`
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("created_at", schema.Editable(false)),
	}
}
`), 0o644))

	gen := NewGenerator(pkgDir, pkgDir)
	gen.SetGenerateAPI(true)
	require.NoError(t, gen.Generate())

	require.NoError(t, os.WriteFile(filepath.Join(pkgDir, "api_runtime_test.go"), []byte(`package catalog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	forgehttp "github.com/forgego/forge/server"
)

func TestGeneratedAPI(t *testing.T) {
	router := forgehttp.NewRouter()
	RegisterAPIRoutes(router)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/products/", strings.NewReader(`+"`"+`{"name":"x","nmae":"typo","id":7,"created_at":"now"}`+"`"+`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "nmae") {
		t.Fatalf("unknown key: got %d %s", rec.Code, rec.Body.String())
	}
	for _, echoed := range []string{"\"id\"", "created_at", "\"name\""} {
		if strings.Contains(rec.Body.String(), echoed) {
			t.Fatalf("known key %s reported as unknown: %s", echoed, rec.Body.String())
		}
	}
}
`), 0o644))

	cmd := exec.Command("go", "test", "-mod=mod", "./...")
	cmd.Dir = moduleDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated API module failed:\n%s", output)
}
