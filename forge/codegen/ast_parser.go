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

// extractFields extracts field definitions from Fields() method
func (p *ASTParser) extractFields(method *ast.FuncDecl) ([]FieldDefinition, error) {
	var fields []FieldDefinition
	processed := make(map[ast.Node]bool) // Track processed nodes

	// Walk the method body to find field definitions
	// We look for return statements with composite literals containing field builder calls
	ast.Inspect(method.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ReturnStmt:
			// Process return statement - it contains the field definitions
			for _, result := range x.Results {
				if compLit, ok := result.(*ast.CompositeLit); ok {
					// This is a slice literal like []schema.Field{...}
					for _, elt := range compLit.Elts {
						if call, ok := elt.(*ast.CallExpr); ok && !processed[call] {
							processed[call] = true
							if field := p.extractFieldFromCall(call); field != nil {
								fields = append(fields, *field)
							}
						}
					}
				}
			}
		}
		return true
	})

	return fields, nil
}

// extractFieldFromCall extracts a field definition from a call expression (which may be a method chain)
func (p *ASTParser) extractFieldFromCall(call *ast.CallExpr) *FieldDefinition {
	// Find the field builder call in the chain (Int64, String, etc.)
	fieldBuilderCall, fieldType, fieldName := p.findFieldBuilderInChain(call)
	if fieldBuilderCall == nil || fieldName == "" {
		return nil
	}

	// Extract all options from the entire method chain
	options := make(map[string]interface{})
	p.extractOptionsFromChain(call, options)

	// Build validation tag
	validationTag := p.buildValidationTag(fieldType, options)

	// Extract field properties from options
	required := false
	if req, ok := options["required"].(bool); ok {
		required = req
	}

	primaryKey := false
	if pk, ok := options["primary"].(bool); ok {
		primaryKey = pk
	} else if pk, ok := options["primary_key"].(bool); ok {
		primaryKey = pk
	}

	autoIncrement := false
	if ai, ok := options["auto_increment"].(bool); ok {
		autoIncrement = ai
	}

	defaultValue := options["default"]

	return &FieldDefinition{
		Name:          fieldName,
		Type:          fieldType,
		GoType:        p.mapFieldTypeToGoType(fieldType),
		ValidationTag: validationTag,
		Required:      required,
		PrimaryKey:    primaryKey,
		AutoIncrement: autoIncrement,
		Default:       defaultValue,
		Options:       options,
	}
}

// findFieldBuilderInChain finds the field builder call in a method chain and returns it along with field type and name
func (p *ASTParser) findFieldBuilderInChain(expr ast.Expr) (*ast.CallExpr, string, string) {
	switch x := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			// Check if this is a field builder call
			fieldType := sel.Sel.Name
			if p.isFieldBuilder(fieldType) {
				fieldName := p.extractStringArg(x)
				return x, fieldType, fieldName
			}

			// Otherwise, continue traversing down the chain
			return p.findFieldBuilderInChain(sel.X)
		}
		// Not a selector, might be the function itself
		return p.findFieldBuilderInChain(x.Fun)
	case *ast.SelectorExpr:
		// Continue traversing
		return p.findFieldBuilderInChain(x.X)
	}
	return nil, "", ""
}

// extractOptionsFromChain extracts all options from a method chain
func (p *ASTParser) extractOptionsFromChain(expr ast.Expr, options map[string]interface{}) {
	switch x := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			methodName := sel.Sel.Name

			// Skip Build() - it's just a finalizer
			// Skip field builder methods (Int64, String, etc.) - we already handled those
			if methodName != "Build" && !p.isFieldBuilder(methodName) {
				// Extract option from this method call
				p.extractOptionFromMethod(methodName, x, options)
			}

			// Continue traversing down the chain
			p.extractOptionsFromChain(sel.X, options)
		} else {
			// Not a selector, continue with the function
			p.extractOptionsFromChain(x.Fun, options)
		}
	case *ast.SelectorExpr:
		// Continue traversing
		p.extractOptionsFromChain(x.X, options)
	}
}

// isFieldBuilder checks if a name is a field builder function
func (p *ASTParser) isFieldBuilder(name string) bool {
	builders := []string{
		"Int64", "Int32", "Int",
		"String", "Text",
		"Bool",
		"Time", "Date", "DateTime",
		"Email", "URL", "UUID",
		"JSON", "Bytes",
		"Float64", "Decimal",
		"ForeignKey", "OneToOne", "OneToMany", "ManyToMany",
		// Functional variants
		"Int64Field", "Int32Field", "IntField",
		"StringField", "TextField",
		"BoolField",
		"TimeField", "DateField", "DateTimeField",
		"EmailField", "URLField", "UUIDField",
		"JSONField", "BytesField",
		"Float64Field", "DecimalField",
		"Float32Field",
		"ForeignKeyField", "OneToOneField", "OneToManyField", "ManyToManyField",
	}
	for _, b := range builders {
		if name == b {
			return true
		}
	}
	return false
}

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

