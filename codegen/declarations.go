package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// DeclaredNames returns package-level Go types and functions in dir. Named
// generated files can be excluded when checking what handwritten code owns.
func DeclaredNames(dir string, skipFiles ...string) (map[string]bool, error) {
	declared := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return declared, nil
	}
	if err != nil {
		return nil, err
	}
	skip := make(map[string]bool, len(skipFiles))
	for _, name := range skipFiles {
		skip[name] = true
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || skip[name] {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if t, ok := spec.(*ast.TypeSpec); ok {
						declared[t.Name.Name] = true
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					declared[d.Name.Name] = true
				}
			}
		}
	}
	return declared, nil
}
