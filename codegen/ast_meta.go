package generator

import (
	"go/ast"
	"go/token"
)

// extractMeta extracts meta definition from Meta() method
func (p *ASTParser) extractMeta(method *ast.FuncDecl) (MetaDefinition, error) {
	meta := MetaDefinition{}

	if method == nil || method.Body == nil {
		return meta, nil
	}

	assignedExprs := collectAssignedExprs(method.Body)

	// Find return statement
	for _, stmt := range method.Body.List {
		if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
			if len(retStmt.Results) > 0 {
				res := p.resolveAssignedExpr(retStmt.Results[0], assignedExprs)
				if unary, ok := res.(*ast.UnaryExpr); ok && unary.Op == token.AND {
					res = unary.X
				}
				// Get the composite literal (struct literal)
				if compLit, ok := res.(*ast.CompositeLit); ok {
					// Extract struct fields
					for _, elt := range compLit.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok {
								switch key.Name {
								case "TableName":
									meta.TableName = p.extractStringFromExpr(kv.Value)
								case "VerboseName":
									meta.VerboseName = p.extractStringFromExpr(kv.Value)
								case "VerboseNamePlural":
									meta.VerboseNamePlural = p.extractStringFromExpr(kv.Value)
								case "OrderBy":
									meta.OrderBy = p.extractStringSliceFromExpr(kv.Value)
								case "Indexes":
									meta.Indexes = p.extractIndexesFromExpr(kv.Value)
								case "Constraints":
									meta.Constraints = p.extractConstraintsFromExpr(kv.Value)
								case "UniqueTogether":
									meta.UniqueTogether = p.extractUniqueTogetherFromExpr(kv.Value)
								case "AppLabel":
									if val := p.extractStringFromExpr(kv.Value); val != "" {
										meta.AppLabel = val
									}
								case "Proxy":
									if val := p.extractBoolFromExpr(kv.Value); val != nil {
										meta.Proxy = *val
									}
								case "Abstract":
									if val := p.extractBoolFromExpr(kv.Value); val != nil {
										meta.Abstract = *val
									}
								case "Managed":
									if val := p.extractBoolFromExpr(kv.Value); val != nil {
										meta.Managed = *val
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return meta, nil
}

// extractStringSliceFromExpr extracts a []string from an AST expression
func (p *ASTParser) extractStringSliceFromExpr(expr ast.Expr) []string {
	var result []string
	if sliceLit, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range sliceLit.Elts {
			if str := p.extractStringFromExpr(elt); str != "" {
				result = append(result, str)
			}
		}
	}
	return result
}

// extractIndexesFromExpr extracts []Index from an AST expression
func (p *ASTParser) extractIndexesFromExpr(expr ast.Expr) []IndexDefinition {
	var indexes []IndexDefinition
	if sliceLit, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range sliceLit.Elts {
			if compLit, ok := elt.(*ast.CompositeLit); ok {
				idx := IndexDefinition{}
				for _, kv := range compLit.Elts {
					if keyVal, ok := kv.(*ast.KeyValueExpr); ok {
						if key, ok := keyVal.Key.(*ast.Ident); ok {
							switch key.Name {
							case "Name":
								idx.Name = p.extractStringFromExpr(keyVal.Value)
							case "Fields":
								idx.Fields = p.extractStringSliceFromExpr(keyVal.Value)
							case "Unique":
								if val := p.extractBoolFromExpr(keyVal.Value); val != nil {
									idx.Unique = *val
								}
							}
						}
					}
				}
				if idx.Name != "" || len(idx.Fields) > 0 {
					indexes = append(indexes, idx)
				}
			}
		}
	}
	return indexes
}

// extractConstraintsFromExpr extracts []Constraint from an AST expression
func (p *ASTParser) extractConstraintsFromExpr(expr ast.Expr) []ConstraintDefinition {
	var constraints []ConstraintDefinition
	if sliceLit, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range sliceLit.Elts {
			if compLit, ok := elt.(*ast.CompositeLit); ok {
				constr := ConstraintDefinition{}
				for _, kv := range compLit.Elts {
					if keyVal, ok := kv.(*ast.KeyValueExpr); ok {
						if key, ok := keyVal.Key.(*ast.Ident); ok {
							switch key.Name {
							case "Name":
								constr.Name = p.extractStringFromExpr(keyVal.Value)
							case "Type":
								constr.Type = p.extractStringFromExpr(keyVal.Value)
							case "Condition":
								constr.Condition = p.extractStringFromExpr(keyVal.Value)
							case "Fields":
								constr.Fields = p.extractStringSliceFromExpr(keyVal.Value)
							}
						}
					}
				}
				if constr.Name != "" || constr.Type != "" {
					constraints = append(constraints, constr)
				}
			}
		}
	}
	return constraints
}

// extractUniqueTogetherFromExpr extracts [][]string from an AST expression
func (p *ASTParser) extractUniqueTogetherFromExpr(expr ast.Expr) [][]string {
	var result [][]string
	if sliceLit, ok := expr.(*ast.CompositeLit); ok {
		for _, elt := range sliceLit.Elts {
			if innerSlice := p.extractStringSliceFromExpr(elt); len(innerSlice) > 0 {
				result = append(result, innerSlice)
			}
		}
	}
	return result
}
