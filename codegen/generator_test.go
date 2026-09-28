package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorGenerateEmitsMustNewManagerDeclaration(t *testing.T) {
	tmpDir := t.TempDir()

	modelSrc := `package testmodels

import (
	"github.com/forgego/forge/schema"
)

type Product struct {
	schema.BaseSchema
	ID int64
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
	assert.Contains(t, string(genBytes), "var ProductObjects = orm.MustNewManager[Product]")

	// Verify api_gen.go
	apiGenFile := filepath.Join(tmpDir, "api_gen.go")
	assert.FileExists(t, apiGenFile)
	apiBytes, err := os.ReadFile(apiGenFile)
	require.NoError(t, err)
	assert.Contains(t, string(apiBytes), "type ProductSerializer struct")
	assert.Contains(t, string(apiBytes), "type ProductViewSet struct")
	assert.Contains(t, string(apiBytes), "vs.ExcludeResponseFields = api.NonSerializableFields(&Product{})")
	assert.Contains(t, string(apiBytes), "vs.ReadOnlyRequestFields = api.NonEditableFields(&Product{})")
	assert.Contains(t, string(apiBytes), "RegisterProductRoutes(router *forgehttp.Router)")
	assert.Contains(t, string(apiBytes), "RegisterAPIRoutes(router *forgehttp.Router)")
}

func TestValidateAPIModels_AcceptsOnlyInt64IDPrimaryKey(t *testing.T) {
	tests := []struct {
		name    string
		model   *ModelDefinition
		wantErr string
	}{
		{name: "auto-increment int64 id", model: &ModelDefinition{Name: "Product", Fields: []FieldDefinition{{Name: "id", Type: "Int64", GoType: "int64", PrimaryKey: true, AutoIncrement: true}}}},
		{name: "non-auto-increment int64 id", model: &ModelDefinition{Name: "Manual", Fields: []FieldDefinition{{Name: "id", Type: "Int64", GoType: "int64", PrimaryKey: true}}}, wantErr: "model Manual"},
		{name: "no primary key", model: &ModelDefinition{Name: "Log", Fields: []FieldDefinition{{Name: "message", Type: "String", GoType: "string"}}}, wantErr: "has no primary key"},
		{name: "int32 id", model: &ModelDefinition{Name: "Counter", Fields: []FieldDefinition{{Name: "id", Type: "Int32", GoType: "int32", PrimaryKey: true}}}, wantErr: "require an int64 primary key"},
		{name: "int64 non id", model: &ModelDefinition{Name: "Product", Fields: []FieldDefinition{{Name: "product_id", Type: "Int64", GoType: "int64", PrimaryKey: true}}}, wantErr: "field \"product_id\""},
		{name: "string id", model: &ModelDefinition{Name: "Category", Fields: []FieldDefinition{{Name: "id", Type: "String", GoType: "string", PrimaryKey: true}}}, wantErr: "non-integer primary key"},
		{name: "two primary keys", model: &ModelDefinition{Name: "Pair", Fields: []FieldDefinition{{Name: "id", Type: "Int64", GoType: "int64", PrimaryKey: true}, {Name: "other_id", Type: "Int64", GoType: "int64", PrimaryKey: true}}}, wantErr: "more than one primary key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIModels([]*ModelDefinition{tt.model})
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			if tt.name == "non-auto-increment int64 id" {
				assert.Contains(t, err.Error(), "API requires an auto-increment id")
			}
		})
	}
}

func TestGeneratorGenerate_APIModelRequiresConcreteWritableIntegerID(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))

	err := NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "model Product")
	assert.Contains(t, err.Error(), "no concrete int64 ID or Id field")
	assert.NoFileExists(t, filepath.Join(tmpDir, "api_gen.go"))
}

func TestGeneratorGenerate_RejectsConcreteIntID(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema; ID int }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))

	err := NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate()
	require.ErrorContains(t, err, "no concrete int64 ID or Id field")
	assert.NoFileExists(t, filepath.Join(tmpDir, "api_gen.go"))
}

func TestGeneratorGenerate_AcceptsModelWithIDMethods(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema; key int64 }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	methodSrc := `package testmodels
func (p *Product) GetID() int64 { return p.key }
func (p *Product) SetID(id int64) { p.key = id }
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "model_id.go"), []byte(methodSrc), 0644))
	require.NoError(t, NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate())
	assert.FileExists(t, filepath.Join(tmpDir, "api_gen.go"))
}

