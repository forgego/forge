package core

import (
	"fmt"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// VersionCommand creates the version command
type VersionCommand struct{}

// NewVersionCommand creates a new instance of VersionCommand
func NewVersionCommand() *VersionCommand {
	return &VersionCommand{}
}

// Definition returns the cobra command definition
func (c *VersionCommand) Definition() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Forge",
		Long:  `All software has versions. This is Forge's`,
	}
}

// Execute runs the command logic
func (c *VersionCommand) Execute(ctx *core.Context, args []string) error {
	fmt.Println("Forge CLI v0.1.0")
	fmt.Println("Forge Framework v0.1.0")
	return nil
}
