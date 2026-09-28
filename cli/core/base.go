package core

import (
	"github.com/spf13/cobra"
)

// Command represents a CLI command
type Command interface {
	// Definition returns the cobra command definition
	Definition() *cobra.Command
	// Execute runs the command logic
	Execute(ctx *Context, args []string) error
}

// CommandGroup represents a group of related commands
type CommandGroup interface {
	Command
	// Commands returns subcommands in this group
	Commands() []Command
}
