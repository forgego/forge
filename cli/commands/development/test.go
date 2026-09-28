package development

import (
	"fmt"
	"os/exec"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// TestCommand creates the test command
type TestCommand struct{}

// NewTestCommand creates a new instance of TestCommand
func NewTestCommand() *TestCommand {
	return &TestCommand{}
}

// Definition returns the cobra command definition
func (c *TestCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		Long:  "Run all tests in the project using go test",
	}
	cmd.Flags().Bool("verbose", false, "Verbose output")
	cmd.Flags().Bool("coverage", false, "Show coverage")
	return cmd
}

// Execute runs the command logic
func (c *TestCommand) Execute(ctx *core.Context, args []string) error {
	testDir := "."
	if len(args) > 0 {
		testDir = args[0]
	}

	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")
	coverage, _ := ctx.Cmd.Flags().GetBool("coverage")

	fmt.Printf("Running tests in %s...\n", testDir)

	cmd := exec.CommandContext(ctx.Cmd.Context(), "go", buildGoTestArgs(verbose, coverage)...)
	cmd.Dir = testDir
	cmd.Stdout = ctx.Cmd.OutOrStdout()
	cmd.Stderr = ctx.Cmd.ErrOrStderr()
	return cmd.Run()
}

func buildGoTestArgs(verbose, coverage bool) []string {
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if coverage {
		args = append(args, "-cover")
	}
	return append(args, "./...")
}
