package generator

import (
	"go/ast"
	"go/token"
)

// extractHooks extracts hooks definition from Hooks() method
func (p *ASTParser) extractHooks(method *ast.FuncDecl) (HooksDefinition, error) {
	hooks := HooksDefinition{}

	if method == nil || method.Body == nil {
		return hooks, nil
	}

	assignedExprs := collectAssignedExprs(method.Body)

	for _, stmt := range method.Body.List {
		if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
			for _, result := range retStmt.Results {
				p.extractHooksFromExpr(p.resolveAssignedExpr(result, assignedExprs), &hooks)
			}
		}
	}

	return hooks, nil
}

func (p *ASTParser) extractHooksFromExpr(expr ast.Expr, hooks *HooksDefinition) {
	switch x := expr.(type) {
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			p.extractHooksFromExpr(x.X, hooks)
		}
	case *ast.CompositeLit:
		p.extractHooksFromCompositeLiteral(x, hooks)
	case *ast.CallExpr:
		p.extractHooksFromCallChain(x, hooks)
	case *ast.ParenExpr:
		p.extractHooksFromExpr(x.X, hooks)
	}
}

func (p *ASTParser) extractHooksFromCompositeLiteral(compLit *ast.CompositeLit, hooks *HooksDefinition) {
	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		p.setHookValue(hooks, keyIdent.Name, p.extractHookReference(kv.Value))
	}
}

func (p *ASTParser) extractHooksFromCallChain(call *ast.CallExpr, hooks *HooksDefinition) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	p.setHookFromBuilderMethod(hooks, sel.Sel.Name, call)
	p.extractHooksFromExpr(sel.X, hooks)
}

func (p *ASTParser) setHookFromBuilderMethod(hooks *HooksDefinition, methodName string, call *ast.CallExpr) {
	if len(call.Args) == 0 {
		return
	}

	switch methodName {
	case "WithBeforeCreate":
		hooks.BeforeCreate = p.extractHookReference(call.Args[0])
	case "WithAfterCreate":
		hooks.AfterCreate = p.extractHookReference(call.Args[0])
	case "WithBeforeUpdate":
		hooks.BeforeUpdate = p.extractHookReference(call.Args[0])
	case "WithAfterUpdate":
		hooks.AfterUpdate = p.extractHookReference(call.Args[0])
	case "WithBeforeSave":
		hooks.BeforeSave = p.extractHookReference(call.Args[0])
	case "WithAfterSave":
		hooks.AfterSave = p.extractHookReference(call.Args[0])
	case "WithBeforeDelete":
		hooks.BeforeDelete = p.extractHookReference(call.Args[0])
	case "WithAfterDelete":
		hooks.AfterDelete = p.extractHookReference(call.Args[0])
	case "WithClean":
		hooks.Clean = p.extractHookReference(call.Args[0])
	}
}

func (p *ASTParser) setHookValue(hooks *HooksDefinition, hookName, value string) {
	switch hookName {
	case "BeforeCreate":
		hooks.BeforeCreate = value
	case "AfterCreate":
		hooks.AfterCreate = value
	case "BeforeUpdate":
		hooks.BeforeUpdate = value
	case "AfterUpdate":
		hooks.AfterUpdate = value
	case "BeforeSave":
		hooks.BeforeSave = value
	case "AfterSave":
		hooks.AfterSave = value
	case "BeforeDelete":
		hooks.BeforeDelete = value
	case "AfterDelete":
		hooks.AfterDelete = value
	case "Clean":
		hooks.Clean = value
	}
}

func (p *ASTParser) extractHookReference(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		if x.Name == "nil" {
			return ""
		}
		return x.Name
	case *ast.SelectorExpr:
		return p.formatSelectorExpr(x)
	case *ast.FuncLit:
		return "<inline>"
	case *ast.UnaryExpr:
		return p.extractHookReference(x.X)
	}
	return ""
}
