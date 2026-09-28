package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// extractRelations extracts relation definitions from Relations() method
func (p *ASTParser) extractRelations(method *ast.FuncDecl) ([]RelationDefinition, error) {
	var relations []RelationDefinition

	if method == nil || method.Body == nil {
		return relations, nil
	}

	assignedExprs := collectAssignedExprs(method.Body)
	processed := make(map[ast.Node]bool)

	// Process slice of relations
	processSlice := func(sliceLit *ast.CompositeLit) {
		for _, elt := range sliceLit.Elts {
			elt = p.resolveAssignedExpr(elt, assignedExprs)
			if processed[elt] {
				continue
			}
			processed[elt] = true
			relation := p.extractRelationFromExpr(elt)
			if relation != nil {
				relations = append(relations, *relation)
			}
		}
	}

	for _, stmt := range method.Body.List {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			for _, result := range s.Results {
				resExpr := p.resolveAssignedExpr(result, assignedExprs)
				if sliceLit, ok := resExpr.(*ast.CompositeLit); ok {
					processSlice(sliceLit)
				}
			}
		case *ast.DeclStmt:
			if gen, ok := s.Decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, val := range valSpec.Values {
							if compLit, ok := val.(*ast.CompositeLit); ok {
								processSlice(compLit)
							}
						}
					}
				}
			}
		case *ast.ExprStmt:
			if call, ok := s.X.(*ast.CallExpr); ok {
				if fnIdent, ok := call.Fun.(*ast.Ident); ok && fnIdent.Name == "append" && len(call.Args) > 1 {
					for _, arg := range call.Args[1:] {
						arg = p.resolveAssignedExpr(arg, assignedExprs)
						if processed[arg] {
							continue
						}
						processed[arg] = true
						if rel := p.extractRelationFromExpr(arg); rel != nil {
							relations = append(relations, *rel)
						}
					}
				}
			}
		case *ast.AssignStmt:
			for _, rhs := range s.Rhs {
				if compLit, ok := rhs.(*ast.CompositeLit); ok {
					processSlice(compLit)
				}
				if call, ok := rhs.(*ast.CallExpr); ok {
					if fnIdent, ok := call.Fun.(*ast.Ident); ok && fnIdent.Name == "append" && len(call.Args) > 1 {
						for _, arg := range call.Args[1:] {
							arg = p.resolveAssignedExpr(arg, assignedExprs)
							if processed[arg] {
								continue
							}
							processed[arg] = true
							if rel := p.extractRelationFromExpr(arg); rel != nil {
								relations = append(relations, *rel)
							}
						}
					}
				}
			}
		}
	}

	return relations, nil
}

// extractRelationFromExpr extracts a relation from an AST expression
func (p *ASTParser) extractRelationFromExpr(expr ast.Expr) *RelationDefinition {
	relation := &RelationDefinition{
		Options: make(map[string]interface{}),
	}

	// Check if it's a composite literal (struct literal)
	if compLit, ok := expr.(*ast.CompositeLit); ok {
		// Try to extract from struct literal
		for _, elt := range compLit.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if key, ok := kv.Key.(*ast.Ident); ok {
					switch key.Name {
					case "Name":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.Name = val
						}
					case "To":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.To = val
						}
					case "Type":
						if val := p.extractIntFromExpr(kv.Value); val != nil {
							relation.Type = fmt.Sprintf("%d", *val)
							relation.Options["type"] = *val
						}
					case "RelatedName":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.Options["related_name"] = val
						}
					case "OnDelete":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.Options["on_delete"] = val
						}
					case "OnUpdate":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.Options["on_update"] = val
						}
					case "Through":
						if val := p.extractStringFromExpr(kv.Value); val != "" {
							relation.Options["through"] = val
						}
					}
				}
			}
		}
	} else if callExpr, ok := expr.(*ast.CallExpr); ok {
		// Check if it's a builder call like schema.ForeignKey() or schema.ForeignKeyField()
		builderCall, builderType := p.findRelationBuilderInChain(callExpr)
		if builderCall != nil {
			relation.Type = builderType
			if len(builderCall.Args) >= 2 {
				if name := p.extractStringFromExpr(builderCall.Args[0]); name != "" {
					relation.Name = name
				}
				if to := p.extractStringFromExpr(builderCall.Args[1]); to != "" {
					relation.To = to
				}
			}
			// Extract builder chain options
			p.extractRelationOptionsFromChain(callExpr, relation.Options)
			// Extract functional variadic options (e.g. schema.OnDelete(...))
			p.extractRelationOptionsFromVariadicArgs(builderCall, relation.Options)
		}
	}

	if relation.Name == "" && relation.To == "" {
		return nil
	}

	return relation
}

