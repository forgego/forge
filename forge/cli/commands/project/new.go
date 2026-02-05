package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/cobra"
)

// NewCommand creates the "new" command for creating new projects
type NewCommand struct{}

// NewNewCommand creates a new instance of NewCommand
func NewNewCommand() *NewCommand {
	return &NewCommand{}
}

// Definition returns the cobra command definition
func (c *NewCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Forge project",
		Long:  "Create a new Forge project with the specified name. Choose between Simple (default) or Advanced template.",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("path", "", "Path where project will be created (default: project-name)")
	cmd.Flags().String("template", "", "Project template: simple or advanced (default: simple)")
	cmd.Flags().String("database", "", "Database type: postgres, mysql, or sqlite (default: postgres)")
	cmd.Flags().Bool("docker", false, "Include Docker setup (Dockerfile and compose.yaml)")
	return cmd
}

// Execute runs the command logic
func (c *NewCommand) Execute(ctx *core.Context, args []string) error {
	projectName := args[0]

	// Get path flag
	projectPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get path flag: %w", err)
	}
	if projectPath == "" {
		projectPath = projectName
	}

	// Check if directory already exists
	if _, err := os.Stat(projectPath); err == nil {
		return fmt.Errorf("directory %s already exists", projectPath)
	}

	// Interactive prompts (if not provided via flags)
	var templateChoice string
	templateFlag, _ := ctx.Cmd.Flags().GetString("template")
	if templateFlag == "" {
		prompt := &survey.Select{
			Message: "Choose project template:",
			Options: []string{"simple", "advanced"},
			Default: "simple",
			Help:    "Simple: Django-like app structure (recommended for most projects)\nAdvanced: Includes domain/infra layers for clean architecture",
		}
		if err := survey.AskOne(prompt, &templateChoice); err != nil {
			return fmt.Errorf("failed to get template choice: %w", err)
		}
	} else {
		templateChoice = templateFlag
	}

	var databaseType string
	dbFlag, _ := ctx.Cmd.Flags().GetString("database")
	if dbFlag == "" {
		prompt := &survey.Select{
			Message: "Choose database type:",
			Options: []string{"postgres", "mysql", "sqlite"},
			Default: "postgres",
		}
		if err := survey.AskOne(prompt, &databaseType); err != nil {
			return fmt.Errorf("failed to get database choice: %w", err)
		}
	} else {
		databaseType = dbFlag
	}

	var includeDocker bool
	dockerFlag, _ := ctx.Cmd.Flags().GetBool("docker")
	if !dockerFlag {
		prompt := &survey.Confirm{
			Message: "Include Docker setup?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &includeDocker); err != nil {
			return fmt.Errorf("failed to get docker choice: %w", err)
		}
	} else {
		includeDocker = true
	}

	// Determine template
	var projectTemplate templates.ProjectTemplate
	if templateChoice == "advanced" {
		projectTemplate = templates.TemplateAdvanced
	} else {
		projectTemplate = templates.TemplateSimple
	}

	// Create project structure
	if err := createProjectStructure(projectPath, projectName, projectTemplate, databaseType, includeDocker); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	fmt.Printf("✓ Created new Forge project: %s\n", projectName)
	fmt.Printf("  Directory: %s\n", projectPath)
	fmt.Printf("  Template: %s\n", templateChoice)
	fmt.Printf("  Database: %s\n", databaseType)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectPath)
	fmt.Printf("  forge generate\n")
	fmt.Printf("  forge migrate\n")
	fmt.Printf("  forge runserver\n")

	return nil
}

// createProjectStructure creates the project structure based on template
func createProjectStructure(projectPath, projectName string, template templates.ProjectTemplate, databaseType string, includeDocker bool) error {
	structure := templates.GetStructure(template)

	// Create directories
	for _, dir := range structure.Directories {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create additional directories
	additionalDirs := []string{"static", "templates"}
	for _, dir := range additionalDirs {
		fullPath := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Prepare template data
	templateData := templates.TemplateData{
		ProjectName:  projectName,
		DatabaseType: databaseType,
	}

	// Render and write files from structure
	for _, file := range structure.Files {
		fullPath := filepath.Join(projectPath, file.Path)
		if err := templates.WriteTemplateFile(fullPath, file.Template, templateData, file.Permissions); err != nil {
			return fmt.Errorf("failed to create file %s: %w", file.Path, err)
		}
	}

	// Create config.yaml with database settings
	if err := createConfigFile(projectPath, databaseType); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Create Docker files if requested
	if includeDocker {
		if err := createDockerFiles(projectPath, projectName); err != nil {
			return fmt.Errorf("failed to create docker files: %w", err)
		}
	}

	// Create .forge directory for CLI config
	forgeDir := filepath.Join(projectPath, ".forge")
	if err := os.MkdirAll(forgeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .forge directory: %w", err)
	}

	// Create .forge/config.yaml
	forgeConfig := fmt.Sprintf(`# Forge CLI Configuration
default_database: %s
default_template: %s
`, databaseType, template)
	if err := os.WriteFile(filepath.Join(forgeDir, "config.yaml"), []byte(forgeConfig), 0644); err != nil {
		return fmt.Errorf("failed to create .forge/config.yaml: %w", err)
	}

	return nil
}

// createConfigFile creates the config.yaml file with database settings
func createConfigFile(projectPath, databaseType string) error {
	var dbConfig string
	switch databaseType {
	case "mysql":
		dbConfig = `database:
  driver: mysql
  host: localhost
  port: 3306
  user: root
  password: root
  dbname: forge_db
`
	case "sqlite":
		dbConfig = `database:
  driver: sqlite3
  dsn: ./forge.db
`
	default: // postgres
		dbConfig = `database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: forge_db
  sslmode: disable
`
	}

	configContent := dbConfig + `
server:
  host: localhost
  port: 8000
  read_timeout: 30
  write_timeout: 30

admin:
  enabled: true
  path: /admin/

security:
  session_secret: change-me-in-production
  csrf_secret_key: change-me-in-production
`

	configPath := filepath.Join(projectPath, "config", "config.yaml")
	return os.WriteFile(configPath, []byte(configContent), 0644)
}

// createDockerFiles creates Dockerfile and compose.yaml
func createDockerFiles(projectPath, projectName string) error {
	templateData := templates.TemplateData{
		ProjectName: projectName,
	}

	// Create Dockerfile
	dockerfilePath := filepath.Join(projectPath, "Dockerfile")
	if err := templates.WriteTemplateFile(dockerfilePath, "dockerfile.tmpl", templateData, 0644); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	// Create compose.yaml
	composePath := filepath.Join(projectPath, "compose.yaml")
	if err := templates.WriteTemplateFile(composePath, "compose.tmpl", templateData, 0644); err != nil {
		return fmt.Errorf("failed to create compose.yaml: %w", err)
	}

	return nil
}
