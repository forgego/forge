package migrations

import (
	"fmt"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db/migrate/generate"
	"github.com/spf13/cobra"
)

// SquashCommand creates the squash command
type SquashCommand struct{}

// NewSquashCommand creates a new instance of SquashCommand
func NewSquashCommand() *SquashCommand {
	return &SquashCommand{}
}

// Definition returns the cobra command definition
func (c *SquashCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "squash <start_version> <end_version> <name>",
		Short: "Squash migrations into a single migration",
		Long:  "Combine multiple migrations from start_version to end_version into a single migration. The old migrations will be marked as replaced.",
		Args:  cobra.ExactArgs(3),
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	return cmd
}

// Execute runs the command logic
func (c *SquashCommand) Execute(ctx *core.Context, args []string) error {
	startVersion := args[0]
	endVersion := args[1]
	name := args[2]

	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	squasher := generate.NewSquasher(migrationsPath)
	if err := squasher.SquashMigrations(startVersion, endVersion, name); err != nil {
		return fmt.Errorf("failed to squash migrations: %w", err)
	}

	fmt.Printf("✓ Successfully squashed migrations from %s to %s into %s\n", startVersion, endVersion, name)
	fmt.Println("  Note: Old migrations should be archived, not deleted")
	fmt.Println("  The new migration includes a 'replaces' field listing the squashed migrations")
	return nil
}
