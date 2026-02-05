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

// AddServiceCommand creates the "add service" command
type AddServiceCommand struct{}

// NewAddServiceCommand creates a new instance of AddServiceCommand
func NewAddServiceCommand() *AddServiceCommand {
	return &AddServiceCommand{}
}

// Definition returns the cobra command definition
func (c *AddServiceCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service [service-name]",
		Short: "Add a new service to an app or domain",
		Long:  "Add a new service (business logic layer) to an app or domain",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("app", "", "App name (for app-level service)")
	cmd.Flags().String("domain", "", "Domain name (for domain-level service)")
	cmd.Flags().Bool("inject", false, "Auto-wire dependencies")
	cmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	return cmd
}

// Execute runs the command logic
func (c *AddServiceCommand) Execute(ctx *core.Context, args []string) error {
	serviceName := args[0]

	// Detect project root
	projectRoot, err := detectProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to detect project root: %w", err)
	}

	// Determine if app or domain service
	appFlag, _ := ctx.Cmd.Flags().GetString("app")
	domainFlag, _ := ctx.Cmd.Flags().GetString("domain")

	var serviceType string
	var packageName string
	var filePath string

	if domainFlag != "" {
		serviceType = "domain"
		packageName = domainFlag
		filePath = filepath.Join(projectRoot, "domain", domainFlag, "service.go")
	} else if appFlag != "" {
		serviceType = "app"
		packageName = appFlag
		appPath := filepath.Join(projectRoot, "app", appFlag)
		filePath = filepath.Join(appPath, "services.go")
	} else {
		// Prompt user
		if err := survey.AskOne(&survey.Select{
			Message: "Service type:",
			Options: []string{"app", "domain"},
			Default: "app",
		}, &serviceType); err != nil {
			return err
		}

		if serviceType == "domain" {
			if err := survey.AskOne(&survey.Input{
				Message: "Domain name:",
			}, &packageName); err != nil {
				return err
			}
			filePath = filepath.Join(projectRoot, "domain", packageName, "service.go")
		} else {
			if err := survey.AskOne(&survey.Input{
				Message: "App name:",
			}, &packageName); err != nil {
				return err
			}
			appPath := filepath.Join(projectRoot, "app", packageName)
			filePath = filepath.Join(appPath, "services.go")
		}
	}

	inject, _ := ctx.Cmd.Flags().GetBool("inject")
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")

	if dryRun {
		fmt.Printf("Dry-run mode - would create service %s\n", serviceName)
		fmt.Printf("Type: %s, Package: %s\n", serviceType, packageName)
		if inject {
			fmt.Printf("Auto-wiring: enabled\n")
		}
		return nil
	}

	// Note: templateData not used in this implementation, using direct code generation

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate service code
	serviceCode := generateServiceCode(packageName, serviceName, inject)

	// Append or create file
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open service file: %w", err)
	}
	defer file.Close()

	// Check if file is empty, add package declaration
	info, _ := file.Stat()
	if info.Size() == 0 {
		file.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	}

	if _, err := file.WriteString("\n" + serviceCode); err != nil {
		return fmt.Errorf("failed to write service code: %w", err)
	}

	fmt.Printf("✓ Added service %s\n", serviceName)
	fmt.Printf("  Type: %s, Package: %s\n", serviceType, packageName)

	return nil
}

// generateServiceCode generates the service code
func generateServiceCode(packageName, serviceName string, inject bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("// %s provides business logic for %s\n", serviceName, serviceName))
	sb.WriteString(fmt.Sprintf("type %s interface {\n", serviceName))
	sb.WriteString("\t// Define your service methods here\n")
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// %sImpl implements %s\n", serviceName, serviceName))
	sb.WriteString(fmt.Sprintf("type %sImpl struct {\n", serviceName))
	if inject {
		sb.WriteString("\t// Dependencies will be auto-wired here\n")
	}
	sb.WriteString("}\n\n")

	sb.WriteString(fmt.Sprintf("// New%s creates a new %s instance\n", serviceName, serviceName))
	sb.WriteString(fmt.Sprintf("func New%s() %s {\n", serviceName, serviceName))
	sb.WriteString(fmt.Sprintf("\treturn &%sImpl{}\n", serviceName))
	sb.WriteString("}\n")

	return sb.String()
}
