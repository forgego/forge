package migrations

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db/migrate"
	"github.com/spf13/cobra"
)

// MakeMigrationsCommand creates the makemigrations command
type MakeMigrationsCommand struct{}

// NewMakeMigrationsCommand creates a new instance of MakeMigrationsCommand
func NewMakeMigrationsCommand() *MakeMigrationsCommand {
	return &MakeMigrationsCommand{}
}

// Definition returns the cobra command definition
func (c *MakeMigrationsCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "makemigrations <name>",
		Short: "Create new migration files",
		Long:  "Create migration files. Use --auto to generate SQL from models, otherwise creates empty files. Uses sequential versioning (e.g., 000001_name.up.sql)",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().String("models", "./models", "Directory containing model definitions (used with --auto)")
	cmd.Flags().Bool("auto", false, "Generate SQL from models automatically")
	cmd.Flags().Bool("empty", false, "Create an empty migration file")
	cmd.Flags().Bool("merge", false, "Enable fixing of migration conflicts")
	cmd.Flags().Bool("verbose", false, "Enable verbose output including parse errors and warnings")
	return cmd
}

// Execute runs the command logic
func (c *MakeMigrationsCommand) Execute(ctx *core.Context, args []string) error {
	migrationName := args[0]

	// Validate migration name
	if migrationName == "" {
		return fmt.Errorf("migration name cannot be empty")
	}
	// Check for invalid filename characters
	if strings.ContainsAny(migrationName, "/\\<>:\"|?*") {
		return fmt.Errorf("migration name contains invalid characters")
	}

	migrationsDir, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	// Check if --merge flag is set
	merge, err := ctx.Cmd.Flags().GetBool("merge")
	if err != nil {
		return fmt.Errorf("failed to get merge flag: %w", err)
	}

	if merge {
		// Handle merge conflicts
		conflicts, err := detectConflicts(migrationsDir)
		if err != nil {
			return fmt.Errorf("failed to detect conflicts: %w", err)
		}

		if len(conflicts) == 0 {
			fmt.Println("No migration conflicts detected")
			return nil
		}

		// Create merge migration for first conflict
		// In practice, you might want to handle all conflicts
		conflict := conflicts[0]
		mergeName := fmt.Sprintf("merge_%s_%s", conflict.Version1, conflict.Version2)
		upPath, downPath, err := createMergeMigration(migrationsDir, mergeName, conflict)
		if err != nil {
			return fmt.Errorf("failed to create merge migration: %w", err)
		}

		fmt.Printf("✓ Created merge migration to resolve conflict:\n")
		fmt.Printf("  %s\n", upPath)
		fmt.Printf("  %s\n", downPath)
		fmt.Printf("  Resolves conflict between %s and %s\n", conflict.Version1, conflict.Version2)
		return nil
	}

	// Check if --empty flag is set
	empty, err := ctx.Cmd.Flags().GetBool("empty")
	if err != nil {
		return fmt.Errorf("failed to get empty flag: %w", err)
	}

	// Check if --auto flag is set
	auto, err := ctx.Cmd.Flags().GetBool("auto")
	if err != nil {
		return fmt.Errorf("failed to get auto flag: %w", err)
	}

	if auto {
		// Generate SQL from models
		modelsDir, err := ctx.Cmd.Flags().GetString("models")
		if err != nil {
			return fmt.Errorf("failed to get models flag: %w", err)
		}
		if modelsDir == "" {
			modelsDir = "./models"
		}

		gen, err := migrate.NewGenerator(modelsDir, migrationsDir)
		if err != nil {
			return fmt.Errorf("failed to create migration generator: %w", err)
		}

		if err := gen.GenerateMigrations(migrationName); err != nil {
			return fmt.Errorf("failed to generate migrations: %w", err)
		}

		fmt.Printf("✓ Generated migration files with SQL from models:\n")
		fmt.Printf("  %s/%s.up.sql\n", migrationsDir, migrationName)
		fmt.Printf("  %s/%s.down.sql\n", migrationsDir, migrationName)
	} else if empty {
		// Create empty migration files
		upPath, downPath, err := createMigrationFiles(migrationsDir, migrationName)
		if err != nil {
			return fmt.Errorf("failed to create migration files: %w", err)
		}

		fmt.Printf("✓ Created empty migration files:\n")
		fmt.Printf("  %s\n", upPath)
		fmt.Printf("  %s\n", downPath)
	} else {
		// Default: create empty migration files
		upPath, downPath, err := createMigrationFiles(migrationsDir, migrationName)
		if err != nil {
			return fmt.Errorf("failed to create migration files: %w", err)
		}

		fmt.Printf("✓ Created migration files:\n")
		fmt.Printf("  %s\n", upPath)
		fmt.Printf("  %s\n", downPath)
	}

	return nil
}

