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

// AddAPICommand creates the "add api" command
type AddAPICommand struct{}

// NewAddAPICommand creates a new instance of AddAPICommand
func NewAddAPICommand() *AddAPICommand {
	return &AddAPICommand{}
}

// Definition returns the cobra command definition
func (c *AddAPICommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api [resource-name]",
		Short: "Add a new API endpoint to an app",
		Long:  "Add a new REST API endpoint (viewset + serializer) to an app's api.go file",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("app", "", "App name (auto-detected if in app directory)")
	cmd.Flags().String("model", "", "Model name for the API")
	cmd.Flags().String("resource", "", "URL resource path (default: resource-name)")
	cmd.Flags().Bool("graphql", false, "Include GraphQL support")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	return cmd
}

// Execute runs the command logic
func (c *AddAPICommand) Execute(ctx *core.Context, args []string) error {
	resourceName := args[0]

	// Detect project root and app
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	appName, err := detectAppName(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to detect app name: %w", err)
	}

	// Get model name
	modelName, _ := ctx.Cmd.Flags().GetString("model")
	if modelName == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "Model name:",
			Default: strings.Title(resourceName),
		}, &modelName); err != nil {
			return err
		}
	}

	// Get resource path
	resourcePath, _ := ctx.Cmd.Flags().GetString("resource")
	if resourcePath == "" {
		resourcePath = resourceName
	}

	graphql, _ := ctx.Cmd.Flags().GetBool("graphql")

	// Check dry-run
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("Dry-run mode - would create API %s in app %s\n", resourceName, appName)
		fmt.Printf("Model: %s, Resource: %s\n", modelName, resourcePath)
		if graphql {
			fmt.Printf("GraphQL: enabled\n")
		}
		return nil
	}

	// Note: templateData not used in this implementation, using direct code generation

	// Read existing api.go
	appPath := filepath.Join(projectRoot, "app", appName)
	apiPath := filepath.Join(appPath, "api.go")

	// Generate API code
	apiCode := generateAPICode(appName, modelName, resourceName, resourcePath, graphql)

	// Append to api.go
	file, err := os.OpenFile(apiPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open api.go: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("\n" + apiCode); err != nil {
		return fmt.Errorf("failed to write API code: %w", err)
	}

	fmt.Printf("✓ Added API %s to app %s\n", resourceName, appName)
	fmt.Printf("  Model: %s, Resource: /api/v1/%s\n", modelName, resourcePath)
	if graphql {
		fmt.Printf("  GraphQL: enabled\n")
	}

	return nil
}

// generateAPICode generates the API code
func generateAPICode(appName, modelName, resourceName, resourcePath string, graphql bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\n// Register%sAPI registers the %s API endpoints\n", strings.Title(resourceName), resourceName))
	sb.WriteString(fmt.Sprintf("func Register%sAPI(router *httplib.Router) {\n", strings.Title(resourceName)))
	sb.WriteString("\t// Create viewset\n")
	sb.WriteString("\tviewset := api.NewBaseViewSet(\n")
	sb.WriteString("\t\tfunc() api.Serializer {\n")
	sb.WriteString(fmt.Sprintf("\t\t\treturn New%sSerializer()\n", modelName))
	sb.WriteString("\t\t},\n")
	sb.WriteString(fmt.Sprintf("\t\t%s.Objects.Filter(),\n", modelName))
	sb.WriteString(fmt.Sprintf("\t\t&%s{},\n", modelName))
	sb.WriteString("\t)\n\n")
	sb.WriteString("\t// Register routes\n")
	sb.WriteString("\tapiRouter := api.NewRouter(\"/api/v1\")\n")
	sb.WriteString(fmt.Sprintf("\tapiRouter.Register(\"%s\", viewset)\n", resourcePath))
	sb.WriteString("\tapiRouter.RegisterRoutes(router)\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// %sSerializer serializes %s model\n", modelName, modelName))
	sb.WriteString(fmt.Sprintf("type %sSerializer struct {\n", modelName))
	sb.WriteString("\t*api.BaseSerializer\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// New%sSerializer creates a new serializer\n", modelName))
	sb.WriteString(fmt.Sprintf("func New%sSerializer() api.Serializer {\n", modelName))
	sb.WriteString(fmt.Sprintf("\treturn &%sSerializer{\n", modelName))
	sb.WriteString("\t\tBaseSerializer: api.NewBaseSerializer(),\n")
	sb.WriteString("\t}\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// Fields returns the fields to serialize\n"))
	sb.WriteString(fmt.Sprintf("func (s *%sSerializer) Fields() []string {\n", modelName))
	sb.WriteString("\treturn []string{\"id\"} // Add your fields here\n")
	sb.WriteString("}\n")

	if graphql {
		sb.WriteString("\n// GraphQL resolver would go here\n")
	}

	return sb.String()
}
