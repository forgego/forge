package project

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/cli/templates"
	"github.com/spf13/cobra"
)

func generateHexSecret() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

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
	cmd.Flags().String("database", "", "Database type: postgres or sqlite (default: postgres)")
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
			Options: []string{"postgres", "sqlite"},
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
	fmt.Printf("\nSecrets: generated into %s/.env (not committed; listed in .gitignore).\n", projectPath)
	fmt.Printf("  In production, supply FORGE_SECURITY_SECRET_KEY, FORGE_SECURITY_SESSION_SECRET\n")
	fmt.Printf("  and FORGE_SECURITY_CSRF_SECRET_KEY from the environment or a secrets manager.\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectPath)
	fmt.Printf("  forge generate\n")
	fmt.Printf("  forge makemigrations\n")
	fmt.Printf("  forge migrate up\n")
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

	// Create config.yaml with database settings and a separate local
	// secrets file (see createConfigFile). The secrets must never be
	// committed; config.yaml references them through the environment.
	if err := createConfigFile(projectPath, databaseType); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	// Write .gitignore so the local secrets file (and build output) are
	// not committed.
	if err := createGitignore(projectPath, projectName); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
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

// createConfigFile creates the config.yaml file with database settings.
// Secret values are never written to config.yaml: three fresh secrets are
// generated into a local .env file (excluded from git via .gitignore) and
// config.yaml references them through environment variables, which
// forge/config binds (FORGE_SECURITY_SECRET_KEY and friends).
func createConfigFile(projectPath, databaseType string) error {
	var dbConfig string
	switch databaseType {
	case "sqlite", "sqlite3":
		dbConfig = `database:
  driver: sqlite3
  name: ./forge.db
`
	case "postgres", "postgresql", "":
		dbConfig = `database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  name: forge_db
  sslmode: disable
`
	default:
		return fmt.Errorf("unsupported database type: %s", databaseType)
	}

	secretKey, err := generateHexSecret()
	if err != nil {
		return fmt.Errorf("failed to generate security.secret_key: %w", err)
	}
	sessionSecret, err := generateHexSecret()
	if err != nil {
		return fmt.Errorf("failed to generate security.session_secret: %w", err)
	}
	csrfSecret, err := generateHexSecret()
	if err != nil {
		return fmt.Errorf("failed to generate security.csrf_secret_key: %w", err)
	}

	configContent := dbConfig + `
server:
  host: localhost
  port: 8000
  read_timeout: 30
  write_timeout: 30

admin:
  enabled: true
  path: /admin

security:
  # Secrets are NOT stored here. They are generated into .env (local use
  # only, never committed) and read from the environment:
  #   FORGE_SECURITY_SECRET_KEY
  #   FORGE_SECURITY_SESSION_SECRET
  #   FORGE_SECURITY_CSRF_SECRET_KEY
  csrf_exempt_paths:
    - /admin/api
    - /api
`

	configPath := filepath.Join(projectPath, "config", "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return err
	}

	envContent := fmt.Sprintf(`# Local development secrets generated by "forge new". DO NOT COMMIT.
# In production, supply these from the environment or a secrets manager instead.
FORGE_SECURITY_SECRET_KEY=%s
FORGE_SECURITY_SESSION_SECRET=%s
FORGE_SECURITY_CSRF_SECRET_KEY=%s
`, secretKey, sessionSecret, csrfSecret)

	return os.WriteFile(filepath.Join(projectPath, ".env"), []byte(envContent), 0600)
}

// createGitignore writes a .gitignore that excludes the local secrets file
// and usual build output.
func createGitignore(projectPath, projectName string) error {
	content := `# Local environment with generated secrets (created by "forge new"). Do not commit.
.env

# Build output
/` + projectName + `
*.exe
*.test
*.out
dist/
`
	return os.WriteFile(filepath.Join(projectPath, ".gitignore"), []byte(content), 0644)
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
