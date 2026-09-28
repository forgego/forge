package migrations

import (
	"context"
	"fmt"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/spf13/cobra"
)

// RollbackCommand creates the rollback command
type RollbackCommand struct{}

// NewRollbackCommand creates a new instance of RollbackCommand
func NewRollbackCommand() *RollbackCommand {
	return &RollbackCommand{}
}

// Definition returns the cobra command definition
func (c *RollbackCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback last migration",
		Long:  "Rollback the last applied migration",
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	return cmd
}

// Execute runs the command logic
func (c *RollbackCommand) Execute(ctx *core.Context, args []string) error {
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

	// Rollback migration
	cmdCtx := context.Background()
	if err := runner.Rollback(cmdCtx); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("✓ Migration rolled back successfully")
	return nil
}
