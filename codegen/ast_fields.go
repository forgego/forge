package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

// extractFields extracts field definitions from Fields() method
func (p *ASTParser) extractFields(method *ast.FuncDecl) ([]FieldDefinition, error) {
	var fields []FieldDefinition
	processed := make(map[ast.Node]bool) // Track processed nodes

	if method == nil || method.Body == nil {
		return fields, nil
	}

	assignedExprs := collectAssignedExprs(method.Body)

	// Helper to extract fields from a slice composite literal
	extractFromSliceLit := func(compLit *ast.CompositeLit) {
		for _, elt := range compLit.Elts {
			elt = p.resolveAssignedExpr(elt, assignedExprs)
			if call, ok := elt.(*ast.CallExpr); ok && !processed[call] {
				processed[call] = true
				if field := p.extractFieldFromCall(call); field != nil {
					fields = append(fields, *field)
				}
			}
		}
	}

	// Walk the method body to find field definitions
	for _, stmt := range method.Body.List {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			for _, result := range s.Results {
				resExpr := p.resolveAssignedExpr(result, assignedExprs)
				if compLit, ok := resExpr.(*ast.CompositeLit); ok {
					extractFromSliceLit(compLit)
				}
			}
		case *ast.DeclStmt:
			if gen, ok := s.Decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					if valSpec, ok := spec.(*ast.ValueSpec); ok {
						for _, val := range valSpec.Values {
							if compLit, ok := val.(*ast.CompositeLit); ok {
								extractFromSliceLit(compLit)
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
						if c, ok := arg.(*ast.CallExpr); ok && !processed[c] {
							processed[c] = true
							if field := p.extractFieldFromCall(c); field != nil {
								fields = append(fields, *field)
							}
						}
					}
				}
			}
		case *ast.AssignStmt:
			for _, rhs := range s.Rhs {
				if compLit, ok := rhs.(*ast.CompositeLit); ok {
					extractFromSliceLit(compLit)
				}
				if call, ok := rhs.(*ast.CallExpr); ok {
					if fnIdent, ok := call.Fun.(*ast.Ident); ok && fnIdent.Name == "append" && len(call.Args) > 1 {
						for _, arg := range call.Args[1:] {
							arg = p.resolveAssignedExpr(arg, assignedExprs)
							if c, ok := arg.(*ast.CallExpr); ok && !processed[c] {
								processed[c] = true
								if field := p.extractFieldFromCall(c); field != nil {
									fields = append(fields, *field)
								}
							}
						}
					}
				}
			}
		}
	}

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

	// Also extract options from variadic arguments of the field builder call.
	// The functional API passes options as variadic args:
	//   schema.StringField("name", schema.Required(), schema.MaxLength(200))
	// These are in fieldBuilderCall.Args[1:] (index 0 is the field name).
	p.extractOptionsFromVariadicArgs(fieldBuilderCall, options)

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
	if primaryKey {
		required = true
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

// extractOptionsFromVariadicArgs extracts options from variadic arguments of
// functional-style field builder calls like:
//
//	schema.StringField("name", schema.Required(), schema.MaxLength(200))
//
// The variadic args are call expressions passed as arguments at index 1+.
func (p *ASTParser) extractOptionsFromVariadicArgs(fieldBuilderCall *ast.CallExpr, options map[string]interface{}) {
	if fieldBuilderCall == nil {
		return
	}
	// Skip argument 0 (the field name string); process remaining variadic option args
	for i := 1; i < len(fieldBuilderCall.Args); i++ {
		arg := fieldBuilderCall.Args[i]
		optionCall, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		// The option call is something like schema.Required() or schema.MaxLength(200)
		var methodName string
		switch fn := optionCall.Fun.(type) {
		case *ast.SelectorExpr:
			methodName = fn.Sel.Name
		case *ast.Ident:
			methodName = fn.Name
		default:
			continue
		}
		p.extractOptionFromMethod(methodName, optionCall, options)
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
		"Float32Field", "FloatField",
		"ForeignKeyField", "OneToOneField", "OneToManyField", "ManyToManyField",
	}
	for _, b := range builders {
		if name == b {
			return true
		}
	}
	return false
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
		"Float":           "float64", // Alias for Float64
		"FloatField":      "float64",
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

var fieldOptionParsers = map[string]func(p *ASTParser, call *ast.CallExpr, options map[string]interface{}){
	"Primary":         parsePrimaryOption,
	"AutoIncrement":   trueOption("auto_increment"),
	"Required":        trueOption("required"),
	"Optional":        parseOptionalOption,
	"Unique":          trueOption("unique"),
	"Blank":           trueOption("blank"),
	"DBIndex":         trueOption("db_index"),
	"DBColumn":        stringOption("db_column"),
	"MaxLength":       intOption("max_length"),
	"MinLength":       intOption("min_length"),
	"MaxValue":        floatOption("max_value"),
	"MinValue":        floatOption("min_value"),
	"Default":         parseDefaultOption,
	"HelpText":        stringOption("help_text"),
	"VerboseName":     stringOption("verbose_name"),
	"AutoNow":         trueOption("auto_now"),
	"AutoNowAdd":      trueOption("auto_now_add"),
	"WriteOnly":       trueOption("write_only"),
	"Editable":        boolOption("editable"),
	"Choices":         parseChoicesOption,
	"MaxDigits":       intOption("max_digits"),
	"DecimalPlaces":   intOption("decimal_places"),
	"DBDefault":       stringOption("db_default"),
	"GeneratedColumn": parseGeneratedColumnOption,
	"DBComment":       stringOption("db_comment"),
	"DBCollation":     stringOption("db_collation"),
	"DBTablespace":    stringOption("db_tablespace"),
	"DBType":          stringOption("db_type"),
}

func trueOption(key string) func(*ASTParser, *ast.CallExpr, map[string]interface{}) {
	return func(_ *ASTParser, _ *ast.CallExpr, options map[string]interface{}) {
		options[key] = true
	}
}

func stringOption(key string) func(*ASTParser, *ast.CallExpr, map[string]interface{}) {
	return func(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
		if len(call.Args) > 0 {
			if val := p.extractStringArg(call); val != "" {
				options[key] = val
			}
		}
	}
}

func intOption(key string) func(*ASTParser, *ast.CallExpr, map[string]interface{}) {
	return func(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
		if len(call.Args) > 0 {
			if val := p.extractIntArg(call); val != nil {
				options[key] = *val
			}
		}
	}
}

func floatOption(key string) func(*ASTParser, *ast.CallExpr, map[string]interface{}) {
	return func(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
		if len(call.Args) > 0 {
			if val := p.extractFloatArg(call); val != nil {
				options[key] = *val
			}
		}
	}
}

func boolOption(key string) func(*ASTParser, *ast.CallExpr, map[string]interface{}) {
	return func(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
		if len(call.Args) > 0 {
			if val := p.extractBoolArg(call); val != nil {
				options[key] = *val
			}
		}
	}
}

func parsePrimaryOption(_ *ASTParser, _ *ast.CallExpr, options map[string]interface{}) {
	options["primary"] = true
	options["primary_key"] = true
}

func parseOptionalOption(_ *ASTParser, _ *ast.CallExpr, options map[string]interface{}) {
	options["required"] = false
}

func parseDefaultOption(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
	if len(call.Args) > 0 {
		if defaultValue := p.extractDefaultValue(call); defaultValue != nil {
			options["default"] = defaultValue
		}
	}
}

func parseChoicesOption(_ *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
	if len(call.Args) > 0 {
		options["has_choices"] = true
		var choices []string
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				choices = append(choices, strings.Trim(lit.Value, `"`))
			} else if comp, ok := arg.(*ast.CompositeLit); ok {
				for _, elt := range comp.Elts {
					if slit, ok := elt.(*ast.BasicLit); ok && slit.Kind == token.STRING {
						choices = append(choices, strings.Trim(slit.Value, `"`))
					}
				}
			}
		}
		if len(choices) > 0 {
			options["choices"] = choices
		}
	}
}

