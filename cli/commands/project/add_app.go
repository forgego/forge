package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/cobra"
)

// AddAppCommand creates the "add app" command
type AddAppCommand struct{}

// NewAddAppCommand creates a new instance of AddAppCommand
func NewAddAppCommand() *AddAppCommand {
	return &AddAppCommand{}
}

// Definition returns the cobra command definition
func (c *AddAppCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app [app-name]",
		Short: "Add a new app to the project",
		Long:  "Create a new app directory with models.go, admin.go, and api.go files",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().Bool("example", false, "Include an example model in models.go")
	return cmd
}

// Execute runs the command logic
func (c *AddAppCommand) Execute(ctx *core.Context, args []string) error {
	appName := args[0]

	// Validate app name
	if err := validateAppName(appName); err != nil {
		return err
	}

	// Detect project root
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	// Check if app already exists
	appPath := filepath.Join(projectRoot, "app", appName)
	if _, err := os.Stat(appPath); err == nil {
		return fmt.Errorf("app %s already exists", appName)
	}

	// Create app directory
	if err := os.MkdirAll(appPath, 0755); err != nil {
		return fmt.Errorf("failed to create app directory: %w", err)
	}

	// Prepare template data
	templateData := templates.TemplateData{
		AppName: appName,
	}

	// Create models.go
	modelsPath := filepath.Join(appPath, "models.go")
	modelsContent := `package ` + appName + `

import (
	"github.com/forgego/forge/schema"
)

// Add your models here
`
	if err := os.WriteFile(modelsPath, []byte(modelsContent), 0644); err != nil {
		return fmt.Errorf("failed to create models.go: %w", err)
	}

	// Create admin.go
	adminPath := filepath.Join(appPath, "admin.go")
	if err := templates.WriteTemplateFile(adminPath, "admin.go.tmpl", templateData, 0644); err != nil {
		return fmt.Errorf("failed to create admin.go: %w", err)
	}

	// Create api.go
	apiPath := filepath.Join(appPath, "api.go")
	apiContent := `package ` + appName + `

import (
	"github.com/forgego/forge/api"
	httplib "github.com/forgego/forge/server"
)

func init() {
	// Auto-register API routes
	// Example:
	// RegisterYourAPI(router)
}

// RegisterYourAPI registers API endpoints
func RegisterYourAPI(router *httplib.Router) {
	// Create viewset and register routes here
	apiRouter := api.NewRouter("/api/v1")
	// apiRouter.Register("resource", viewset)
	apiRouter.RegisterRoutes(router)
}
`
	if err := os.WriteFile(apiPath, []byte(apiContent), 0644); err != nil {
		return fmt.Errorf("failed to create api.go: %w", err)
	}

	// Optionally create example model
	exampleFlag, _ := ctx.Cmd.Flags().GetBool("example")
	if exampleFlag {
		exampleModel := `package ` + appName + `

import (
	"github.com/forgego/forge/schema"
)

// Example represents an example model
type Example struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Example
func (Example) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.Bool("is_active").Default(true).Build(),
	}
}

// Meta returns model metadata
func (Example) Meta() schema.Meta {
	return schema.Meta{
		TableName:        "examples",
		VerboseName:      "Example",
		VerboseNamePlural: "Examples",
	}
}

// Relations returns all relationship definitions
func (Example) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Hooks returns model lifecycle hooks
func (Example) Hooks() *schema.ModelHooks {
	return nil
}
`
		if err := os.WriteFile(modelsPath, []byte(exampleModel), 0644); err != nil {
			return fmt.Errorf("failed to create example model: %w", err)
		}
	}

	fmt.Printf("✓ Created app: %s\n", appName)
	fmt.Printf("  Location: %s\n", appPath)
	fmt.Printf("  Files: models.go, admin.go, api.go\n")

	return nil
}

// validateAppName validates the app name
func validateAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("app name cannot contain spaces")
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("app name cannot contain path separators")
	}
	return nil
}

// detectProjectRoot detects the project root by looking for go.mod
func detectProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found (no go.mod)")
}
