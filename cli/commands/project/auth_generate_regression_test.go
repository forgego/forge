package project

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/forgego/forge/cli/core"
	codegen "github.com/forgego/forge/codegen"
	"github.com/stretchr/testify/require"
)

// countUserObjectsDecls counts top-level `UserObjects` var declarations in dir.
func countUserObjectsDecls(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	count := 0
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || len(entry.Name()) < 4 || entry.Name()[len(entry.Name())-3:] != ".go" {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		require.NoError(t, err)
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if name.Name == "UserObjects" {
						count++
					}
				}
			}
		}
	}
	return count
}

// TestAuthScaffoldGenerateProducesCompilablePackage ensures the documented
// workflow `forge auth` + `forge generate --models app --output app` leaves a
// compilable package with exactly one UserObjects declaration owned by the
// generated code.
func TestAuthScaffoldGenerateProducesCompilablePackage(t *testing.T) {
	forgePath, err := filepath.Abs("../../..")
	require.NoError(t, err)
	projectPath := t.TempDir()
	require.NoError(t, os.WriteFile(
		filepath.Join(projectPath, "go.mod"),
		[]byte("module example.com/authregen\n\ngo 1.23\n\nrequire github.com/forgego/forge v0.0.0\n\nreplace github.com/forgego/forge => "+forgePath+"\n"),
		0644,
	))

	originalWD, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalWD) })
	require.NoError(t, os.Chdir(projectPath))
	require.NoError(t, NewAuthCommand().Execute(&core.Context{}, nil))

	appPath := filepath.Join(projectPath, "app", "auth")
	gen := codegen.NewGenerator(appPath, appPath)
	require.NoError(t, gen.Generate(), "forge generate over the auth scaffold failed")

	require.Equal(t, 1, countUserObjectsDecls(t, appPath),
		"expected exactly one UserObjects declaration after generation")

	cmd := exec.Command("go", "build", "-mod=mod", "./...")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "scaffolded auth app does not compile after generation:\n%s", output)
}
