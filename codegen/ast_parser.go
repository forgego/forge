package generator

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// ASTParser parses Go AST to extract schema definitions
type ASTParser struct {
	fset        *token.FileSet
	diagnostics []Diagnostic
}

// NewASTParser creates a new AST parser
func NewASTParser() *ASTParser {
	return &ASTParser{
		fset: token.NewFileSet(),
	}
}

// ParseDirectory parses all Go files in a directory and extracts schema definitions
func (p *ASTParser) ParseDirectory(dir string) ([]*ModelDefinition, error) {
	p.diagnostics = nil
	var definitions []*ModelDefinition
	// declaredIDMethods records, per type, whether each declared GetID/SetID
	// has the signature orm.ModelWithID requires.
	declaredIDMethods := make(map[string]map[string]bool)
	embeddedTypesByModel := make(map[string][]embeddedType)
	buildContext := build.Default
	buildContext.BuildTags = buildTagsFromGOFLAGS(os.Getenv("GOFLAGS"))

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		matched, matchErr := buildContext.MatchFile(filepath.Dir(path), filepath.Base(path))
		if matchErr != nil {
			return fmt.Errorf("match build constraints for %s: %w", path, matchErr)
		}
		if !matched {
			return nil
		}

		// Skip generated files
		if strings.HasSuffix(path, ".gen.go") || strings.HasSuffix(path, "gen.go") {
			return nil
		}
		node, parseErr := parser.ParseFile(p.fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		for _, decl := range node.Decls {
			if types, ok := decl.(*ast.GenDecl); ok && types.Tok == token.TYPE {
				for _, spec := range types.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					key := node.Name.Name + "." + typeSpec.Name.Name
					for _, field := range structType.Fields.List {
						if len(field.Names) != 0 {
							continue
						}
						embedded := field.Type
						pointer := false
						if star, ok := embedded.(*ast.StarExpr); ok {
							embedded = star.X
							pointer = true
						}
						if ident, ok := embedded.(*ast.Ident); ok {
							embeddedTypesByModel[key] = append(embeddedTypesByModel[key], embeddedType{key: node.Name.Name + "." + ident.Name, pointer: pointer})
						}
					}
				}
			}
			method, ok := decl.(*ast.FuncDecl)
			if !ok || method.Recv == nil || len(method.Recv.List) != 1 {
				continue
			}
			receiver := method.Recv.List[0].Type
			if pointer, ok := receiver.(*ast.StarExpr); ok {
				receiver = pointer.X
			}
			receiverName, ok := receiver.(*ast.Ident)
			if !ok {
				continue
			}
			key := node.Name.Name + "." + receiverName.Name
			var valid bool
			switch method.Name.Name {
			case "GetID":
				valid = fieldListHasTypes(method.Type.Params, nil) && fieldListHasTypes(method.Type.Results, []string{"int64"})
			case "SetID":
				valid = fieldListHasTypes(method.Type.Params, []string{"int64"}) && fieldListHasTypes(method.Type.Results, nil)
			default:
				continue
			}
			if declaredIDMethods[key] == nil {
				declaredIDMethods[key] = make(map[string]bool)
			}
			declaredIDMethods[key][method.Name.Name] = valid
		}

		defs, err := p.ParseFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		definitions = append(definitions, defs...)
		return nil
	})

	if err == nil {
		hasIDMethod := func(key, name string) bool {
			_, usable, found := resolveIDMethod(key, name, declaredIDMethods, embeddedTypesByModel, make(map[string]bool))
			return found && usable
		}
		for _, definition := range definitions {
			key := definition.Package + "." + definition.Name
			definition.hasWritableIntegerID = definition.hasWritableIntegerID || (hasIDMethod(key, "GetID") && hasIDMethod(key, "SetID"))
		}
	}
	return definitions, err
}

// embeddedType is an anonymous struct field that can promote methods.
type embeddedType struct {
	key     string
	pointer bool
}

// resolveIDMethod follows Go's method promotion rules for name on the type
// key: a declared method shadows promoted ones, the shallowest promotion
// wins, and two at the same depth are ambiguous. found reports whether the
// type has a method with that name at all; usable reports whether it has the
// required signature, is unambiguous, and is not promoted through an
// embedded pointer, which is nil on a freshly allocated model and would
// panic when Create calls SetID after the insert.
func resolveIDMethod(key, name string, declared map[string]map[string]bool, embedded map[string][]embeddedType, visiting map[string]bool) (depth int, usable, found bool) {
	if valid, ok := declared[key][name]; ok {
		return 0, valid, true
	}
	if visiting[key] {
		return 0, false, false
	}
	visiting[key] = true
	defer delete(visiting, key)

	best, matches := -1, 0
	for _, e := range embedded[key] {
		d, ok, f := resolveIDMethod(e.key, name, declared, embedded, visiting)
		if !f {
			continue
		}
		d++
		ok = ok && !e.pointer
		switch {
		case best == -1 || d < best:
			best, matches, usable = d, 1, ok
		case d == best:
			matches++
		}
	}
	if best == -1 {
		return 0, false, false
	}
	return best, usable && matches == 1, true
}

