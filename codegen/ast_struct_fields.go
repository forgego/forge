package generator

import (
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/forgego/forge/utils"
)

// reportUndeclaredStructFields reports exported struct fields that have no
// matching Fields() or Relations() entry. Generation and migrations only see
// schema entries, so such fields are otherwise dropped without notice.
func (p *ASTParser) reportUndeclaredStructFields(def *ModelDefinition, typeSpec *ast.TypeSpec, structType *ast.StructType, fieldsMethod *ast.FuncDecl) {
	if fieldsMethod == nil {
		if !embedsBaseSchema(structType) {
			return
		}
		p.reportDiagnostic(typeSpec.Pos(), def.Name, "Fields", "model embeds schema.BaseSchema but has no Fields() method; no columns will be generated")
		return
	}
	if hasLiteralFieldEntries(fieldsMethod) {
		// schema.Field{...} literals are not parsed, so the declared set is incomplete.
		return
	}

	declared := make(map[string]bool, len(def.Fields)+len(def.Relations))
	for _, field := range def.Fields {
		declared[field.Name] = true
		if column, ok := field.Options["db_column"].(string); ok {
			declared[column] = true
		}
	}
	for _, relation := range def.Relations {
		// Relation names follow either the Go field name or the column name.
		name := strings.ToLower(relation.Name)
		declared[name] = true
		declared[utils.ToSnake(relation.Name)] = true
		declared[name+"_id"] = true
	}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		column := structTagColumn(field)
		if column == "-" {
			continue
		}
		for _, name := range field.Names {
			if !name.IsExported() || structFieldDeclared(declared, name.Name, column) {
				continue
			}
			want := column
			if want == "" {
				want = utils.ToSnake(name.Name)
			}
			p.reportDiagnostic(name.Pos(), def.Name, "Fields", fmt.Sprintf(
				"struct field %s has no Fields() or Relations() entry for %q; it will not be generated or migrated. Declare it in Fields() or tag it db:\"-\"",
				name.Name, want))
		}
	}
}

func structTagColumn(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	raw, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}
	column, _, _ := strings.Cut(reflect.StructTag(raw).Get("db"), ",")
	return column
}

func structFieldDeclared(declared map[string]bool, goName, column string) bool {
	if column != "" && declared[column] {
		return true
	}
	return declared[utils.ToSnake(goName)] || declared[strings.ToLower(goName)]
}

func embedsBaseSchema(structType *ast.StructType) bool {
	for _, field := range structType.Fields.List {
		if len(field.Names) != 0 {
			continue
		}
		if sel, ok := field.Type.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "schema" && (sel.Sel.Name == "Schema" || sel.Sel.Name == "BaseSchema") {
				return true
			}
		}
	}
	return false
}

func hasLiteralFieldEntries(method *ast.FuncDecl) bool {
	found := false
	ast.Inspect(method.Body, func(n ast.Node) bool {
		slice, ok := n.(*ast.CompositeLit)
		if !ok || found {
			return !found
		}
		for _, elt := range slice.Elts {
			if _, ok := elt.(*ast.CompositeLit); ok {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
