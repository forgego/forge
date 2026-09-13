package main

import (
	"fmt"
	"os"

	"github.com/forgego/forge/cli"
)

func main() {
	// Build root command with all registered commands
	rootCmd := cli.BuildRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
