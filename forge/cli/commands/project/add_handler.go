package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/cobra"
)

// AddHandlerCommand creates the "add handler" command
type AddHandlerCommand struct{}

// NewAddHandlerCommand creates a new instance of AddHandlerCommand
func NewAddHandlerCommand() *AddHandlerCommand {
	return &AddHandlerCommand{}
}

// Definition returns the cobra command definition
func (c *AddHandlerCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "handler [handler-name]",
		Short: "Add a new HTTP handler to an app",
		Long:  "Add a new HTTP handler function to an app's handlers.go file",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("app", "", "App name (auto-detected if in app directory)")
	cmd.Flags().String("method", "GET", "HTTP method (GET, POST, PUT, DELETE, PATCH)")
	cmd.Flags().String("path", "", "URL path for the handler")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	return cmd
}

// Execute runs the command logic
func (c *AddHandlerCommand) Execute(ctx *core.Context, args []string) error {
	handlerName := args[0]

	// Detect project root and app
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	appName, err := detectAppName(ctx, projectRoot)
	if err != nil {
		return fmt.Errorf("failed to detect app name: %w", err)
	}

	// Get method
	method, _ := ctx.Cmd.Flags().GetString("method")
	if method == "" {
		if err := survey.AskOne(&survey.Select{
			Message: "HTTP method:",
			Options: []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
			Default: "GET",
		}, &method); err != nil {
			return err
		}
	}

	// Get path
	path, _ := ctx.Cmd.Flags().GetString("path")
	if path == "" {
		if err := survey.AskOne(&survey.Input{
			Message: "URL path:",
			Default: "/" + strings.ToLower(handlerName),
		}, &path); err != nil {
			return err
		}
	}

	// Check dry-run
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Printf("Dry-run mode - would create handler %s in app %s\n", handlerName, appName)
		fmt.Printf("Method: %s, Path: %s\n", method, path)
		return nil
	}

	// Prepare template data
	templateData := templates.TemplateData{
		AppName:     appName,
		HandlerName: handlerName,
		Method:      method,
		HandlerPath: path,
	}

	// Create or append to handlers.go
	appPath := filepath.Join(projectRoot, "app", appName)
	handlersPath := filepath.Join(appPath, "handlers.go")

	// Check if handlers.go exists
	if _, err := os.Stat(handlersPath); os.IsNotExist(err) {
		// Create new handlers.go file
		packageDecl := fmt.Sprintf("package %s\n\n", appName)
		if err := os.WriteFile(handlersPath, []byte(packageDecl), 0644); err != nil {
			return fmt.Errorf("failed to create handlers.go: %w", err)
		}
	}

	// Render handler code
	handlerCode, err := templates.RenderTemplate("handlers.go.tmpl", templateData)
	if err != nil {
		return fmt.Errorf("failed to render handler template: %w", err)
	}

	// Append to handlers.go
	file, err := os.OpenFile(handlersPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open handlers.go: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString("\n" + string(handlerCode)); err != nil {
		return fmt.Errorf("failed to write handler: %w", err)
	}

	fmt.Printf("✓ Added handler %s to app %s\n", handlerName, appName)
	fmt.Printf("  Method: %s, Path: %s\n", method, path)

	return nil
}
