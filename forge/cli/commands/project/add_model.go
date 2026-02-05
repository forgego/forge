package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// AddModelCommand creates the "add model" command
type AddModelCommand struct{}

// NewAddModelCommand creates a new instance of AddModelCommand
func NewAddModelCommand() *AddModelCommand {
	return &AddModelCommand{}
}

// Definition returns the cobra command definition
func (c *AddModelCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model [model-name]",
		Short: "Add a new model to an app",
		Long:  "Add a new model definition to an app's models.go file",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("app", "", "App name (auto-detected if in app directory)")
	cmd.Flags().String("table", "", "Table name (default: pluralized model name)")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	return cmd
}

// Execute runs the command logic
func (c *AddModelCommand) Execute(ctx *core.Context, args []string) error {
	modelName := args[0]

	// Validate model name
	if err := validateModelName(modelName); err != nil {
		return err
	}

	// Detect project root and app
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	appName, err := detectAppName(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to detect app name: %w", err)
	}

	// Check if app exists
	appPath := filepath.Join(projectRoot, "app", appName)
	if _, err := os.Stat(appPath); err != nil {
		return fmt.Errorf("app %s does not exist", appName)
	}

	// Get table name
	tableName, _ := ctx.Cmd.Flags().GetString("table")
	if tableName == "" {
		tableName = pluralizeModelName(modelName)
	}

	// Interactive field generation
	fields, err := promptForFields()
	if err != nil {
		return fmt.Errorf("failed to get fields: %w", err)
	}

	// Check dry-run
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("Dry-run mode - would create model %s in app %s\n", modelName, appName)
		fmt.Printf("Table name: %s\n", tableName)
		fmt.Printf("Fields: %d\n", len(fields))
		return nil
	}

	// Generate model code
	modelCode := generateModelCode(appName, modelName, tableName, fields)

	// Append to models.go
	modelsPath := filepath.Join(appPath, "models.go")
	if err := appendToModelsFile(modelsPath, modelCode); err != nil {
		return fmt.Errorf("failed to add model to models.go: %w", err)
	}

	// Update admin.go to register the model
	adminPath := filepath.Join(appPath, "admin.go")
	if err := registerModelInAdmin(adminPath, appName, modelName); err != nil {
		return fmt.Errorf("failed to register model in admin: %w", err)
	}

	fmt.Printf("✓ Added model %s to app %s\n", modelName, appName)
	fmt.Printf("  Table: %s\n", tableName)
	fmt.Printf("  Fields: %d\n", len(fields))

	return nil
}

// validateModelName validates the model name
func validateModelName(name string) error {
	if name == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if !strings.HasPrefix(name, strings.ToUpper(name[:1])) {
		return fmt.Errorf("model name must start with uppercase letter")
	}
	return nil
}

// detectAppName detects the app name from current directory or flag
func detectAppName(ctx *core.Context, projectRoot string) (string, error) {
	// Check flag first
	appFlag, _ := ctx.Cmd.Flags().GetString("app")
	if appFlag != "" {
		return appFlag, nil
	}

	// Try to detect from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Check if we're in an app directory
	appDir := filepath.Join(projectRoot, "app")
	relPath, err := filepath.Rel(appDir, cwd)
	if err == nil && !strings.HasPrefix(relPath, "..") {
		parts := strings.Split(relPath, string(filepath.Separator))
		if len(parts) > 0 && parts[0] != "." {
			return parts[0], nil
		}
	}

	// Prompt for app name
	var appName string
	prompt := &survey.Input{
		Message: "Enter app name:",
	}
	if err := survey.AskOne(prompt, &appName); err != nil {
		return "", fmt.Errorf("failed to get app name: %w", err)
	}

	return appName, nil
}

// FieldDefinition represents a field in a model
type FieldDefinition struct {
	Name      string
	Type      string
	Required  bool
	Unique    bool
	Default   string
	MaxLength int
}

// promptForFields prompts the user for fields interactively
func promptForFields() ([]FieldDefinition, error) {
	var fields []FieldDefinition
	addMore := true

	for addMore {
		var field FieldDefinition

		// Field name
		if err := survey.AskOne(&survey.Input{
			Message: "Field name:",
		}, &field.Name); err != nil {
			return nil, err
		}

		// Field type
		if err := survey.AskOne(&survey.Select{
			Message: "Field type:",
			Options: []string{"String", "Int64", "Bool", "Time", "Float64"},
			Default: "String",
		}, &field.Type); err != nil {
			return nil, err
		}

		// Required
		if err := survey.AskOne(&survey.Confirm{
			Message: "Required?",
			Default: false,
		}, &field.Required); err != nil {
			return nil, err
		}

		// Unique
		if err := survey.AskOne(&survey.Confirm{
			Message: "Unique?",
			Default: false,
		}, &field.Unique); err != nil {
			return nil, err
		}

		// MaxLength for String fields
		if field.Type == "String" {
			if err := survey.AskOne(&survey.Input{
				Message: "Max length (0 for unlimited):",
				Default: "255",
			}, &field.MaxLength); err != nil {
				return nil, err
			}
		}

		fields = append(fields, field)

		// Add more fields?
		if err := survey.AskOne(&survey.Confirm{
			Message: "Add another field?",
			Default: true,
		}, &addMore); err != nil {
			return nil, err
		}
	}

	return fields, nil
}

