package generation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/cli/core"
	codegen "github.com/forgego/forge/codegen"
	"github.com/spf13/cobra"
)

// GenerateCommand creates the generate command
type GenerateCommand struct{}

// NewGenerateCommand creates a new instance of GenerateCommand
func NewGenerateCommand() *GenerateCommand {
	return &GenerateCommand{}
}

// Definition returns the cobra command definition
func (c *GenerateCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate code from schema definitions",
		Long:  "Parse schema definitions and generate type-safe models, managers, and querysets",
	}
	cmd.Flags().String("models", "./models", "Directory containing schema definitions")
	cmd.Flags().String("output", "./models", "Output directory for generated code")
	return cmd
}

// Execute runs the command logic
func (c *GenerateCommand) Execute(ctx *core.Context, args []string) error {
	modelsDir, err := ctx.Cmd.Flags().GetString("models")
	if err != nil {
		return fmt.Errorf("failed to get models flag: %w", err)
	}
	if modelsDir == "" {
		// Default to "app" if it exists, otherwise "models"
		if _, err := os.Stat("app"); err == nil {
			modelsDir = "app"
		} else {
			modelsDir = "models"
		}
	}

	outputDir, err := ctx.Cmd.Flags().GetString("output")
	if err != nil {
		return fmt.Errorf("failed to get output flag: %w", err)
	}

	// If outputDir is not specified and modelsDir is "app", default to "app" as well
	if outputDir == "" && modelsDir == "app" {
		outputDir = "app"
	} else if outputDir == "" {
		outputDir = modelsDir
	}

	// If scanning "app/" directory, look for submodules
	if modelsDir == "app" || strings.HasSuffix(modelsDir, "/app") {
		fmt.Printf("Scanning apps in %s...\n", modelsDir)
		entries, err := os.ReadDir(modelsDir)
		if err != nil {
			return fmt.Errorf("failed to read app directory: %w", err)
		}

		generatedCount := 0
		for _, entry := range entries {
			if entry.IsDir() {
				appPath := filepath.Join(modelsDir, entry.Name())
				// Check for models.go
				if _, err := os.Stat(filepath.Join(appPath, "models.go")); err == nil {
					// Found an app with models
					targetOutput := outputDir
					if targetOutput == "" || targetOutput == "models" {
						targetOutput = appPath // Colocate generated files
					} else {
						targetOutput = filepath.Join(targetOutput, entry.Name())
					}

					fmt.Printf("  Generating for %s...\n", entry.Name())
					gen := codegen.NewGenerator(appPath, targetOutput)
					if err := gen.Generate(); err != nil {
						return fmt.Errorf("generation failed for %s: %w", entry.Name(), err)
					}
					generatedCount++
				}
			}
		}

		if generatedCount > 0 {
			fmt.Printf("✓ Generated code for %d apps\n", generatedCount)
			return nil
		}
		// If no apps found, fall through to single directory scan
	}

	if outputDir == "" {
		outputDir = modelsDir
	}

	// Create generator and run for single directory
	gen := codegen.NewGenerator(modelsDir, outputDir)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("✓ Generated code from %s to %s\n", modelsDir, outputDir)
	return nil
}
