package development

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// ShellCommand creates the shell command
type ShellCommand struct{}

// NewShellCommand creates a new instance of ShellCommand
func NewShellCommand() *ShellCommand {
	return &ShellCommand{}
}

// Definition returns the cobra command definition
func (c *ShellCommand) Definition() *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Interactive shell",
		Long:  "Start an interactive shell with framework context",
	}
}

// Execute runs the command logic
func (c *ShellCommand) Execute(ctx *core.Context, args []string) error {
	fmt.Println("forge Interactive Shell")
	fmt.Println("Type 'help' for available commands, 'exit' or 'quit' to exit")
	fmt.Println()

	shell := NewInteractiveShell(ctx)
	return shell.Run(context.Background())
}

// InteractiveShell represents an interactive shell session
type InteractiveShell struct {
	ctx      *core.Context
	commands map[string]ShellCommandFunc
	history  []string
}

// ShellCommandFunc is a function that handles a shell command
type ShellCommandFunc func(args []string) error

// NewInteractiveShell creates a new interactive shell
func NewInteractiveShell(ctx *core.Context) *InteractiveShell {
	shell := &InteractiveShell{
		ctx:      ctx,
		commands: make(map[string]ShellCommandFunc),
		history:  make([]string, 0),
	}

	// Register built-in commands
	shell.RegisterCommand("help", shell.cmdHelp)
	shell.RegisterCommand("history", shell.cmdHistory)
	shell.RegisterCommand("clear", shell.cmdClear)
	shell.RegisterCommand("models", shell.cmdModels)
	shell.RegisterCommand("dbinfo", shell.cmdDBInfo)

	return shell
}

// RegisterCommand registers a shell command
func (s *InteractiveShell) RegisterCommand(name string, fn ShellCommandFunc) {
	s.commands[name] = fn
}

// Run runs the interactive shell
func (s *InteractiveShell) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("forge> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Add to history
		s.history = append(s.history, line)

		// Check for exit commands
		if line == "exit" || line == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Parse command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		args := parts[1:]

		// Execute command
		if cmd, exists := s.commands[cmdName]; exists {
			if err := cmd(args); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		} else {
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmdName)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

// cmdHelp shows help for available commands
func (s *InteractiveShell) cmdHelp(args []string) error {
	fmt.Println("Available commands:")
	fmt.Println("  help      - Show this help message")
	fmt.Println("  history   - Show command history")
	fmt.Println("  clear     - Clear the screen")
	fmt.Println("  models    - List registered models")
	fmt.Println("  dbinfo    - Show database information")
	fmt.Println("  exit/quit - Exit the shell")
	return nil
}

// cmdHistory shows command history
func (s *InteractiveShell) cmdHistory(args []string) error {
	if len(s.history) == 0 {
		fmt.Println("No command history")
		return nil
	}

	fmt.Println("Command history:")
	for i, cmd := range s.history {
		fmt.Printf("%4d  %s\n", i+1, cmd)
	}
	return nil
}

// cmdClear clears the screen
func (s *InteractiveShell) cmdClear(args []string) error {
	fmt.Print("\033[H\033[2J")
	return nil
}

// cmdModels lists registered models
func (s *InteractiveShell) cmdModels(args []string) error {
	fmt.Println("Registered models:")
	fmt.Println("  (Model registry integration pending)")
	return nil
}

// cmdDBInfo shows database information
func (s *InteractiveShell) cmdDBInfo(args []string) error {
	fmt.Println("Database information:")
	fmt.Println("  (Database connection info integration pending)")
	return nil
}
