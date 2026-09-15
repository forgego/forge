package migrations

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/spf13/cobra"
)

// StatusCommand creates the migration status command
type StatusCommand struct{}

// NewStatusCommand creates a new instance of StatusCommand
func NewStatusCommand() *StatusCommand {
	return &StatusCommand{}
}

// Definition returns the cobra command definition
func (c *StatusCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  "Display the current migration status, including applied and pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := core.NewContext()
			ctx.Cmd = cmd
			return c.Execute(ctx, args)
		},
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	return cmd
}

// Execute runs the command logic
func (c *StatusCommand) Execute(ctx *core.Context, args []string) error {
	// Get migrations path
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	out := ctx.Cmd.OutOrStdout()

	// List migration files first (works without DB)
	pattern := filepath.Join(migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Fprintf(out, "[WARN] Could not read migration files from %q: %v\n\n", migrationsPath, err)
	} else {
		renderMigrationFiles(out, matches)
	}

	// Try to connect to database for detailed status
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		fmt.Fprintln(out, "[WARN] Could not connect to database - showing file listing only")
		fmt.Fprintln(out, "       To see database status, ensure database is configured and running")
		return nil
	}
	defer database.Close()

	// Checksum baseline verification. A verification error (e.g. dirty DB) is
	// non-fatal here: the normal status is still printed below, then the
	// verification error is printed and a non-nil error is returned.
	rec := execute.NewRecovery(database.DB)
	baselineReport, verifyCheckErr := rec.VerifyAgainstBaseline(ctx.Cmd.Context(), migrationsPath)
	var baselineErr error
	if verifyCheckErr != nil {
		baselineErr = fmt.Errorf("failed to verify checksum baseline: %w", verifyCheckErr)
	} else {
		renderChecksumBaseline(out, baselineReport)

		if !baselineReport.AllVerified() {
			baselineErr = fmt.Errorf("checksum baseline verification failed")
		}
	}

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		fmt.Fprintln(out, "[WARN] Could not create migration runner - showing file listing only")
		if verifyCheckErr != nil {
			fmt.Fprintf(out, "Verification error: %v\n", verifyCheckErr)
		}
		return baselineErr
	}
	defer runner.Close()

	// Get status
	cmdCtx := context.Background()
	status, err := runner.Status(cmdCtx)
	if err != nil {
		fmt.Fprintln(out, "[WARN] Could not get database status - showing file listing only")
		if verifyCheckErr != nil {
			fmt.Fprintf(out, "Verification error: %v\n", verifyCheckErr)
		}
		return baselineErr
	}
	if status == nil {
		fmt.Fprintln(out, "[WARN] Database returned empty migration status - showing file listing only")
		if verifyCheckErr != nil {
			fmt.Fprintf(out, "Verification error: %v\n", verifyCheckErr)
		}
		return baselineErr
	}

	// Try to get detailed status (if available)
	detailedStatus, err := runner.GetDetailedStatus(cmdCtx)
	if err != nil {
		detailedStatus = nil
	}

	renderMigrationStatus(out, status, detailedStatus)

	if verifyCheckErr != nil {
		fmt.Fprintf(out, "Verification error: %v\n", verifyCheckErr)
	}

	return baselineErr
}

func renderChecksumBaseline(out io.Writer, report *execute.BaselineReport) {
	if report == nil {
		return
	}
	counts := report.Counts()
	fmt.Fprintln(out, "\nChecksum baseline:")
	fmt.Fprintf(out, "  Counts: %d verified, %d mismatched, %d missing, %d unverified\n",
		counts[execute.BaselineStatusVerified],
		counts[execute.BaselineStatusMismatched],
		counts[execute.BaselineStatusMissing],
		counts[execute.BaselineStatusUnverified],
	)
	for _, entry := range report.Entries {
		if entry.Status != execute.BaselineStatusVerified {
			if entry.Detail != "" {
				fmt.Fprintf(out, "  %d  %s  %s\n", entry.Version, entry.Status, entry.Detail)
			} else {
				fmt.Fprintf(out, "  %d  %s\n", entry.Version, entry.Status)
			}
		}
	}
}

func renderMigrationFiles(out io.Writer, matches []string) {
	if len(matches) == 0 {
		fmt.Fprintln(out, "Migration Files (0): none found")
		fmt.Fprintln(out)
		return
	}

	files := make([]string, len(matches))
	for i, match := range matches {
		files[i] = filepath.Base(match)
	}
	sort.Strings(files)

	fmt.Fprintf(out, "Migration Files (%d):\n", len(files))
	for _, file := range files {
		fmt.Fprintf(out, "  - %s\n", file)
	}
	fmt.Fprintln(out)
}

func renderMigrationStatus(out io.Writer, status *db.MigrationStatus, detailedStatus *db.DetailedMigrationStatus) {
	if status == nil {
		fmt.Fprintln(out, "Database Migration Status:")
		fmt.Fprintln(out, "  [WARN] Migration status unavailable")
		if detailedStatus == nil {
			fmt.Fprintln(out, "\n  (Detailed status not available)")
		}
		return
	}

	fmt.Fprintln(out, "Database Migration Status:")
	fmt.Fprintf(out, "  Current Version: %d\n", status.Version)
	if status.Dirty {
		fmt.Fprintln(out, "  Status: DIRTY (migration failed, manual intervention required)")
		fmt.Fprintln(out, "\n  [WARN] Database is in a dirty state!")
		fmt.Fprintln(out, "         Manual intervention required before running migrations.")
	} else if status.Version == 0 {
		fmt.Fprintln(out, "  Status: No migrations applied")
	} else {
		fmt.Fprintln(out, "  Status: OK")
	}

	if detailedStatus == nil {
		fmt.Fprintln(out, "\n  (Detailed status not available)")
		return
	}

	if len(detailedStatus.Applied) > 0 {
		fmt.Fprintf(out, "\n  Applied Migrations (%d):\n", len(detailedStatus.Applied))
		for _, mig := range detailedStatus.Applied {
			fmt.Fprintf(out, "    [x] %s\n", mig)
		}
	}

	if len(detailedStatus.Pending) > 0 {
		fmt.Fprintf(out, "\n  Pending Migrations (%d):\n", len(detailedStatus.Pending))
		for _, mig := range detailedStatus.Pending {
			fmt.Fprintf(out, "    [ ] %s\n", mig)
		}
	}

	if len(detailedStatus.OutOfOrder) > 0 {
		fmt.Fprintf(out, "\n  [WARN] Out-of-Order Migrations (%d):\n", len(detailedStatus.OutOfOrder))
		for _, mig := range detailedStatus.OutOfOrder {
			fmt.Fprintf(out, "    [!] %s (applied before current version)\n", mig)
		}
	}

	if detailedStatus.Next != "" && detailedStatus.Next != "Already at latest version" {
		fmt.Fprintf(out, "\n  Next Migration: [%s]\n", detailedStatus.Next)
	} else if detailedStatus.Next == "Already at latest version" {
		fmt.Fprintln(out, "\n  Next Migration: Already at latest version")
	}
}
