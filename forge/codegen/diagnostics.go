package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

// Diagnostic reports a model expression that code generation cannot evaluate.
type Diagnostic struct {
	File    string // absolute or as-parsed path
	Line    int
	Column  int
	Model   string // model type name, empty if unknown
	Method  string // "Fields", "Relations" or "Meta"
	Message string
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s:%d:%d: %s.%s: %s", d.File, d.Line, d.Column, d.Model, d.Method, d.Message)
}

// Diagnostics returns parser diagnostics in source order.
func (p *ASTParser) Diagnostics() []Diagnostic {
	diagnostics := append([]Diagnostic(nil), p.diagnostics...)
	sort.Slice(diagnostics, func(i, j int) bool {
		if diagnostics[i].File != diagnostics[j].File {
			return diagnostics[i].File < diagnostics[j].File
		}
		if diagnostics[i].Line != diagnostics[j].Line {
			return diagnostics[i].Line < diagnostics[j].Line
		}
		return diagnostics[i].Column < diagnostics[j].Column
	})
	return diagnostics
}

func (p *ASTParser) reportDiagnostic(pos token.Pos, model, method, message string) {
	position := p.fset.Position(pos)
	p.diagnostics = append(p.diagnostics, Diagnostic{File: position.Filename, Line: position.Line, Column: position.Column, Model: model, Method: method, Message: message})
}

func (p *ASTParser) reportUnsupportedExpressions(model, methodName string, method *ast.FuncDecl) {
	if method.Body == nil {
		return
	}
	assigned := collectAssignedExprs(method.Body)
	candidates := collectCandidateVars(method.Body, assigned)
	seenSlices := make(map[*ast.CompositeLit]bool)
	seenCalls := make(map[*ast.CallExpr]bool)
	seenAppends := make(map[*ast.CallExpr]bool)
	for _, stmt := range method.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			if branchContributes(stmt, assigned, candidates) {
				p.reportDiagnostic(stmt.Pos(), model, methodName, "loops cannot be evaluated during generation; list each field explicitly")
			}
		case *ast.IfStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			if branchContributes(stmt, assigned, candidates) {
				p.reportDiagnostic(stmt.Pos(), model, methodName, "conditional assembly cannot be evaluated during generation; list each field explicitly")
			}
		}

		for _, expr := range methodExpressions(stmt) {
			resolved := p.resolveAssignedExpr(expr, assigned)
			if methodName == "Meta" {
				if comp, ok := resolved.(*ast.CompositeLit); ok {
					if !seenSlices[comp] {
						seenSlices[comp] = true
						p.reportUnsupportedMetaValues(comp, assigned, model, methodName, seenCalls)
					}
					continue
				}
				if call, ok := resolved.(*ast.CallExpr); ok {
					if isAppend(call) {
						if !seenAppends[call] {
							seenAppends[call] = true
							p.reportUnsupportedAppend(call, assigned, model, methodName, seenCalls)
						}
					} else if _, ok := stmt.(*ast.ReturnStmt); ok {
						p.reportHelperCallOnce(call, model, methodName, seenCalls)
					}
					continue
				}
				// Meta returning an identifier that resolves to nothing static is dropped.
				if ident, ok := expr.(*ast.Ident); ok {
					if _, ok := assigned[ident.Name]; !ok && !isBuiltinIdent(ident.Name) {
						if _, ok := resolved.(*ast.BasicLit); !ok {
							p.reportDiagnostic(expr.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", ident.Name))
						}
					}
				}
				continue
			}
			if slice, ok := resolved.(*ast.CompositeLit); ok && !seenSlices[slice] {
				seenSlices[slice] = true
				p.reportUnsupportedSliceElements(slice, assigned, model, methodName, seenCalls)
			}
			if call, ok := resolved.(*ast.CallExpr); ok {
				if isAppend(call) {
					if seenAppends[call] {
						continue
					}
					seenAppends[call] = true
					p.reportUnsupportedAppend(call, assigned, model, methodName, seenCalls)
				} else if _, ok := stmt.(*ast.ReturnStmt); ok {
					if !p.isSchemaConstructor(methodName, call) {
						p.reportHelperCallOnce(call, model, methodName, seenCalls)
					} else {
						p.reportNestedConstructorHelpers(call, assigned, model, methodName, seenCalls)
					}
				} else if direct, ok := expr.(*ast.CallExpr); ok && direct == call && !isAppend(direct) {
					// Direct non-returned constructor calls (e.g. nested) are
					// covered through slice/append traversal; nothing to do.
					_ = direct
				}
				continue
			}
			// Unresolved non-call returned directly (e.g. return nameField).
			if ident, ok := expr.(*ast.Ident); ok {
				if _, ok := assigned[ident.Name]; !ok && !isBuiltinIdent(ident.Name) {
					p.reportDiagnostic(expr.Pos(), model, methodName, fmt.Sprintf("field element %s cannot be evaluated during generation; declare fields with direct schema constructor calls", ident.Name))
				}
			}
		}
	}
}

