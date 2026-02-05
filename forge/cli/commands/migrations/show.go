package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/codegen"
	dbstate "github.com/forgego/forge/db/migrate/state"
	"github.com/forgego/forge/db/migrate"
	"github.com/spf13/cobra"
)

// ShowCommand creates the migration show command
type ShowCommand struct{}

// NewShowCommand creates a new instance of ShowCommand
func NewShowCommand() *ShowCommand {
	return &ShowCommand{}
}

// Definition returns the cobra command definition
func (c *ShowCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [version]",
		Short: "Show migration plan or specific migration",
		Long:  "Display the migration plan that would be generated, or show a specific migration file",
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().String("models", "./models", "Directory containing model definitions")
	cmd.Flags().Bool("sql", false, "Show raw SQL (for specific migration, shows both up and down)")
	cmd.Flags().Bool("verbose", false, "Enable verbose output including parse errors")
	return cmd
}

// Execute runs the command logic
func (c *ShowCommand) Execute(ctx *core.Context, args []string) error {
	// Get migrations path
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	// Get models directory
	modelsDir, err := ctx.Cmd.Flags().GetString("models")
	if err != nil {
		return fmt.Errorf("failed to get models flag: %w", err)
	}
	if modelsDir == "" {
		modelsDir = "./models"
	}

	// Check if --sql flag is set
	showSQL, err := ctx.Cmd.Flags().GetBool("sql")
	if err != nil {
		return fmt.Errorf("failed to get sql flag: %w", err)
	}

	// If version specified, show that migration
	if len(args) > 0 {
		version := args[0]
		pattern := filepath.Join(migrationsPath, version+"_*.up.sql")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to find migration: %w", err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("migration %s not found", version)
		}

		upFile := matches[0]
		upContent, err := os.ReadFile(upFile)
		if err != nil {
			return fmt.Errorf("failed to read migration: %w", err)
		}

		if showSQL {
			// Show both up and down SQL
			fmt.Printf("=== Up Migration: %s ===\n\n", filepath.Base(upFile))
			fmt.Println(string(upContent))

			// Try to find corresponding down migration
			downFile := strings.Replace(upFile, ".up.sql", ".down.sql", 1)
			if downContent, err := os.ReadFile(downFile); err == nil {
				fmt.Printf("\n=== Down Migration: %s ===\n\n", filepath.Base(downFile))
				fmt.Println(string(downContent))
			} else {
				fmt.Printf("\n(No down migration found)\n")
			}
		} else {
			// Show file content as-is
			fmt.Printf("Migration: %s\n\n", filepath.Base(upFile))
			fmt.Println(string(upContent))
		}
		return nil
	}

	// Otherwise, show what migration would be generated
	// Parse current models
	parser := generator.NewASTParser()
	currentDefs, err := parser.ParseDirectory(modelsDir)
	if err != nil {
		return fmt.Errorf("failed to parse models: %w", err)
	}

	if len(currentDefs) == 0 {
		fmt.Println("No model definitions found")
		return nil
	}

	// Get verbose flag
	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")

	// Load previous state
	state, err := migrate.LoadState(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// If verbose, also show parse errors from state loader
	if verbose {
		loader := dbstate.NewFileStateLoaderWithOptions(migrationsPath, dbstate.LoaderOptions{Verbose: true})
		_, loadErr := loader.Load()
		if loadErr == nil {
			// Show parse errors if any
			if fileLoader, ok := loader.(*dbstate.FileStateLoader); ok {
				parseErrors := fileLoader.GetParseErrors()
				if len(parseErrors) > 0 {
					fmt.Println("⚠️  Parse Warnings:")
					for _, perr := range parseErrors {
						fmt.Printf("  %s", perr.File)
						if perr.Line > 0 {
							fmt.Printf(":%d", perr.Line)
						}
						fmt.Printf(": %s\n", perr.Message)
					}
					fmt.Println()
				}
			}
		}
	}

	// Convert state to model definitions
	previousDefs := state.ToModelDefinitions()

	// Detect changes
	changes, err := migrate.DetectChanges(currentDefs, previousDefs)
	if err != nil {
		return fmt.Errorf("failed to detect changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("No changes detected - no migration would be generated")
		return nil
	}

	// Generate SQL preview
	driver := ctx.Config.GetDriver()
	sqlGen, err := migrate.NewSQLGenerator(driver)
	if err != nil {
		return fmt.Errorf("failed to create SQL generator: %w", err)
	}

	upSQL, err := sqlGen.GenerateUpSQL(changes)
	if err != nil {
		return fmt.Errorf("failed to generate SQL: %w", err)
	}

	downSQL, err := sqlGen.GenerateDownSQL(changes)
	if err != nil {
		return fmt.Errorf("failed to generate down SQL: %w", err)
	}

	if showSQL {
		fmt.Println("=== Up Migration SQL ===")
		fmt.Println(upSQL)
		fmt.Println("\n=== Down Migration SQL ===")
		fmt.Println(downSQL)
	} else {
		fmt.Println("Migration Plan Preview:")
		fmt.Printf("  Changes detected: %d\n", len(changes))
		fmt.Println("\n  Up Migration SQL:")
		fmt.Println("  " + strings.ReplaceAll(upSQL, "\n", "\n  "))
		fmt.Println("\n  (Use --sql flag to see full SQL including down migration)")
	}

	return nil
}
