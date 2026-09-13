package server

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// RunServerCommand creates the runserver command
type RunServerCommand struct{}

// NewRunServerCommand creates a new instance of RunServerCommand
func NewRunServerCommand() *RunServerCommand {
	return &RunServerCommand{}
}

// Definition returns the cobra command definition
func (c *RunServerCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runserver",
		Short: "Run development server",
		Long:  "Start the development web server",
	}
	cmd.Flags().String("port", "", "Port to run server on (overrides config)")
	return cmd
}

// Execute runs the command logic
func (c *RunServerCommand) Execute(ctx *core.Context, args []string) error {
	// Check if cmd/server/main.go exists
	if _, err := os.Stat("cmd/server/main.go"); os.IsNotExist(err) {
		return fmt.Errorf("cmd/server/main.go not found. Are you in a Forge project root?")
	}

	fmt.Println("Starting development server...")

	// Prepare command: go run cmd/server/main.go
	// In the future, we can add file watching here (using air or similar)
	cmd := exec.Command("go", "run", "cmd/server/main.go")

	// Pass through arguments if needed
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args...)
	}

	// Connect pipes
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("server crashed: %w", err)
	}

	return nil
}
