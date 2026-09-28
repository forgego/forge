package project

import (
	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// AddGroup represents the "add" command group
type AddGroup struct{}

// NewAddGroup creates a new instance of AddGroup
func NewAddGroup() *AddGroup {
	return &AddGroup{}
}

// Definition returns the cobra command definition
func (g *AddGroup) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add components to your project",
		Long:  "Add apps, models, handlers, APIs, and services to your Forge project",
	}

	// Add subcommands
	cmd.AddCommand(NewAddAppCommand().Definition())
	cmd.AddCommand(NewAddModelCommand().Definition())
	cmd.AddCommand(NewAddHandlerCommand().Definition())
	cmd.AddCommand(NewAddAPICommand().Definition())
	cmd.AddCommand(NewAddServiceCommand().Definition())

	return cmd
}

// Commands returns subcommands in this group
func (g *AddGroup) Commands() []core.Command {
	return []core.Command{
		NewAddAppCommand(),
		NewAddModelCommand(),
		NewAddHandlerCommand(),
		NewAddAPICommand(),
		NewAddServiceCommand(),
	}
}

// Execute runs the command logic (shows help if no subcommand)
func (g *AddGroup) Execute(ctx *core.Context, args []string) error {
	return ctx.Cmd.Help()
}