func TestGeneratorGenerate_IgnoresIDMethodsOutsideActiveProductionFiles(t *testing.T) {
	for _, tt := range []struct {
		name     string
		filename string
		source   string
	}{
		{name: "test file", filename: "model_id_test.go", source: `package testmodels
func (p *Product) GetID() int64 { return p.key }
func (p *Product) SetID(id int64) { p.key = id }
`},
		{name: "build excluded file", filename: "model_id_excluded.go", source: `//go:build ignore

package testmodels
func (p *Product) GetID() int64 { return p.key }
func (p *Product) SetID(id int64) { p.key = id }
`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema; key int64 }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, tt.filename), []byte(tt.source), 0644))

			err := NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "model Product")
			assert.Contains(t, err.Error(), "no concrete int64 ID or Id field")
		})
	}
}

func TestGeneratorGenerate_AcceptsPromotedModelWithIDMethods(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Identified struct { key int64 }
func (i *Identified) GetID() int64 { return i.key }
func (i *Identified) SetID(id int64) { i.key = id }
type Product struct { schema.BaseSchema; Identified }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	require.NoError(t, NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate())
	assert.FileExists(t, filepath.Join(tmpDir, "api_gen.go"))
}

func TestGeneratorGenerate_InvalidAPIModelPreservesGeneratedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Category struct { schema.BaseSchema }
func (Category) Fields() []schema.Field {
	return []schema.Field{schema.String("id").Primary().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	oldGen := []byte("old models output\n")
	oldAPI := []byte("old API output\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "gen.go"), oldGen, 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "api_gen.go"), oldAPI, 0644))

	err := NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate()
	require.Error(t, err)
	genContents, readErr := os.ReadFile(filepath.Join(tmpDir, "gen.go"))
	require.NoError(t, readErr)
	apiContents, readErr := os.ReadFile(filepath.Join(tmpDir, "api_gen.go"))
	require.NoError(t, readErr)
	assert.Equal(t, oldGen, genContents)
	assert.Equal(t, oldAPI, apiContents)
}

func TestGeneratorGenerate_APIWritePreparationFailurePreservesGen(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Product struct { schema.BaseSchema; ID int64 }
func (Product) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	oldGen := []byte("old models output\n")
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "gen.go"), oldGen, 0644))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "api_gen.go"), 0755))

	err := NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate()
	require.Error(t, err)
	genContents, readErr := os.ReadFile(filepath.Join(tmpDir, "gen.go"))
	require.NoError(t, readErr)
	assert.Equal(t, oldGen, genContents)
}

func TestGeneratorGenerate_PluralizesCategoryRoute(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type Category struct { schema.BaseSchema; ID int64 }
func (Category) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	require.NoError(t, NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate())

	contents, err := os.ReadFile(filepath.Join(tmpDir, "api_gen.go"))
	require.NoError(t, err)
	assert.Contains(t, string(contents), `apiRouter.Register("categories", NewCategoryViewSet())`)
}

func TestGeneratorGenerate_UsesKebabCasePluralRouteForCompoundModel(t *testing.T) {
	tmpDir := t.TempDir()
	modelSrc := `package testmodels
import "github.com/forgego/forge/schema"
type ProductVariant struct { schema.BaseSchema; ID int64 }
func (ProductVariant) Fields() []schema.Field {
	return []schema.Field{schema.Int64("id").Primary().AutoIncrement().Build()}
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "models.go"), []byte(modelSrc), 0644))
	require.NoError(t, NewGenerator(tmpDir, tmpDir).SetGenerateAPI(true).Generate())

	contents, err := os.ReadFile(filepath.Join(tmpDir, "api_gen.go"))
	require.NoError(t, err)
	assert.Contains(t, string(contents), `apiRouter.Register("product-variants", NewProductVariantViewSet())`)
	assert.NotContains(t, string(contents), "product_variants")
}
