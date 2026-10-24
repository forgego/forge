package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtectHandwrittenFiles(t *testing.T) {
	t.Run("refuses handwritten file without changing it", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "gen.go")
		before := []byte("package models\n// keep this code\n")
		require.NoError(t, os.WriteFile(path, before, 0644))
		err := NewWriter().writeBytes([]byte(generatedHeader+"\npackage models\n"), path)
		require.ErrorContains(t, err, path)
		require.ErrorContains(t, err, "rename or delete")
		require.ErrorContains(t, err, "non-generated file")
		after, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Equal(t, before, after)
	})

	t.Run("overwrites generated file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "gen.go")
		require.NoError(t, os.WriteFile(path, []byte(generatedHeader+"\nold\n"), 0644))
		want := []byte(generatedHeader + "\nnew\n")
		require.NoError(t, NewWriter().writeBytes(want, path))
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("creates missing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "gen.go")
		want := []byte(generatedHeader + "\npackage models\n")
		require.NoError(t, NewWriter().writeBytes(want, path))
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

func TestDeterministicGeneration(t *testing.T) {
	definitions := []*ModelDefinition{
		{Package: "catalog", Name: "Product", Meta: MetaDefinition{TableName: "products"}, Fields: []FieldDefinition{
			{Name: "id", Type: "Int64Field", GoType: "int64", PrimaryKey: true, AutoIncrement: true},
			{Name: "name", Type: "StringField", GoType: "string"},
		}},
		{Package: "catalog", Name: "Category", Meta: MetaDefinition{TableName: "categories"}, Fields: []FieldDefinition{
			{Name: "id", Type: "Int64Field", GoType: "int64", PrimaryKey: true, AutoIncrement: true},
			{Name: "title", Type: "StringField", GoType: "string"},
		}},
	}

	var baseline []byte
	for i := 0; i < 100; i++ {
		dir := t.TempDir()
		require.NoError(t, NewWriter().WriteCombined(definitions, dir))
		got, err := os.ReadFile(filepath.Join(dir, "gen.go"))
		require.NoError(t, err)
		if i == 0 {
			baseline = got
			continue
		}
		require.Equalf(t, baseline, got, "generation %d differs", i+1)
	}
}