// buildTagsFromGOFLAGS returns the build tags set with -tags in goflags, so
// model files are selected with the same constraints the go command uses.
func buildTagsFromGOFLAGS(goflags string) []string {
	var tags []string
	fields := strings.Fields(goflags)
	for i := 0; i < len(fields); i++ {
		flag := strings.TrimPrefix(strings.TrimPrefix(fields[i], "-"), "-")
		var value string
		switch {
		case strings.HasPrefix(flag, "tags="):
			value = strings.TrimPrefix(flag, "tags=")
		case flag == "tags" && i+1 < len(fields):
			i++
			value = fields[i]
		default:
			continue
		}
		tags = tags[:0]
		for _, tag := range strings.Split(value, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				tags = append(tags, tag)
			}
		}
	}
	return tags
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
		Package:              packageName,
		Name:                 modelName,
		Fields:               []FieldDefinition{},
		Relations:            []RelationDefinition{},
		Meta:                 MetaDefinition{},
		Hooks:                HooksDefinition{},
		structFieldsKnown:    true,
		hasWritableIntegerID: hasAPICompatibleID(modelName, structType, file),
	}

	// Find Fields() method
	fieldsMethod := p.findMethod(file, modelName, "Fields")
	if fieldsMethod != nil {
		p.reportUnsupportedExpressions(modelName, "Fields", fieldsMethod)
		fields, err := p.extractFields(fieldsMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract fields: %w", err)
		}
		def.Fields = fields
	}

	// Find Relations() method
	relationsMethod := p.findMethod(file, modelName, "Relations")
	if relationsMethod != nil {
		p.reportUnsupportedExpressions(modelName, "Relations", relationsMethod)
		relations, err := p.extractRelations(relationsMethod)
		if err != nil {
			return nil, fmt.Errorf("failed to extract relations: %w", err)
		}
		def.Relations = relations
	}

	// Find Meta() method
	metaMethod := p.findMethod(file, modelName, "Meta")
	if metaMethod != nil {
		p.reportUnsupportedExpressions(modelName, "Meta", metaMethod)
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

	p.reportUndeclaredStructFields(def, typeSpec, structType, fieldsMethod)

	return def, nil
}

func hasAPICompatibleID(modelName string, structType *ast.StructType, file *ast.File) bool {
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == modelName+"Generated" {
				return true
			}
			continue
		}
		ident, ok := field.Type.(*ast.Ident)
		if !ok || ident.Name != "int64" {
			continue
		}
		for _, name := range field.Names {
			if name.Name == "ID" || name.Name == "Id" {
				return true
			}
		}
	}
	return implementsModelWithID(modelName, file)
}

func implementsModelWithID(modelName string, file *ast.File) bool {
	getID := findModelMethod(file, modelName, "GetID")
	setID := findModelMethod(file, modelName, "SetID")
	return getID != nil && setID != nil &&
		fieldListHasTypes(getID.Type.Params, nil) && fieldListHasTypes(getID.Type.Results, []string{"int64"}) &&
		fieldListHasTypes(setID.Type.Params, []string{"int64"}) && fieldListHasTypes(setID.Type.Results, nil)
}

func findModelMethod(file *ast.File, modelName, methodName string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		method, ok := decl.(*ast.FuncDecl)
		if !ok || method.Name.Name != methodName || method.Recv == nil || len(method.Recv.List) != 1 {
			continue
		}
		receiver := method.Recv.List[0].Type
		if pointer, ok := receiver.(*ast.StarExpr); ok {
			receiver = pointer.X
		}
		if ident, ok := receiver.(*ast.Ident); ok && ident.Name == modelName {
			return method
		}
	}
	return nil
}

func fieldListHasTypes(fields *ast.FieldList, types []string) bool {
	if fields == nil {
		return len(types) == 0
	}
	actual := make([]string, 0, fields.NumFields())
	for _, field := range fields.List {
		ident, ok := field.Type.(*ast.Ident)
		if !ok {
			return false
		}
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			actual = append(actual, ident.Name)
		}
	}
	if len(actual) != len(types) {
		return false
	}
	for i := range actual {
		if actual[i] != types[i] {
			return false
		}
	}
	return true
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