func parseGeneratedColumnOption(p *ASTParser, call *ast.CallExpr, options map[string]interface{}) {
	options["generated"] = true
	if len(call.Args) > 0 {
		if expr := p.extractStringArg(call); expr != "" {
			options["generated_expr"] = expr
		}
	}
	if len(call.Args) > 1 {
		if stored := p.extractBoolArgAt(call, 1); stored != nil {
			options["generated_stored"] = *stored
		}
	}
}

// extractOptionFromMethod extracts option value from a method call
func (p *ASTParser) extractOptionFromMethod(methodName string, call *ast.CallExpr, options map[string]interface{}) {
	if parser, ok := fieldOptionParsers[methodName]; ok {
		parser(p, call, options)
	}
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

	// Unique
	if unique, ok := options["unique"].(bool); ok && unique {
		tags = append(tags, "unique")
	}

	// Choices
	if choices, ok := options["choices"].([]string); ok && len(choices) > 0 {
		tags = append(tags, "oneof="+strings.Join(choices, " "))
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

	// MaxValue (numeric)
	if maxVal, ok := options["max_value"]; ok {
		switch v := maxVal.(type) {
		case float64:
			tags = append(tags, "lte="+strconv.FormatFloat(v, 'f', -1, 64))
		case int:
			tags = append(tags, fmt.Sprintf("lte=%d", v))
		}
	}

	// MinValue (numeric)
	if minVal, ok := options["min_value"]; ok {
		switch v := minVal.(type) {
		case float64:
			tags = append(tags, "gte="+strconv.FormatFloat(v, 'f', -1, 64))
		case int:
			tags = append(tags, fmt.Sprintf("gte=%d", v))
		}
	}

	// MaxDigits (decimal)
	if maxDigits, ok := options["max_digits"]; ok {
		if maxDigitsInt, ok := maxDigits.(int); ok && maxDigitsInt > 0 {
			tags = append(tags, fmt.Sprintf("decimal_max_digits=%d", maxDigitsInt))
		}
	}

	// DecimalPlaces
	if decPlaces, ok := options["decimal_places"]; ok {
		if decPlacesInt, ok := decPlaces.(int); ok && decPlacesInt >= 0 {
			tags = append(tags, fmt.Sprintf("decimal_places=%d", decPlacesInt))
		}
	}

	if len(tags) == 0 {
		return ""
	}

	return strings.Join(tags, ",")
}
