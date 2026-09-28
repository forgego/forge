package development

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/spf13/cobra"
)

// CheckCommand creates the check command
type CheckCommand struct{}

// NewCheckCommand creates a new instance of CheckCommand
func NewCheckCommand() *CheckCommand {
	return &CheckCommand{}
}

// Definition returns the cobra command definition
func (c *CheckCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run code quality checks",
		Long:  "Run static analysis and code quality checks (go vet, staticcheck, golangci-lint)",
	}
	cmd.Flags().Bool("vet", true, "Run go vet")
	cmd.Flags().Bool("staticcheck", false, "Run staticcheck (requires staticcheck to be installed)")
	cmd.Flags().Bool("golangci-lint", false, "Run golangci-lint (requires golangci-lint to be installed)")
	cmd.Flags().Bool("all", false, "Run all available checks")
	cmd.Flags().String("path", ".", "Path to check (default: current directory)")
	return cmd
}

// Execute runs the command logic
func (c *CheckCommand) Execute(ctx *core.Context, args []string) error {
	// Get flags
	runVet, _ := ctx.Cmd.Flags().GetBool("vet")
	runStaticcheck, _ := ctx.Cmd.Flags().GetBool("staticcheck")
	runGolangciLint, _ := ctx.Cmd.Flags().GetBool("golangci-lint")
	runAll, _ := ctx.Cmd.Flags().GetBool("all")
	checkPath, _ := ctx.Cmd.Flags().GetString("path")

	if runAll {
		runVet = true
		runStaticcheck = true
		runGolangciLint = true
	}

	// Resolve path
	absPath, err := filepath.Abs(checkPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	var errors []string
	var warnings []string

	// Run go vet
	if runVet {
		fmt.Println("Running go vet...")
		if err := runGoVet(absPath); err != nil {
			errors = append(errors, fmt.Sprintf("go vet failed: %v", err))
		} else {
			fmt.Println("✓ go vet passed")
		}
	}

	// Run staticcheck if requested
	if runStaticcheck {
		fmt.Println("\nRunning staticcheck...")
		if err := runStaticcheckTool(absPath); err != nil {
			if strings.Contains(err.Error(), "executable file not found") {
				warnings = append(warnings, "staticcheck not found (install with: go install honnef.co/go/tools/cmd/staticcheck@latest)")
			} else {
				errors = append(errors, fmt.Sprintf("staticcheck failed: %v", err))
			}
		} else {
			fmt.Println("✓ staticcheck passed")
		}
	}

	// Run golangci-lint if requested
	if runGolangciLint {
		fmt.Println("\nRunning golangci-lint...")
		if err := runGolangciLintTool(absPath); err != nil {
			if strings.Contains(err.Error(), "executable file not found") {
				warnings = append(warnings, "golangci-lint not found (install from: https://golangci-lint.run/)")
			} else {
				errors = append(errors, fmt.Sprintf("golangci-lint failed: %v", err))
			}
		} else {
			fmt.Println("✓ golangci-lint passed")
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
	}

	// Print errors and exit
	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, errMsg := range errors {
			fmt.Printf("  ✗ %s\n", errMsg)
		}
		return fmt.Errorf("code quality checks failed")
	}

	if !runVet && !runStaticcheck && !runGolangciLint {
		fmt.Println("No checks selected. Use --all or specify --vet, --staticcheck, or --golangci-lint")
		return nil
	}

	fmt.Println("\n✓ All checks passed!")
	return nil
}

// runGoVet runs go vet on the specified path
func runGoVet(path string) error {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runStaticcheckTool runs staticcheck on the specified path
func runStaticcheckTool(path string) error {
	cmd := exec.Command("staticcheck", "./...")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGolangciLintTool runs golangci-lint on the specified path
func runGolangciLintTool(path string) error {
	cmd := exec.Command("golangci-lint", "run")
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
