package migrations

import (
	"fmt"
	"path/filepath"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db/migrate/state"
	"github.com/forgego/forge/db/migrate/verify"
	"github.com/spf13/cobra"
)

// LintCommand creates the migration lint command
type LintCommand struct{}

// NewLintCommand creates a new instance of LintCommand
func NewLintCommand() *LintCommand {
	return &LintCommand{}
}

// Definition returns the cobra command definition
func (c *LintCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint",
		Short: "Lint migration files",
		Long:  "Check migration files for common issues and best practices",
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("verbose", false, "Enable verbose output including parse errors")
	return cmd
}

// Execute runs the command logic
func (c *LintCommand) Execute(ctx *core.Context, args []string) error {
	// Get migrations path
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	// Get verbose flag
	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")

	// Create linter
	linter := verify.NewLinter()

	// Lint all migrations
	results, err := linter.LintMigrationsDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to lint migrations: %w", err)
	}

	// If verbose, also show parse errors from state loader
	if verbose {
		// Load state with verbose mode to get parse errors
		loader := state.NewFileStateLoaderWithOptions(migrationsPath, state.LoaderOptions{Verbose: true})
		_, err := loader.Load()
		if err == nil {
			// Get parse errors if available
			if fileLoader, ok := loader.(*state.FileStateLoader); ok {
				parseErrors := fileLoader.GetParseErrors()
				if len(parseErrors) > 0 {
					fmt.Println("\n📋 Parse Errors (from state loader):")
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

	// Group results by level
	errors := []verify.LintResult{}
	warnings := []verify.LintResult{}
	infos := []verify.LintResult{}

	for _, result := range results {
		switch result.Level {
		case "error":
			errors = append(errors, result)
		case "warning":
			warnings = append(warnings, result)
		case "info":
			infos = append(infos, result)
		}
	}

	// Print results
	if len(errors) > 0 {
		fmt.Println("❌ Errors:")
		for _, result := range errors {
			fmt.Printf("  %s: %s\n", filepath.Base(result.File), result.Message)
		}
		fmt.Println()
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, result := range warnings {
			fmt.Printf("  %s: %s\n", filepath.Base(result.File), result.Message)
		}
		fmt.Println()
	}

	if len(infos) > 0 {
		fmt.Println("ℹ️  Info:")
		for _, result := range infos {
			fmt.Printf("  %s: %s\n", filepath.Base(result.File), result.Message)
		}
		fmt.Println()
	}

	if len(results) == 0 {
		fmt.Println("✓ No issues found")
	} else {
		fmt.Printf("Summary: %d errors, %d warnings, %d info messages\n",
			len(errors), len(warnings), len(infos))

		if len(errors) > 0 {
			return fmt.Errorf("linting found %d error(s)", len(errors))
		}
	}

	return nil
}
