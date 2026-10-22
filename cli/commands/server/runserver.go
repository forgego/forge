package server

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

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
	port, err := ctx.Cmd.Flags().GetString("port")
	if err != nil {
		return fmt.Errorf("failed to get port flag: %w", err)
	}
	if port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return fmt.Errorf("invalid port %q: expected 1-65535", port)
		}
		cmd.Env = append(os.Environ(), "FORGE_SERVER_PORT="+port)
	}

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
