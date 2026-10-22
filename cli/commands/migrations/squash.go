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
		Short: "Migration squashing is not implemented yet",
		Long:  "Migration squashing is not implemented yet.",
		Args:  cobra.ExactArgs(3),
	}
	cmd.Flags().String("path", "./migrations", "Reserved migration directory path (squash is not implemented yet)")
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
	return squasher.SquashMigrations(startVersion, endVersion, name)
}
