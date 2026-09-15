package migrations

import (
	"fmt"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/spf13/cobra"
)

// BaselineCommand reports or adopts migration checksum baselines.
type BaselineCommand struct{}

// NewBaselineCommand creates a baseline command.
func NewBaselineCommand() *BaselineCommand { return &BaselineCommand{} }

// Definition returns the command and its flags.
func (c *BaselineCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{Use: "baseline", Short: "Verify or adopt migration checksum baselines", RunE: func(cmd *cobra.Command, args []string) error {
		ctx := core.NewContext()
		ctx.Cmd = cmd
		return c.Execute(ctx, args)
	}}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("adopt", false, "Record current file checksums for applied migrations without a baseline")
	return cmd
}

// Execute verifies the baseline, optionally adopting previously unverified versions.
func (c *BaselineCommand) Execute(ctx *core.Context, args []string) error {
	path, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return err
	}
	if path == "" {
		path = "./migrations"
	}
	adopt, err := ctx.Cmd.Flags().GetBool("adopt")
	if err != nil {
		return err
	}
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()
	cmdCtx := ctx.Cmd.Context()
	if adopt {
		runner, err := db.NewMigrationRunner(database, path)
		if err != nil {
			return fmt.Errorf("failed to create migration runner: %w", err)
		}
		defer runner.Close()
		versions, err := runner.AdoptChecksumBaseline(cmdCtx)
		if err != nil {
			return err
		}
		fmt.Fprintf(ctx.Cmd.OutOrStdout(), "Adopted versions: %v\n", versions)
	}
	report, err := execute.NewRecovery(database.DB).VerifyAgainstBaseline(cmdCtx, path)
	if err != nil {
		return fmt.Errorf("failed to verify checksum baseline: %w", err)
	}
	renderChecksumBaseline(ctx.Cmd.OutOrStdout(), report)
	if !report.AllVerified() {
		return fmt.Errorf("checksum baseline verification failed")
	}
	return nil
}
