package generator

import (
	"go/ast"

	"github.com/forgego/forge/schema"
)

// Keep constructor, option and embedding recognition consistent for ordinary
// schema imports and aliases. The canonical spelling remains supported by
// AST-only fixtures that intentionally omit import declarations.
func (p *ASTParser) isSchemaQualifier(name string) bool {
	return name == "schema" || p.schemaAliases[name]
}

// schemaConstantString handles only public constants whose values are known
// to this generator. Arbitrary package selectors still require diagnostics;
// accepting a familiar symbol from an unrelated import would lose that guard.
func (p *ASTParser) schemaConstantString(expr ast.Expr) (string, bool) {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok || !p.schemaAliases[qualifier.Name] {
		return "", false
	}
	switch selector.Sel.Name {
	case "ConstraintTypeCheck":
		return string(schema.ConstraintTypeCheck), true
	case "ConstraintTypeUnique":
		return string(schema.ConstraintTypeUnique), true
	case "ConstraintTypePrimaryKey":
		return string(schema.ConstraintTypePrimaryKey), true
	case "ConstraintTypeForeignKey":
		return string(schema.ConstraintTypeForeignKey), true
	case "ConstraintTypeExclusion":
		return string(schema.ConstraintTypeExclusion), true
	default:
		return "", false
	}
}