// mapFieldTypeToGoType maps field builder type to Go type
func (p *ASTParser) mapFieldTypeToGoType(fieldType string) string {
	mapping := map[string]string{
		"Int64":           "int64",
		"Int64Field":      "int64",
		"Int32":           "int32",
		"Int32Field":      "int32",
		"Int":             "int64", // Alias for Int64
		"IntField":        "int64",
		"String":          "string",
		"StringField":     "string",
		"Text":            "string",
		"TextField":       "string",
		"Bool":            "bool",
		"BoolField":       "bool",
		"Time":            "time.Time",
		"TimeField":       "time.Time",
		"Date":            "time.Time",
		"DateField":       "time.Time",
		"DateTime":        "time.Time",
		"DateTimeField":   "time.Time",
		"Email":           "string",
		"EmailField":      "string",
		"URL":             "string",
		"URLField":        "string",
		"UUID":            "string",
		"UUIDField":       "string",
		"JSON":            "[]byte",
		"JSONField":       "[]byte",
		"Bytes":           "[]byte",
		"BytesField":      "[]byte",
		"Float64":         "float64",
		"Float64Field":    "float64",
		"Float32":         "float32",
		"Float32Field":    "float32",
		"Decimal":         "float64",
		"DecimalField":    "float64",
		"ForeignKey":      "int64", // Foreign keys are typically int64
		"ForeignKeyField": "int64",
		"OneToOne":        "int64",
		"OneToOneField":   "int64",
		"OneToMany":       "int64",
		"OneToManyField":  "int64",
		"ManyToMany":      "int64",
		"ManyToManyField": "int64",
	}

	if goType, ok := mapping[fieldType]; ok {
		return goType
	}

	return "interface{}"
}