func methodExpressions(stmt ast.Stmt) []ast.Expr {
	switch stmt := stmt.(type) {
	case *ast.ReturnStmt:
		return stmt.Results
	case *ast.AssignStmt:
		return stmt.Rhs
	case *ast.DeclStmt:
		if decl, ok := stmt.Decl.(*ast.GenDecl); ok && decl.Tok == token.VAR {
			var expressions []ast.Expr
			for _, spec := range decl.Specs {
				if value, ok := spec.(*ast.ValueSpec); ok {
					expressions = append(expressions, value.Values...)
				}
			}
			return expressions
		}
	case *ast.ExprStmt:
		return []ast.Expr{stmt.X}
	}
	return nil
}

func collectCandidateVars(body *ast.BlockStmt, assigned map[string]ast.Expr) map[string]bool {
	candidates := make(map[string]bool)
	for name := range assigned {
		candidates[name] = true
	}
	if body == nil {
		return candidates
	}
	for _, stmt := range body.List {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			for _, res := range ret.Results {
				if ident, ok := res.(*ast.Ident); ok {
					candidates[ident.Name] = true
				}
			}
		}
	}
	return candidates
}

// branchContributes reports whether a loop or conditional body writes to (or
// returns from) the schema collection held in candidate variables.
func branchContributes(stmt ast.Stmt, assigned map[string]ast.Expr, candidates map[string]bool) bool {
	contributes := false
	ast.Inspect(stmt, func(n ast.Node) bool {
		if contributes {
			return false
		}
		switch x := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt, *ast.IfStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			// Skip the outermost node itself; inspect its body.
			return true
		case *ast.CallExpr:
			if isAppend(x) && len(x.Args) > 0 {
				if ident, ok := x.Args[0].(*ast.Ident); ok && candidates[ident.Name] {
					contributes = true
					return false
				}
			}
		case *ast.AssignStmt:
			for _, lhs := range x.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && candidates[ident.Name] {
					contributes = true
					return false
				}
			}
		case *ast.ReturnStmt:
			contributes = true
			return false
		case *ast.DeclStmt:
			if gen, ok := x.Decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							if candidates[name.Name] {
								contributes = true
								return false
							}
						}
					}
				}
			}
		}
		return true
	})
	return contributes
}

func (p *ASTParser) reportUnsupportedSliceElements(slice *ast.CompositeLit, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool) {
	for _, element := range slice.Elts {
		if name, ok := p.computedValue(element, assigned, methodName); ok {
			p.reportDiagnostic(element.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", name))
			continue
		}
		resolved := p.resolveAssignedExpr(element, assigned)
		if call, ok := resolved.(*ast.CallExpr); ok {
			if p.isSchemaConstructor(methodName, call) {
				p.reportNestedConstructorHelpers(call, assigned, model, methodName, seenCalls)
				continue
			}
			p.reportHelperCallOnce(call, model, methodName, seenCalls)
			continue
		}
		if p.isSupportedElement(methodName, resolved) {
			continue
		}
		p.reportDiagnostic(element.Pos(), model, methodName, fmt.Sprintf("field element %s cannot be evaluated during generation; declare fields with direct schema constructor calls", expressionName(element)))
	}
}

func (p *ASTParser) reportUnsupportedAppend(call *ast.CallExpr, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool) {
	for _, arg := range call.Args[1:] {
		resolved := p.resolveAssignedExpr(arg, assigned)
		if nested, ok := resolved.(*ast.CallExpr); ok && p.isSchemaConstructor(methodName, nested) {
			p.reportNestedConstructorHelpers(nested, assigned, model, methodName, seenCalls)
			continue
		}
		if call.Ellipsis.IsValid() || p.isComputedSlice(arg, assigned, methodName) {
			// A direct schema constructor resolved through an assignment is
			// handled by the extractor, so it is not a computed slice.
			if nested, ok := resolved.(*ast.CallExpr); ok && p.isSchemaConstructor(methodName, nested) {
				p.reportNestedConstructorHelpers(nested, assigned, model, methodName, seenCalls)
				continue
			}
			p.reportDiagnostic(arg.Pos(), model, methodName, "append of a computed slice cannot be evaluated during generation")
			continue
		}
		if nested, ok := resolved.(*ast.CallExpr); ok && !p.isSchemaConstructor(methodName, nested) {
			p.reportHelperCallOnce(nested, model, methodName, seenCalls)
			continue
		}
		if p.isSupportedElement(methodName, resolved) {
			continue
		}
		p.reportDiagnostic(arg.Pos(), model, methodName, fmt.Sprintf("field element %s cannot be evaluated during generation; declare fields with direct schema constructor calls", expressionName(arg)))
	}
}

// reportUnsupportedMetaValues inspects the values of Meta key/value entries.
// The extractor only understands literals, so any helper call (or unresolved
// value) used for a Meta option is silently dropped without this diagnostic.
func (p *ASTParser) reportUnsupportedMetaValues(comp *ast.CompositeLit, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool) {
	for _, elt := range comp.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		p.reportUnsupportedMetaValue(kv.Value, assigned, model, methodName, seenCalls)
	}
}

