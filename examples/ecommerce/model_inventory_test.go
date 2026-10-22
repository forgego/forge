package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

// The ecommerce reference app contains the documented app and model inventory.
func TestEcommerceContains57ModelsAcrossTenApps(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("app", "*", "models.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 10 {
		t.Fatalf("apps=%d, want 10", len(files))
	}
	models := 0
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			group, ok := decl.(*ast.GenDecl)
			if !ok || group.Tok != token.TYPE {
				continue
			}
			for _, spec := range group.Specs {
				definition, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := definition.Type.(*ast.StructType); ok {
					models++
				}
			}
		}
	}
	if models != 57 {
		t.Fatalf("models=%d, want 57", models)
	}
}