// extractRelationOptionsFromVariadicArgs extracts options from variadic arguments of relation builders
func (p *ASTParser) extractRelationOptionsFromVariadicArgs(builderCall *ast.CallExpr, options map[string]interface{}) {
	if len(builderCall.Args) <= 2 {
		return
	}
	for _, arg := range builderCall.Args[2:] {
		if call, ok := arg.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				p.extractRelationOptionFromMethod(sel.Sel.Name, call, options)
			}
		}
	}
}

// findRelationBuilderInChain finds the relation builder call in a method chain
func (p *ASTParser) findRelationBuilderInChain(expr ast.Expr) (*ast.CallExpr, string) {
	switch x := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "schema" {
				builderType := sel.Sel.Name
				if strings.HasPrefix(builderType, "ForeignKey") ||
					strings.HasPrefix(builderType, "OneToOne") ||
					strings.HasPrefix(builderType, "ManyToMany") ||
					strings.HasPrefix(builderType, "OneToMany") {
					return x, builderType
				}
			}
			// Continue traversing down the chain
			return p.findRelationBuilderInChain(sel.X)
		}
		return p.findRelationBuilderInChain(x.Fun)
	case *ast.SelectorExpr:
		return p.findRelationBuilderInChain(x.X)
	}
	return nil, ""
}

// extractRelationOptionsFromChain extracts options from a relation builder chain
func (p *ASTParser) extractRelationOptionsFromChain(expr ast.Expr, options map[string]interface{}) {
	switch x := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			methodName := sel.Sel.Name

			// Skip Build() - it's just a finalizer
			if methodName != "Build" {
				p.extractRelationOptionFromMethod(methodName, x, options)
			}

			// Continue traversing down the chain
			p.extractRelationOptionsFromChain(sel.X, options)
		} else {
			p.extractRelationOptionsFromChain(x.Fun, options)
		}
	case *ast.SelectorExpr:
		p.extractRelationOptionsFromChain(x.X, options)
	}
}

// extractRelationOptionFromMethod extracts option value from a relation method call
func (p *ASTParser) extractRelationOptionFromMethod(methodName string, call *ast.CallExpr, options map[string]interface{}) {
	switch methodName {
	case "OnDelete":
		if len(call.Args) > 0 {
			if sel, ok := call.Args[0].(*ast.SelectorExpr); ok {
				// Handle schema.CascadeCASCADE, etc.
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "schema" {
					options["on_delete"] = strings.TrimPrefix(sel.Sel.Name, "Cascade")
				}
			} else if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["on_delete"] = strings.TrimPrefix(val, "Cascade")
			}
		}
	case "OnUpdate":
		if len(call.Args) > 0 {
			if sel, ok := call.Args[0].(*ast.SelectorExpr); ok {
				// Handle schema.CascadeCASCADE, etc.
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "schema" {
					options["on_update"] = strings.TrimPrefix(sel.Sel.Name, "Cascade")
				}
			} else if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["on_update"] = strings.TrimPrefix(val, "Cascade")
			}
		}
	case "RelatedName":
		if len(call.Args) > 0 {
			if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["related_name"] = val
			}
		}
	case "Through", "ThroughTable":
		if len(call.Args) > 0 {
			if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["through"] = val
			}
		}
	case "CascadeOnDelete":
		options["on_delete"] = "CASCADE"
	}
}
