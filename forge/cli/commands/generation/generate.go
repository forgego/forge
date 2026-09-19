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
	cmd.Flags().Bool("api", false, "Generate REST API ViewSets, Serializers, and routes")
	cmd.Flags().Bool("strict", false, "Fail when model expressions cannot be evaluated during generation")
	return cmd
}

// Execute runs the command logic
func (c *GenerateCommand) Execute(ctx *core.Context, args []string) error {
	strict, err := ctx.Cmd.Flags().GetBool("strict")
	if err != nil {
		return fmt.Errorf("failed to get strict flag: %w", err)
	}
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

	generateAPI, err := ctx.Cmd.Flags().GetBool("api")
	if err != nil {
		return fmt.Errorf("failed to get api flag: %w", err)
	}

	// If scanning "app/" directory, look for submodules
	if modelsDir == "app" || strings.HasSuffix(modelsDir, "/app") {
		fmt.Printf("Scanning apps in %s...\n", modelsDir)
		entries, err := os.ReadDir(modelsDir)
		if err != nil {
			return fmt.Errorf("failed to read app directory: %w", err)
		}

		generatedCount := 0
		type pendingApp struct {
			appPath      string
			targetOutput string
			name         string
		}
		var pending []pendingApp
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
					pending = append(pending, pendingApp{appPath: appPath, targetOutput: targetOutput, name: entry.Name()})
				}
			}
		}

		if len(pending) > 0 {
			if strict {
				var preDiagnostics []codegen.Diagnostic
				for _, app := range pending {
					preDiagnostics = append(preDiagnostics, collectDiagnostics(app.appPath)...)
				}
				if err := printDiagnostics(preDiagnostics, strict); err != nil {
					return err
				}
			}
			if generateAPI {
				for _, app := range pending {
					parser := codegen.NewASTParser()
					definitions, parseErr := parser.ParseDirectory(app.appPath)
					if parseErr != nil {
						return fmt.Errorf("generation failed for %s: failed to parse schemas: %w", app.name, parseErr)
					}
					if validateErr := codegen.ValidateAPIModels(definitions); validateErr != nil {
						return fmt.Errorf("generation failed for %s: failed to generate API code: %w", app.name, validateErr)
					}
				}
			}
			var diagnostics []codegen.Diagnostic
			for _, app := range pending {
				fmt.Printf("  Generating for %s...\n", app.name)
				gen := codegen.NewGenerator(app.appPath, app.targetOutput).SetGenerateAPI(generateAPI)
				if err := gen.Generate(); err != nil {
					return fmt.Errorf("generation failed for %s: %w", app.name, err)
				}
				diagnostics = append(diagnostics, gen.Diagnostics()...)
				generatedCount++
			}

			if generatedCount > 0 {
				if err := printDiagnostics(diagnostics, strict); err != nil {
					return err
				}
				fmt.Printf("✓ Generated code for %d apps\n", generatedCount)
				return nil
			}
		}

		// If no apps found, fall through to single directory scan
	}

	if outputDir == "" {
		outputDir = modelsDir
	}

	// Create generator and run for single directory.
	// In strict mode, parse first and reject diagnostics before any write so
	// a failed run cannot truncate the last valid generated file.
	if strict {
		if err := printDiagnostics(collectDiagnostics(modelsDir), strict); err != nil {
			return err
		}
	}
	gen := codegen.NewGenerator(modelsDir, outputDir).SetGenerateAPI(generateAPI)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}
	if err := printDiagnostics(gen.Diagnostics(), strict); err != nil {
		return err
	}

	fmt.Printf("✓ Generated code from %s to %s\n", modelsDir, outputDir)
	return nil
}

func printDiagnostics(diagnostics []codegen.Diagnostic, strict bool) error {
	for _, diagnostic := range diagnostics {
		fmt.Fprintf(os.Stderr, "warning: %s\n", diagnostic)
	}
	if strict && len(diagnostics) > 0 {
		return fmt.Errorf("generation found %d model expressions that cannot be evaluated; see warnings", len(diagnostics))
	}
	return nil
}

// collectDiagnostics parses models without writing any output, so strict
// runs can fail before truncating previously generated files.
func collectDiagnostics(modelsDir string) []codegen.Diagnostic {
	parser := codegen.NewASTParser()
	if _, err := parser.ParseDirectory(modelsDir); err != nil {
		return nil
	}
	return parser.Diagnostics()
}
