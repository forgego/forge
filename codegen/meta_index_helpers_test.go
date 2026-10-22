package generator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMetaIndexHelpersAreParsedWithAliases(t *testing.T) {
	for _, qualifier := range []string{"schema", "s"} {
		t.Run(qualifier, func(t *testing.T) {
			alias := ""
			if qualifier == "s" {
				alias = "s "
			}
			source := `package models
import ` + alias + `"github.com/forgego/forge/schema"
type Item struct{` + qualifier + `.BaseSchema}
func(Item)Fields()[]` + qualifier + `.Field{return []` + qualifier + `.Field{` + qualifier + `.Int64Field("id",` + qualifier + `.Primary()),` + qualifier + `.StringField("name")}}
func(Item)Meta()` + qualifier + `.Meta{return ` + qualifier + `.Meta{TableName:"items",Indexes:[]` + qualifier + `.Index{
 ` + qualifier + `.IndexOn("idx_name","name"),
 ` + qualifier + `.UniqueIndexOn("idx_unique_name","name","id"),
}}}
`
			path := filepath.Join(t.TempDir(), "models.go")
			require.NoError(t, os.WriteFile(path, []byte(source), 0600))
			parser := NewASTParser()
			definitions, err := parser.ParseFile(path)
			require.NoError(t, err)
			require.Empty(t, parser.Diagnostics())
			require.Len(t, definitions, 1)
			require.Len(t, definitions[0].Meta.Indexes, 2)
			assert.Equal(t, IndexDefinition{Name: "idx_name", Fields: []string{"name"}}, definitions[0].Meta.Indexes[0])
			assert.Equal(t, IndexDefinition{Name: "idx_unique_name", Fields: []string{"name", "id"}, Unique: true}, definitions[0].Meta.Indexes[1])
			dynamic := strings.Replace(source, `"idx_name","name"`, `"idx_name",fieldName()`, 1)
			require.NoError(t, os.WriteFile(path, []byte(dynamic), 0600))
			_, err = parser.ParseFile(path)
			require.NoError(t, err)
			require.NotEmpty(t, parser.Diagnostics(), "a helper is supported only when every argument is statically understood")
			wrongSlot := strings.Replace(source, `TableName:"items"`, `TableName:`+qualifier+`.IndexOn("not_a_table","name")`, 1)
			require.NoError(t, os.WriteFile(path, []byte(wrongSlot), 0600))
			wrongParser := NewASTParser()
			_, err = wrongParser.ParseFile(path)
			require.NoError(t, err)
			require.NotEmpty(t, wrongParser.Diagnostics(), "index constructors are not string-valued Meta expressions")
			spoofed := strings.Replace(source, alias+`"github.com/forgego/forge/schema"`, qualifier+` "example.com/unrelated"`, 1)
			require.NoError(t, os.WriteFile(path, []byte(spoofed), 0600))
			spoofedParser := NewASTParser()
			spoofedDefinitions, err := spoofedParser.ParseFile(path)
			require.NoError(t, err)
			// A non-schema alias may make the entire model unrecognized. If
			// a legacy canonical qualifier is recognized, its helper must
			// still be diagnosed and never interpreted as a schema index.
			for _, definition := range spoofedDefinitions {
				require.Empty(t, definition.Meta.Indexes)
			}
			if len(spoofedDefinitions) > 0 {
				require.NotEmpty(t, spoofedParser.Diagnostics(), "only actual schema imports may use static constructor semantics")
			}
		})
	}
}
