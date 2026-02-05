package generator

import (
	"fmt"
)

// Generator is the main code generator
type Generator struct {
	parser    *ASTParser
	writer    *Writer
	modelsDir string
	outputDir string
}

// GetParser returns the AST parser
func (g *Generator) GetParser() *ASTParser {
	return g.parser
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

	// Generate all models in a single gen.go file
	if err := g.generateCombined(definitions); err != nil {
		return fmt.Errorf("failed to generate combined code: %w", err)
	}

	return nil
}

// generateCombined generates all models in a single gen.go file
func (g *Generator) generateCombined(definitions []*ModelDefinition) error {
	return g.writer.WriteCombined(definitions, g.outputDir)
}

// ModelDefinition represents a parsed model definition
type ModelDefinition struct {
	Hooks     HooksDefinition
	Package   string
	Name      string
	Meta      MetaDefinition
	Fields    []FieldDefinition
	Relations []RelationDefinition
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
