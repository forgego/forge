package generator

import (
	"fmt"
)

// Generator is the main code generator
type Generator struct {
	parser      *ASTParser
	writer      *Writer
	modelsDir   string
	outputDir   string
	generateAPI bool
}

// GetParser returns the AST parser
func (g *Generator) GetParser() *ASTParser {
	return g.parser
}

// Diagnostics returns expressions the generator could not evaluate.
func (g *Generator) Diagnostics() []Diagnostic {
	return g.parser.Diagnostics()
}

// SetGenerateAPI enables or disables REST API generation
func (g *Generator) SetGenerateAPI(enabled bool) *Generator {
	g.generateAPI = enabled
	return g
}

// NewGenerator creates a new generator
func NewGenerator(modelsDir, outputDir string) *Generator {
	return &Generator{
		modelsDir: modelsDir,
		outputDir: outputDir,
		parser:    NewASTParser(),
		writer:    NewWriter(),
	}
}

// Generate generates code from schema definitions
func (g *Generator) Generate() error {
	// Parse all schema files
	definitions, err := g.parser.ParseDirectory(g.modelsDir)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}

	if len(definitions) == 0 {
		fmt.Println("No schema definitions found")
		return nil
	}

	// Validate all API constraints before writing either generated file so a
	// failed API generation cannot leave models and API output out of sync.
	if g.generateAPI {
		if err := ValidateAPIModels(definitions); err != nil {
			return fmt.Errorf("failed to generate API code: %w", err)
		}
	}

	if g.generateAPI {
		if err := g.writer.WriteCombinedAndAPI(definitions, g.outputDir); err != nil {
			return fmt.Errorf("failed to generate API code: %w", err)
		}
		return nil
	}

	// Generate all models in a single gen.go file.
	if err := g.generateCombined(definitions); err != nil {
		return fmt.Errorf("failed to generate combined code: %w", err)
	}

	return nil
}

// GenerateAPI parses schemas and generates REST API code (api_gen.go)
func (g *Generator) GenerateAPI() error {
	definitions, err := g.parser.ParseDirectory(g.modelsDir)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}

	if len(definitions) == 0 {
		return nil
	}
	if err := ValidateAPIModels(definitions); err != nil {
		return fmt.Errorf("failed to generate API code: %w", err)
	}

	return g.writer.WriteAPI(definitions, g.outputDir)
}

// generateCombined generates all models in a single gen.go file
func (g *Generator) generateCombined(definitions []*ModelDefinition) error {
	return g.writer.WriteCombined(definitions, g.outputDir)
}

// int64APIKeyFieldTypes are the schema field builder names codegen maps to
// int64, used as a fallback when GoType is empty.
var int64APIKeyFieldTypes = map[string]bool{
	"BigInt":          true,
	"ForeignKey":      true,
	"ForeignKeyField": true,
	"Int":             true,
	"IntField":        true,
	"Int64":           true,
	"Int64Field":      true,
	"ManyToMany":      true,
	"ManyToManyField": true,
	"OneToMany":       true,
	"OneToManyField":  true,
	"OneToOne":        true,
	"OneToOneField":   true,
}

// ValidateAPIModels rejects models whose primary key cannot be used by the
// generated REST API manager, which requires an int64 field named id.
func ValidateAPIModels(definitions []*ModelDefinition) error {
	for _, def := range definitions {
		if def == nil {
			continue
		}
		var primaryKeys []FieldDefinition
		for _, f := range def.Fields {
			if f.PrimaryKey {
				primaryKeys = append(primaryKeys, f)
			}
		}
		if len(primaryKeys) == 0 {
			return fmt.Errorf("model %s has no primary key; generated REST APIs require an int64 primary key named \"id\"", def.Name)
		}
		if len(primaryKeys) > 1 {
			return fmt.Errorf("model %s has more than one primary key field (%q and %q); generated REST APIs require exactly one int64 primary key named \"id\"", def.Name, primaryKeys[0].Name, primaryKeys[1].Name)
		}

		primaryKey := primaryKeys[0]
		label := primaryKey.Type
		if label == "" {
			label = primaryKey.GoType
		}
		if label == "" {
			label = "unknown"
		}
		isInt64 := primaryKey.GoType == "int64" || (primaryKey.GoType == "" && int64APIKeyFieldTypes[primaryKey.Type])
		if !isInt64 {
			if (primaryKey.GoType == "" && !isKnownIntegerAPIKeyType(primaryKey.Type)) ||
				(primaryKey.GoType != "" && !isIntegerGoType(primaryKey.GoType)) {
				return fmt.Errorf("model %s has non-integer primary key field %q (type %s): generated REST APIs require an int64 primary key named \"id\"", def.Name, primaryKey.Name, label)
			}
			return fmt.Errorf("model %s has primary key field %q with type %s; generated REST APIs require an int64 primary key named \"id\"", def.Name, primaryKey.Name, label)
		}
		if primaryKey.Name != "id" {
			return fmt.Errorf("model %s has primary key field %q; generated REST APIs require an int64 primary key named \"id\"", def.Name, primaryKey.Name)
		}
		if !primaryKey.AutoIncrement {
			return fmt.Errorf("model %s has a non-auto-increment primary key field \"id\"; the generated API requires an auto-increment id", def.Name)
		}
		if def.structFieldsKnown && !def.hasWritableIntegerID {
			return fmt.Errorf("model %s declares an auto-increment id in Fields() but has no concrete int64 ID or Id field and does not implement orm.ModelWithID", def.Name)
		}
	}
	return nil
}

func isKnownIntegerAPIKeyType(fieldType string) bool {
	return fieldType == "Int32" || fieldType == "Int32Field" || int64APIKeyFieldTypes[fieldType]
}

func isIntegerGoType(goType string) bool {
	switch goType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}

// ModelDefinition represents a parsed model definition
type ModelDefinition struct {
	Hooks                HooksDefinition
	Package              string
	Name                 string
	Meta                 MetaDefinition
	Fields               []FieldDefinition
	Relations            []RelationDefinition
	structFieldsKnown    bool
	hasWritableIntegerID bool
}

// FieldDefinition represents a field definition
type FieldDefinition struct {
	Default       interface{}
	Options       map[string]interface{}
	Name          string
	Type          string
	GoType        string
	ValidationTag string
	Required      bool
	PrimaryKey    bool
	AutoIncrement bool
}

// RelationDefinition represents a relation definition
type RelationDefinition struct {
	Options map[string]interface{}
	Name    string
	Type    string
	To      string
}

// MetaDefinition represents model metadata
type MetaDefinition struct {
	TableName         string
	DBColumn          string
	OrderBy           []string
	VerboseName       string
	VerboseNamePlural string
	Indexes           []IndexDefinition
	Constraints       []ConstraintDefinition
	UniqueTogether    [][]string
	AppLabel          string
	Proxy             bool
	Abstract          bool
	Managed           bool
	// ... other Meta options
}

// IndexDefinition represents an index
type IndexDefinition struct {
	Name   string
	Fields []string
	Unique bool
}

// ConstraintDefinition represents a database constraint
type ConstraintDefinition struct {
	Name      string
	Type      string
	Condition string
	Fields    []string
}

// HooksDefinition represents model hooks
type HooksDefinition struct {
	BeforeCreate string
	AfterCreate  string
	BeforeUpdate string
	AfterUpdate  string
	BeforeSave   string
	AfterSave    string
	BeforeDelete string
	AfterDelete  string
	Clean        string
}