// extractOptionFromMethod extracts option value from a method call
func (p *ASTParser) extractOptionFromMethod(methodName string, call *ast.CallExpr, options map[string]interface{}) {
	switch methodName {
	case "Primary":
		options["primary"] = true
		options["primary_key"] = true
	case "AutoIncrement":
		options["auto_increment"] = true
	case "Required":
		options["required"] = true
	case "Optional":
		options["required"] = false
	case "Unique":
		options["unique"] = true
	case "Blank":
		options["blank"] = true
	case "DBIndex":
		options["db_index"] = true
	case "DBColumn":
		if len(call.Args) > 0 {
			if colName := p.extractStringArg(call); colName != "" {
				options["db_column"] = colName
			}
		}
	case "MaxLength":
		if len(call.Args) > 0 {
			if maxLen := p.extractIntArg(call); maxLen != nil {
				options["max_length"] = *maxLen
			}
		}
	case "MinLength":
		if len(call.Args) > 0 {
			if minLen := p.extractIntArg(call); minLen != nil {
				options["min_length"] = *minLen
			}
		}
	case "MaxValue":
		if len(call.Args) > 0 {
			if maxVal := p.extractFloatArg(call); maxVal != nil {
				options["max_value"] = *maxVal
			}
		}
	case "MinValue":
		if len(call.Args) > 0 {
			if minVal := p.extractFloatArg(call); minVal != nil {
				options["min_value"] = *minVal
			}
		}
	case "Default":
		if len(call.Args) > 0 {
			if defaultValue := p.extractDefaultValue(call); defaultValue != nil {
				options["default"] = defaultValue
			}
		}
	case "HelpText":
		if len(call.Args) > 0 {
			if helpText := p.extractStringArg(call); helpText != "" {
				options["help_text"] = helpText
			}
		}
	case "VerboseName":
		if len(call.Args) > 0 {
			if verboseName := p.extractStringArg(call); verboseName != "" {
				options["verbose_name"] = verboseName
			}
		}
	case "AutoNow":
		options["auto_now"] = true
	case "AutoNowAdd":
		options["auto_now_add"] = true
	case "WriteOnly":
		options["write_only"] = true
	case "Editable":
		if len(call.Args) > 0 {
			if editable := p.extractBoolArg(call); editable != nil {
				options["editable"] = *editable
			}
		}
	case "Choices":
		// Choices is more complex, extract if needed
		if len(call.Args) > 0 {
			options["has_choices"] = true
		}
	case "MaxDigits":
		if len(call.Args) > 0 {
			if maxDigits := p.extractIntArg(call); maxDigits != nil {
				options["max_digits"] = *maxDigits
			}
		}
	case "DecimalPlaces":
		if len(call.Args) > 0 {
			if decimalPlaces := p.extractIntArg(call); decimalPlaces != nil {
				options["decimal_places"] = *decimalPlaces
			}
		}
	}
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
	if len(call.Args) == 0 {
		return nil
	}

	if ident, ok := call.Args[0].(*ast.Ident); ok {
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

// extractDefaultValue extracts default value from a function call
func (p *ASTParser) extractDefaultValue(call *ast.CallExpr) interface{} {
	if len(call.Args) == 0 {
		return nil
	}

	arg := call.Args[0]

	// Try string
	if str := p.extractStringArg(call); str != "" {
		return str
	}

	// Try int
	if intVal := p.extractIntArg(call); intVal != nil {
		return *intVal
	}

	// Try float
	if floatVal := p.extractFloatArg(call); floatVal != nil {
		return *floatVal
	}

	// Try bool
	if boolVal := p.extractBoolArg(call); boolVal != nil {
		return *boolVal
	}

	// Try identifier (like true, false, nil)
	if ident, ok := arg.(*ast.Ident); ok {
		return ident.Name
	}

	return nil
}

// buildValidationTag builds a validation tag from field type and options
func (p *ASTParser) buildValidationTag(fieldType string, options map[string]interface{}) string {
	var tags []string

	// Check if Required() was called
	if required, ok := options["required"].(bool); ok && required {
		tags = append(tags, "required")
	}

	// Type-specific validations
	switch fieldType {
	case "Email":
		tags = append(tags, "email")
	case "URL":
		tags = append(tags, "url")
	case "UUID":
		tags = append(tags, "uuid")
	}

	// MaxLength
	if maxLen, ok := options["max_length"]; ok {
		if maxLenInt, ok := maxLen.(int); ok && maxLenInt > 0 {
			tags = append(tags, fmt.Sprintf("max=%d", maxLenInt))
		}
	}

	// MinLength
	if minLen, ok := options["min_length"]; ok {
		if minLenInt, ok := minLen.(int); ok && minLenInt > 0 {
			tags = append(tags, fmt.Sprintf("min=%d", minLenInt))
		}
	}

	if len(tags) == 0 {
		return ""
	}

	return strings.Join(tags, ",")
}

// extractRelations extracts relation definitions from Relations() method
func (p *ASTParser) extractRelations(method *ast.FuncDecl) ([]RelationDefinition, error) {
	var relations []RelationDefinition

	if method.Body == nil {
		return relations, nil
	}

	// Find return statement
	for _, stmt := range method.Body.List {
		if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
			if len(retStmt.Results) > 0 {
				// Get the slice literal
				if sliceLit, ok := retStmt.Results[0].(*ast.CompositeLit); ok {
					// Iterate through slice elements
					for _, elt := range sliceLit.Elts {
						relation := p.extractRelationFromExpr(elt)
						if relation != nil {
							relations = append(relations, *relation)
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
		// Check if it's a builder call like schema.ForeignKey()
		// Find the builder call in the chain
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
		}
	}

	if relation.Name == "" && relation.To == "" {
		return nil
	}

	return relation
}

// findRelationBuilderInChain finds the relation builder call in a method chain
func (p *ASTParser) findRelationBuilderInChain(expr ast.Expr) (*ast.CallExpr, string) {
	switch x := expr.(type) {
	case *ast.CallExpr:
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "schema" {
				builderType := sel.Sel.Name
				if builderType == "ForeignKey" || builderType == "OneToOne" || builderType == "ManyToMany" || builderType == "OneToMany" {
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
					// Extract the constant name (e.g., "CascadeCASCADE")
					options["on_delete"] = sel.Sel.Name
				}
			} else if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["on_delete"] = val
			}
		}
	case "OnUpdate":
		if len(call.Args) > 0 {
			if sel, ok := call.Args[0].(*ast.SelectorExpr); ok {
				// Handle schema.CascadeCASCADE, etc.
				if pkgIdent, ok := sel.X.(*ast.Ident); ok && pkgIdent.Name == "schema" {
					options["on_update"] = sel.Sel.Name
				}
			} else if val := p.extractStringFromExpr(call.Args[0]); val != "" {
				options["on_update"] = val
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

// extractMeta extracts meta definition from Meta() method
func (p *ASTParser) extractMeta(method *ast.FuncDecl) (MetaDefinition, error) {
	meta := MetaDefinition{}

	if method.Body == nil {
		return meta, nil
	}

	// Find return statement
	for _, stmt := range method.Body.List {
		if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
			if len(retStmt.Results) > 0 {
				// Get the composite literal (struct literal)
				if compLit, ok := retStmt.Results[0].(*ast.CompositeLit); ok {
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

// extractHooks extracts hooks definition from Hooks() method
func (p *ASTParser) extractHooks(method *ast.FuncDecl) (HooksDefinition, error) {
	hooks := HooksDefinition{}

	// TODO: Implement hooks extraction
	// Extract hook function names or references

	return hooks, nil
}
