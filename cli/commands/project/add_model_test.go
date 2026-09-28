package project

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
)

// generateFieldCode must emit the functional constructors the schema package
// exports; the builder form (schema.String("x").Build()) does not compile.
func TestGenerateFieldCodeUsesFunctionalConstructors(t *testing.T) {
	cases := []struct {
		field FieldDefinition
		want  string
	}{
		{FieldDefinition{Name: "title", Type: "String", Required: true, MaxLength: 200},
			`schema.StringField("title", schema.MaxLength(200), schema.Required())`},
		{FieldDefinition{Name: "views", Type: "Int64", Default: "0"},
			`schema.Int64Field("views", schema.Default(0))`},
		{FieldDefinition{Name: "published", Type: "Bool"}, `schema.BoolField("published")`},
		{FieldDefinition{Name: "sent_at", Type: "Time", Unique: true}, `schema.TimeField("sent_at", schema.Unique())`},
		{FieldDefinition{Name: "price", Type: "Float64"}, `schema.Float64Field("price")`},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, generateFieldCode(tc.field))
	}
}

func TestGenerateModelCodeParses(t *testing.T) {
	code := "package blog\n\nimport \"github.com/forgego/forge/schema\"\n" +
		generateModelCode("blog", "Post", "posts", []FieldDefinition{{Name: "title", Type: "String", Required: true}})
	_, err := parser.ParseFile(token.NewFileSet(), "models.go", code, 0)
	require.NoError(t, err)
	require.Contains(t, code, `schema.Int64Field("id", schema.Primary(), schema.AutoIncrement())`)
	require.NotContains(t, code, ".Build()")
}