func (p *ASTParser) reportUnsupportedMetaValue(expr ast.Expr, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool) {
	resolved := p.resolveAssignedExpr(expr, assigned)
	switch value := resolved.(type) {
	case *ast.BasicLit:
		return
	case *ast.CallExpr:
		p.reportHelperCallOnce(value, model, methodName, seenCalls)
		return
	case *ast.Ident:
		if isBuiltinIdent(value.Name) {
			return
		}
		// A plain identifier (constant or variable) is not evaluated by
		// extractMeta, so it is dropped silently.
		p.reportDiagnostic(expr.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", value.Name))
		return
	case *ast.CompositeLit:
		for _, elt := range value.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				p.reportUnsupportedMetaValue(kv.Value, assigned, model, methodName, seenCalls)
				continue
			}
			p.reportUnsupportedMetaValue(elt, assigned, model, methodName, seenCalls)
		}
		return
	case *ast.SelectorExpr:
		// Qualified constants such as schema.Cascade are not evaluated for
		// Meta string fields; literals are the only supported form. Leaf
		// selector values are reported as computed to avoid silent drops.
		// Unary and other wrappers are handled by the default branch below.
		p.reportDiagnostic(expr.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", expressionName(expr)))
		return
	default:
		ast.Inspect(resolved, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				p.reportHelperCallOnce(call, model, methodName, seenCalls)
				return false
			}
			return true
		})
	}
}

// reportNestedConstructorHelpers traverses constructor arguments, including
// chained option arguments, and reports helper calls whose values the
// extractor drops.
func (p *ASTParser) reportNestedConstructorHelpers(call *ast.CallExpr, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool) {
	visited := make(map[*ast.CallExpr]bool)
	var visitChain func(expr ast.Expr)
	visitChain = func(expr ast.Expr) {
		callExpr, ok := expr.(*ast.CallExpr)
		if !ok {
			if sel, ok := expr.(*ast.SelectorExpr); ok {
				visitChain(sel.X)
			}
			return
		}
		if visited[callExpr] {
			return
		}
		visited[callExpr] = true
		method, isSchemaPkg := chainMethod(callExpr)
		start := 0
		if isSchemaPkg && p.isFieldBuilder(method) {
			if len(callExpr.Args) > 0 {
				p.reportConstructorArg(callExpr.Args[0], assigned, model, methodName, seenCalls, true)
			}
			start = 1
		}
		for i := start; i < len(callExpr.Args); i++ {
			p.reportConstructorArg(callExpr.Args[i], assigned, model, methodName, seenCalls, false)
		}
		if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			visitChain(sel.X)
		} else {
			visitChain(callExpr.Fun)
		}
	}
	visitChain(call)
}

func (p *ASTParser) reportConstructorArg(arg ast.Expr, assigned map[string]ast.Expr, model, methodName string, seenCalls map[*ast.CallExpr]bool, isFieldName bool) {
	resolved := p.resolveAssignedExpr(arg, assigned)
	if lit, ok := resolved.(*ast.BasicLit); ok {
		_ = lit
		return
	}
	if ident, ok := resolved.(*ast.Ident); ok && isBuiltinIdent(ident.Name) {
		return
	}
	call, ok := resolved.(*ast.CallExpr)
	if !ok {
		if ident, ok := arg.(*ast.Ident); ok {
			if _, ok := assigned[ident.Name]; ok {
				// Assigned non-call values (e.g. strings) used as option
				// arguments are not evaluated by the extractor.
				p.reportDiagnostic(arg.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", ident.Name))
				return
			}
		}
		// Non-literal field names and option values are dropped silently.
		if !isFieldName {
			if ident, ok := resolved.(*ast.Ident); ok {
				p.reportDiagnostic(arg.Pos(), model, methodName, fmt.Sprintf("value of %s is computed at run time and cannot be evaluated during generation", ident.Name))
			}
		}
		return
	}
	if p.isSupportedOptionCall(methodName, call) {
		// A recognized schema option: recurse into its own arguments so
		// helpers such as schema.Default(loadDefault()) are reported.
		for _, nested := range call.Args {
			p.reportConstructorArg(nested, assigned, model, methodName, seenCalls, false)
		}
		// Chained helpers inside the option call itself (rare) are visited too.
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			p.reportNestedConstructorHelpers(call, assigned, model, methodName, seenCalls)
			_ = sel
		}
		return
	}
	p.reportHelperCallOnce(call, model, methodName, seenCalls)
}

