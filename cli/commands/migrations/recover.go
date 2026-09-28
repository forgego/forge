package migrations

import (
	"context"
	"fmt"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/spf13/cobra"
)

// RecoverCommand creates the recover command
type RecoverCommand struct{}

// NewRecoverCommand creates a new instance of RecoverCommand
func NewRecoverCommand() *RecoverCommand {
	return &RecoverCommand{}
}

// Definition returns the cobra command definition
func (c *RecoverCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover",
		Short: "Recover from dirty migrations and verify checksum integrity",
		Long:  "Detects failed/dirty migrations, prints actionable recovery steps, allows marking clean, and verifies migration file checksums",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := core.NewContext()
			ctx.Cmd = cmd
			return c.Execute(ctx, args)
		},
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("clean", false, "Mark dirty migration as clean after manual remediation")
	cmd.Flags().Uint("version", 0, "Specific migration version to mark clean (defaults to current dirty version)")
	cmd.Flags().Bool("verify", false, "Compare applied migration files against recorded checksum baselines")
	return cmd
}

// Execute runs the command logic
func (c *RecoverCommand) Execute(ctx *core.Context, args []string) error {
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return err
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	cleanFlag, _ := ctx.Cmd.Flags().GetBool("clean")
	versionFlag, _ := ctx.Cmd.Flags().GetUint("version")
	verifyFlag, _ := ctx.Cmd.Flags().GetBool("verify")
	out := ctx.Cmd.OutOrStdout()

	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	rec := execute.NewRecovery(database.DB)
	cmdCtx := context.Background()

	runVerify := func() error {
		fmt.Fprintf(out, "Verifying migration file integrity against baseline in %s...\n", migrationsPath)
		report, err := rec.VerifyAgainstBaseline(cmdCtx, migrationsPath)
		if err != nil {
			fmt.Fprintf(out, "Verification error: %v\n", err)
			return fmt.Errorf("integrity check failed: %w", err)
		}

		for _, entry := range report.Entries {
			if entry.Status != execute.BaselineStatusVerified {
				if entry.Detail != "" {
					fmt.Fprintf(out, "  %d  %s  %s\n", entry.Version, entry.Status, entry.Detail)
				} else {
					fmt.Fprintf(out, "  %d  %s\n", entry.Version, entry.Status)
				}
			}
		}

		counts := report.Counts()
		fmt.Fprintf(out, "Summary: %d verified, %d mismatched, %d missing, %d unverified\n",
			counts[execute.BaselineStatusVerified],
			counts[execute.BaselineStatusMismatched],
			counts[execute.BaselineStatusMissing],
			counts[execute.BaselineStatusUnverified],
		)

		if !report.AllVerified() {
			return fmt.Errorf("checksum baseline verification failed")
		}
		return nil
	}

	var verifyErr error

	// 1. Verify against baseline if requested
	if verifyFlag {
		verifyErr = runVerify()
	}

	// 2. Check for dirty migration state
	dirtyMigration, err := rec.RecoverDirtyState(cmdCtx, migrationsPath)
	if err != nil {
		// If table doesn't exist or no migrations yet, inform user cleanly
		fmt.Fprintf(out, "Database check: %v\n", err)
		if verifyErr != nil {
			return verifyErr
		}
		return nil
	}

	if dirtyMigration == nil {
		fmt.Fprintln(out, "✓ Database migration state is clean (no dirty migrations found).")
		if verifyErr != nil {
			return verifyErr
		}
		return nil
	}

	fmt.Fprintf(out, "⚠️  Dirty migration detected at version %d!\n", dirtyMigration.Version)
	fmt.Fprintf(out, "   Error: %s\n\n", dirtyMigration.ErrorMsg)
	fmt.Fprintln(out, "Recommended recovery steps:")
	steps := rec.GetRecoverySteps(fmt.Sprintf("%d", dirtyMigration.Version), dirtyMigration.ErrorMsg)
	for _, step := range steps {
		fmt.Fprintf(out, "   %s\n", step)
	}
	fmt.Fprintln(out)

	if cleanFlag {
		targetVersion := dirtyMigration.Version
		if versionFlag > 0 {
			targetVersion = versionFlag
		}

		if err := rec.MarkMigrationClean(cmdCtx, targetVersion); err != nil {
			return fmt.Errorf("failed to mark migration as clean: %w", err)
		}
		fmt.Fprintf(out, "✓ Migration %d has been marked as clean.\n", targetVersion)
		if verifyFlag {
			// Re-run verification after cleaning; only the post-clean result matters.
			verifyErr = runVerify()
		}
	} else {
		fmt.Fprintln(out, "To mark this migration as clean after fixing the database manually, run:")
		fmt.Fprintf(out, "   forge migrate recover --clean --version %d\n", dirtyMigration.Version)
	}

	if verifyErr != nil {
		return verifyErr
	}
	return nil
}
