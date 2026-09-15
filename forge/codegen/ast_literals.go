package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// extractStringArg extracts a string argument from a function call
func (p *ASTParser) extractStringArg(call *ast.CallExpr) string {
	if len(call.Args) == 0 {
		return ""
	}

	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return strings.Trim(lit.Value, `"`)
	}

	return ""
}

// extractIntArg extracts an integer argument from a function call
func (p *ASTParser) extractIntArg(call *ast.CallExpr) *int {
	if len(call.Args) == 0 {
		return nil
	}

	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.INT {
		var val int
		if _, err := fmt.Sscanf(lit.Value, "%d", &val); err == nil {
			return &val
		}
	}

	return nil
}

// extractFloatArg extracts a float argument from a function call
func (p *ASTParser) extractFloatArg(call *ast.CallExpr) *float64 {
	if len(call.Args) == 0 {
		return nil
	}

	if lit, ok := call.Args[0].(*ast.BasicLit); ok && (lit.Kind == token.FLOAT || lit.Kind == token.INT) {
		var val float64
		if _, err := fmt.Sscanf(lit.Value, "%f", &val); err == nil {
			return &val
		}
	}

	return nil
}

// extractBoolArg extracts a boolean argument from a function call
func (p *ASTParser) extractBoolArg(call *ast.CallExpr) *bool {
	return p.extractBoolArgAt(call, 0)
}

// extractBoolArgAt extracts a boolean argument at the given index from a function call
func (p *ASTParser) extractBoolArgAt(call *ast.CallExpr, index int) *bool {
	if len(call.Args) <= index {
		return nil
	}

	if ident, ok := call.Args[index].(*ast.Ident); ok {
		if ident.Name == "true" {
			val := true
			return &val
		} else if ident.Name == "false" {
			val := false
			return &val
		}
	}

	return nil
}

// extractStringFromExpr extracts a string value from an AST expression
func (p *ASTParser) extractStringFromExpr(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		// Remove quotes
		if len(lit.Value) >= 2 {
			return lit.Value[1 : len(lit.Value)-1]
		}
	}
	return ""
}

// extractIntFromExpr extracts an int value from an AST expression
func (p *ASTParser) extractIntFromExpr(expr ast.Expr) *int {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		var val int
		if _, err := fmt.Sscanf(lit.Value, "%d", &val); err == nil {
			return &val
		}
	}
	return nil
}

// extractBoolFromExpr extracts a bool value from an AST expression
func (p *ASTParser) extractBoolFromExpr(expr ast.Expr) *bool {
	if ident, ok := expr.(*ast.Ident); ok {
		if ident.Name == "true" {
			val := true
			return &val
		}
		if ident.Name == "false" {
			val := false
			return &val
		}
	}
	return nil
}
