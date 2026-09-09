package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator_GenerateAndAPI(t *testing.T) {
	tmpDir := t.TempDir()

	modelSrc := `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Product struct {
	schema.BaseSchema
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.Decimal("price").Required().Build(),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_products",
		VerboseName: "Product",
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{}
}

func (Product) Hooks() *schema.ModelHooks {
	return nil
}
`
	modelFile := filepath.Join(tmpDir, "models.go")
	require.NoError(t, os.WriteFile(modelFile, []byte(modelSrc), 0644))

	// Run generator with API enabled
	gen := NewGenerator(tmpDir, tmpDir)
	gen.SetGenerateAPI(true)

	err := gen.Generate()
	require.NoError(t, err)

	// Verify gen.go
	genFile := filepath.Join(tmpDir, "gen.go")
	assert.FileExists(t, genFile)
	genBytes, err := os.ReadFile(genFile)
	require.NoError(t, err)
	assert.Contains(t, string(genBytes), "type ProductGenerated struct")
	assert.Contains(t, string(genBytes), "var ProductObjects, _ = orm.NewManager[Product]")

	// Verify api_gen.go
	apiGenFile := filepath.Join(tmpDir, "api_gen.go")
	assert.FileExists(t, apiGenFile)
	apiBytes, err := os.ReadFile(apiGenFile)
	require.NoError(t, err)
	assert.Contains(t, string(apiBytes), "type ProductSerializer struct")
	assert.Contains(t, string(apiBytes), "type ProductViewSet struct")
	assert.Contains(t, string(apiBytes), "RegisterProductRoutes(router *forgehttp.Router)")
	assert.Contains(t, string(apiBytes), "RegisterAPIRoutes(router *forgehttp.Router)")
}