func chainMethod(call *ast.CallExpr) (string, bool) {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "schema" {
			return sel.Sel.Name, true
		}
		return sel.Sel.Name, false
	}
	if ident, ok := call.Fun.(*ast.Ident); ok {
		return ident.Name, false
	}
	return "", false
}

func (p *ASTParser) isSupportedOptionCall(methodName string, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "schema" {
		return false
	}
	name := sel.Sel.Name
	if p.isFieldBuilder(name) {
		return true
	}
	if _, ok := fieldOptionParsers[name]; ok {
		return true
	}
	switch name {
	case "OnDelete", "OnUpdate", "RelatedName", "Through", "ThroughTable", "CascadeOnDelete",
		"ForeignKey", "ForeignKeyField", "OneToOne", "OneToOneField", "OneToMany", "OneToManyField", "ManyToMany", "ManyToManyField",
		"Required", "Optional", "Primary", "Unique", "Default", "Build":
		return true
	}
	return methodName == "Relations"
}

// isSupportedElement mirrors the extractor: it reports whether a resolved
// slice/append element would be consumed during generation.
func (p *ASTParser) isSupportedElement(methodName string, resolved ast.Expr) bool {
	if call, ok := resolved.(*ast.CallExpr); ok {
		return p.isSchemaConstructor(methodName, call)
	}
	if methodName == "Relations" {
		return p.extractRelationFromExpr(resolved) != nil
	}
	return false
}

func (p *ASTParser) reportHelperCallOnce(call *ast.CallExpr, model, methodName string, seen map[*ast.CallExpr]bool) {
	if seen[call] {
		return
	}
	seen[call] = true
	p.reportDiagnostic(call.Pos(), model, methodName, fmt.Sprintf("helper call %s(...) cannot be evaluated during generation; declare fields with direct schema constructor calls", callName(call)))
}

func (p *ASTParser) computedValue(expr ast.Expr, assigned map[string]ast.Expr, methodName string) (string, bool) {
	switch expr := expr.(type) {
	case *ast.Ident:
		value, ok := assigned[expr.Name]
		if ok {
			if call, ok := value.(*ast.CallExpr); ok {
				if p.isSchemaConstructor(methodName, call) {
					return "", false
				}
				return expr.Name, true
			}
		}
	case *ast.SelectorExpr:
		if _, ok := p.computedValue(expr.X, assigned, methodName); ok {
			return expressionName(expr), true
		}
		if call, ok := expr.X.(*ast.CallExpr); ok {
			if p.isSchemaConstructor(methodName, call) {
				return "", false
			}
			return expressionName(expr), true
		}
	}
	return "", false
}

func (p *ASTParser) isComputedSlice(expr ast.Expr, assigned map[string]ast.Expr, methodName string) bool {
	if call, ok := expr.(*ast.CallExpr); ok {
		return !p.isSchemaConstructor(methodName, call)
	}
	if resolved, ok := assigned[identName(expr)]; ok {
		if call, ok := resolved.(*ast.CallExpr); ok {
			return !p.isSchemaConstructor(methodName, call)
		}
	}
	if _, ok := p.computedValue(expr, assigned, methodName); ok {
		return true
	}
	_, ok := expr.(*ast.SliceExpr)
	return ok
}

func identName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func isBuiltinIdent(name string) bool {
	switch name {
	case "true", "false", "nil":
		return true
	}
	return false
}

func isAppend(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "append" && len(call.Args) > 1
}

func (p *ASTParser) isSchemaConstructor(methodName string, call *ast.CallExpr) bool {
	switch methodName {
	case "Fields":
		return p.extractFieldFromCall(call) != nil
	case "Relations":
		return p.extractRelationFromExpr(call) != nil
	}
	return false
}

func callName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	}
	return "expression"
}

func expressionName(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return expressionName(expr.X) + "." + expr.Sel.Name
	case *ast.CallExpr:
		return callName(expr) + "()"
	}
	return "expression"
}
