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

// ASTParser parses Go AST to extract schema definitions
type ASTParser struct {
	fset *token.FileSet
}

// NewASTParser creates a new AST parser
func NewASTParser() *ASTParser {
	return &ASTParser{
		fset: token.NewFileSet(),
	}
}

// ParseDirectory parses all Go files in a directory and extracts schema definitions
func (p *ASTParser) ParseDirectory(dir string) ([]*ModelDefinition, error) {
	var definitions []*ModelDefinition

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip generated files
		if strings.HasSuffix(path, ".gen.go") || strings.HasSuffix(path, "gen.go") {
			return nil
		}

		defs, err := p.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		definitions = append(definitions, defs...)
		return nil
	})

	return definitions, err
}

// ParseFile parses a single Go file and extracts schema definitions
func (p *ASTParser) ParseFile(filename string) ([]*ModelDefinition, error) {
	node, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var definitions []*ModelDefinition
	packageName := node.Name.Name

	// Find all structs that embed schema.Schema
	ast.Inspect(node, func(n ast.Node) bool {
		// nolint:gocritic // singleCaseSwitch: switch is idiomatic for type assertions, allows easy extension
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}

					// Check if struct embeds schema.Schema
					if p.embedsSchema(st) {
						def, err := p.extractModelDefinition(packageName, ts, st, node)
						if err != nil {
							// Log error but continue
							fmt.Printf("Warning: failed to extract model definition for %s: %v\n", ts.Name.Name, err)
							return true
						}

						if def != nil {
							definitions = append(definitions, def)
						}
					}
				}
			}
		}
		return true
	})

	return definitions, nil
}

// embedsSchema checks if a struct embeds schema.Schema or a generated model base.
func (p *ASTParser) embedsSchema(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			// Embedded field
			if sel, ok := field.Type.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "schema" {
					if sel.Sel.Name == "Schema" || sel.Sel.Name == "BaseSchema" {
						return true
					}
				}
			}
			if ident, ok := field.Type.(*ast.Ident); ok {
				if strings.HasSuffix(ident.Name, "Generated") {
					return true
				}
			}
		}
	}
	return false
}

// extractModelDefinition extracts model definition from AST
func (p *ASTParser) extractModelDefinition(packageName string, typeSpec *ast.TypeSpec, structType *ast.StructType, file *ast.File) (*ModelDefinition, error) {
	modelName := typeSpec.Name.Name

	def := &ModelDefinition{
		Package:   packageName,
		Name:      modelName,
		Fields:    []FieldDefinition{},
		Relations: []RelationDefinition{},
		Meta:      MetaDefinition{},
		Hooks:     HooksDefinition{},
	}

	// Find Fields() method
	fieldsMethod := p.findMethod(file, modelName, "Fields")
	if fieldsMethod != nil {
		fields, err := p.extractFields(fieldsMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract fields: %w", err)
		}
		def.Fields = fields
	}

	// Find Relations() method
	relationsMethod := p.findMethod(file, modelName, "Relations")
	if relationsMethod != nil {
		relations, err := p.extractRelations(relationsMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract relations: %w", err)
		}
		def.Relations = relations
	}

	// Find Meta() method
	metaMethod := p.findMethod(file, modelName, "Meta")
	if metaMethod != nil {
		meta, err := p.extractMeta(metaMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract meta: %w", err)
		}
		def.Meta = meta
	}

	// Find Hooks() method
	hooksMethod := p.findMethod(file, modelName, "Hooks")
	if hooksMethod != nil {
		hooks, err := p.extractHooks(hooksMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract hooks: %w", err)
		}
		def.Hooks = hooks
	}

	return def, nil
}

// findMethod finds a method by name
func (p *ASTParser) findMethod(file *ast.File, receiverName, methodName string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Name.Name != methodName {
			continue
		}

		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}

		// Check receiver type
		recv := fn.Recv.List[0].Type
		if ident, ok := recv.(*ast.Ident); ok && ident.Name == receiverName {
			return fn
		}
		if star, ok := recv.(*ast.StarExpr); ok {
			if ident, ok := star.X.(*ast.Ident); ok && ident.Name == receiverName {
				return fn
			}
		}
	}
	return nil
}

func collectAssignedExprs(body *ast.BlockStmt) map[string]ast.Expr {
	assignedExprs := make(map[string]ast.Expr)
	if body == nil {
		return assignedExprs
	}

	for _, stmt := range body.List {
		switch s := stmt.(type) {
		case *ast.AssignStmt:
			if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
				if ident, ok := s.Lhs[0].(*ast.Ident); ok {
					assignedExprs[ident.Name] = s.Rhs[0]
				}
			}
		case *ast.DeclStmt:
			gen, ok := s.Decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if !ok || len(valSpec.Names) != 1 || len(valSpec.Values) != 1 {
					continue
				}
				assignedExprs[valSpec.Names[0].Name] = valSpec.Values[0]
			}
		}
	}
	return assignedExprs
}

func (p *ASTParser) resolveAssignedExpr(expr ast.Expr, assignedExprs map[string]ast.Expr) ast.Expr {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return expr
	}

	resolved, ok := assignedExprs[ident.Name]
	if !ok {
		return expr
	}
	return resolved
}

func (p *ASTParser) formatSelectorExpr(sel *ast.SelectorExpr) string {
	if prefix, ok := sel.X.(*ast.Ident); ok {
		return prefix.Name + "." + sel.Sel.Name
	}
	if nested, ok := sel.X.(*ast.SelectorExpr); ok {
		return p.formatSelectorExpr(nested) + "." + sel.Sel.Name
	}
	return sel.Sel.Name
}
