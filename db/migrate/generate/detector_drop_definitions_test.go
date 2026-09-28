package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

func makePreviousDefs() []*generator.ModelDefinition {
	return []*generator.ModelDefinition{
		{
			Name: "User",
			Meta: generator.MetaDefinition{TableName: "users"},
			Fields: []generator.FieldDefinition{
				{Name: "id", Type: "Int64", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "String"},
				{Name: "email", Type: "String"},
			},
		},
		{
			Name: "Post",
			Meta: generator.MetaDefinition{TableName: "posts"},
			Fields: []generator.FieldDefinition{
				{Name: "id", Type: "Int64", PrimaryKey: true, AutoIncrement: true},
				{Name: "title", Type: "String"},
			},
		},
	}
}

func makeCurrentDefs() []*generator.ModelDefinition {
	return []*generator.ModelDefinition{
		{
			Name: "User",
			Meta: generator.MetaDefinition{TableName: "users"},
			Fields: []generator.FieldDefinition{
				{Name: "id", Type: "Int64", PrimaryKey: true, AutoIncrement: true},
				{Name: "name", Type: "String"},
			},
		},
	}
}

func TestDetector_DropTableAndColumnRetainPreviousDefinitions(t *testing.T) {
	detector := NewDetector()
	changes, err := detector.DetectChanges(makeCurrentDefs(), makePreviousDefs())
	require.NoError(t, err)

	var dropTable *core.DropTable
	var dropColumn *core.DropColumn
	for _, change := range changes {
		switch c := change.(type) {
		case *core.DropTable:
			if c.Table == "posts" {
				dropTable = c
			}
		case *core.DropColumn:
			if c.Table == "users" && c.ColumnName == "email" {
				dropColumn = c
			}
		}
	}

	require.NotNil(t, dropTable, "expected DropTable change for posts")
	require.NotNil(t, dropTable.Definition, "DropTable must carry previous table definition")
	assert.Equal(t, "Post", dropTable.Definition.Name)
	assert.Equal(t, "posts", dropTable.Definition.Meta.TableName)

	require.NotNil(t, dropColumn, "expected DropColumn change for email")
	require.NotNil(t, dropColumn.Column, "DropColumn must carry previous column definition")
	assert.Equal(t, "email", dropColumn.Column.Name)
	assert.Equal(t, "String", dropColumn.Column.Type)
}