// generateModelCode generates the model code
func generateModelCode(appName, modelName, tableName string, fields []FieldDefinition) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n// %s represents a %s model\n", modelName, modelName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))
	sb.WriteString("\tschema.BaseSchema\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Fields returns all field definitions for %s\n", modelName))
	sb.WriteString(fmt.Sprintf("func (%s) Fields() []schema.Field {\n", modelName))
	sb.WriteString("\treturn []schema.Field{\n")
	sb.WriteString("\t\tschema.Int64(\"id\").Primary().AutoIncrement().Build(),\n")

	for _, field := range fields {
		fieldCode := generateFieldCode(field)
		sb.WriteString(fmt.Sprintf("\t\t%s,\n", fieldCode))
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Meta returns model metadata\n"))
	sb.WriteString(fmt.Sprintf("func (%s) Meta() schema.Meta {\n", modelName))
	sb.WriteString("\treturn schema.Meta{\n")
	sb.WriteString(fmt.Sprintf("\t\tTableName:        \"%s\",\n", tableName))
	sb.WriteString(fmt.Sprintf("\t\tVerboseName:      \"%s\",\n", modelName))
	sb.WriteString(fmt.Sprintf("\t\tVerboseNamePlural: \"%s\",\n", pluralizeModelName(modelName)))
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Relations returns all relationship definitions\n"))
	sb.WriteString(fmt.Sprintf("func (%s) Relations() []schema.Relation {\n", modelName))
	sb.WriteString("\treturn []schema.Relation{}\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Hooks returns model lifecycle hooks\n"))
	sb.WriteString(fmt.Sprintf("func (%s) Hooks() *schema.ModelHooks {\n", modelName))
	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	return sb.String()
}

// generateFieldCode generates code for a single field
func generateFieldCode(field FieldDefinition) string {
	var parts []string

	switch field.Type {
	case "String":
		parts = append(parts, fmt.Sprintf("schema.String(\"%s\")", field.Name))
		if field.MaxLength > 0 {
			parts = append(parts, fmt.Sprintf("MaxLength(%d)", field.MaxLength))
		}
	case "Int64":
		parts = append(parts, fmt.Sprintf("schema.Int64(\"%s\")", field.Name))
	case "Bool":
		parts = append(parts, fmt.Sprintf("schema.Bool(\"%s\")", field.Name))
	case "Time":
		parts = append(parts, fmt.Sprintf("schema.Time(\"%s\")", field.Name))
	case "Float64":
		parts = append(parts, fmt.Sprintf("schema.Float64(\"%s\")", field.Name))
	}

	if field.Required {
		parts = append(parts, "Required()")
	}
	if field.Unique {
		parts = append(parts, "Unique()")
	}
	if field.Default != "" {
		parts = append(parts, fmt.Sprintf("Default(%s)", field.Default))
	}

	parts = append(parts, "Build()")

	return strings.Join(parts, ".")
}

// appendToModelsFile appends model code to models.go
func appendToModelsFile(filePath string, modelCode string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(modelCode)
	return err
}

// registerModelInAdmin registers the model in admin.go
func registerModelInAdmin(adminPath, appName, modelName string) error {
	// Read existing admin.go
	content, err := os.ReadFile(adminPath)
	if err != nil {
		return err
	}

		// Check if model is already registered
	registrationComment := fmt.Sprintf("// Register %s model for admin\n\t// schemaInstance := &%sSchema{}\n\t// manager := orm.NewManager[%s](db)\n\t// config := &admin.Config[%s]{...}\n\t// admin.Register(schemaInstance, manager, config)", modelName, modelName, modelName, modelName)
	if strings.Contains(string(content), modelName+"Schema") || strings.Contains(string(content), "Register") {
		return nil // Already registered or has registration pattern
	}

	// Find the init function and add registration
	contentStr := string(content)
	if !strings.Contains(contentStr, registrationComment) {
		// Add import if needed
		if !strings.Contains(contentStr, "github.com/forgego/forge/admin") {
			// This is a simplified version - in production, use AST parsing
		}

		// Add registration in init function
		// Format: Comment with example registration code
		newContent := strings.Replace(contentStr,
			"func init() {",
			fmt.Sprintf("func init() {\n\t%s", registrationComment),
			1)

		return os.WriteFile(adminPath, []byte(newContent), 0644)
	}

	return nil
}

// pluralizeModelName creates a simple plural form
func pluralizeModelName(name string) string {
	if strings.HasSuffix(name, "y") {
		return strings.TrimSuffix(name, "y") + "ies"
	}
	if strings.HasSuffix(name, "s") || strings.HasSuffix(name, "x") || strings.HasSuffix(name, "z") {
		return name + "es"
	}
	return name + "s"
}
