package migrations

import (
	"context"
	"fmt"
	"strconv"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/spf13/cobra"
)

// ForceCommand creates the force command
type ForceCommand struct{}

// NewForceCommand creates a new instance of ForceCommand
func NewForceCommand() *ForceCommand {
	return &ForceCommand{}
}

// Definition returns the cobra command definition
func (c *ForceCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "force <version>",
		Short: "Force set migration version (for dirty state recovery)",
		Long:  "Force set a migration version and mark it as clean. Use with caution - only after manually fixing a failed migration.",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	return cmd
}

// Execute runs the command logic
func (c *ForceCommand) Execute(ctx *core.Context, args []string) error {
	versionStr := args[0]

	// Parse version
	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version: %s (must be a number)", versionStr)
	}

	// Get migrations path
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	// Connect to database
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}
	defer runner.Close()

	// Get current status
	cmdCtx := context.Background()
	currentStatus, err := runner.Status(cmdCtx)
	if err != nil {
		return fmt.Errorf("failed to get current status: %w", err)
	}

	// Safety check: warn if not dirty
	if !currentStatus.Dirty {
		fmt.Println("⚠️  WARNING: Database is not in a dirty state.")
		fmt.Println("   Force command should only be used after manually fixing a failed migration.")
		fmt.Print("   Continue anyway? (yes/no): ")

		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Force set version using golang-migrate's Force method
	// We need to access the underlying migrate instance
	// For now, we'll add a method to MigrationRunner
	fmt.Printf("⚠️  Forcing migration version to %d...\n", version)
	fmt.Println("   This will mark the database as clean at this version.")
	fmt.Println("   Make sure you have manually fixed any issues before proceeding.")

	// Force set version
	if err := runner.Force(cmdCtx, uint(version)); err != nil {
		return fmt.Errorf("failed to force version: %w", err)
	}

	fmt.Printf("✓ Migration version forced to %d (marked as clean)\n", version)
	return nil
}
