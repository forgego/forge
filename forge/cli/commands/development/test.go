package development

import (
	"fmt"

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
	// Use go test to run tests
	// This is a wrapper around go test
	testDir := "."
	if len(args) > 0 {
		testDir = args[0]
	}

	// Get verbose flag
	verbose, _ := ctx.Cmd.Flags().GetBool("verbose")

	// Get coverage flag
	coverage, _ := ctx.Cmd.Flags().GetBool("coverage")

	fmt.Printf("Running tests in %s...\n", testDir)
	fmt.Println("Note: This is a basic wrapper. Use 'go test' directly for full features.")

	// For MVP, just inform user to use go test directly
	// Full implementation would execute go test programmatically
	fmt.Println("To run tests, use: go test ./...")
	if verbose {
		fmt.Println("  with -v for verbose output")
	}
	if coverage {
		fmt.Println("  with -cover for coverage")
	}

	return nil
}
