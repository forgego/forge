package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

const foreignKeyModelsSrc = `package models

import "github.com/forgego/forge/schema"

type Project struct{ schema.BaseSchema }

func (Project) Fields() []schema.Field {
	return []schema.Field{schema.Int64Field("id", schema.Primary(), schema.AutoIncrement())}
}

func (Project) Meta() schema.Meta { return schema.Meta{TableName: "projects"} }

type Task struct{ schema.BaseSchema }

func (Task) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("project_id", schema.Required()),
	}
}

func (Task) Meta() schema.Meta { return schema.Meta{TableName: "tasks"} }

func (Task) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("project_id", "Project", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
`

// TestDetector_ParsedRelationsAddForeignKeys parses a model that declares a
// foreign key with schema.ForeignKeyField. The new table's migration must add
// the constraint; the parser used to record the relation as "ForeignKeyField",
// which the detector ignored, so no constraint was generated.
func TestDetector_ParsedRelationsAddForeignKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.go")
	require.NoError(t, os.WriteFile(path, []byte(foreignKeyModelsSrc), 0o644))
	defs, err := generator.NewASTParser().ParseFile(path)
	require.NoError(t, err)

	changes, err := NewDetector().DetectChanges(defs, nil)
	require.NoError(t, err)

	var fks []*core.AddForeignKey
	for _, change := range changes {
		if fk, ok := change.(*core.AddForeignKey); ok {
			fks = append(fks, fk)
		}
	}
	require.Len(t, fks, 1)
	assert.Equal(t, "tasks", fks[0].Table)
	assert.Equal(t, "project_id", fks[0].Relation.Name)
	assert.Equal(t, "projects", fks[0].TargetTable)
}
