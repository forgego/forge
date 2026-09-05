package cli

import (
	"github.com/forgego/forge/cli/commands"
	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// BuildRootCommand builds and returns the root cobra command with all registered commands
func BuildRootCommand() *cobra.Command {
	// Register all commands
	commands.RegisterAllCommands()

	// Build and return the root command
	registry := core.GetRegistry()
	return registry.BuildRootCommand()
}

