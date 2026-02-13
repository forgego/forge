package migrations

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
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

	// List migration files first (works without DB)
	pattern := filepath.Join(migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("[WARN] Could not read migration files from %q: %v\n\n", migrationsPath, err)
	} else {
		renderMigrationFiles(os.Stdout, matches)
	}

	// Try to connect to database for detailed status
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		fmt.Println("[WARN] Could not connect to database - showing file listing only")
		fmt.Println("       To see database status, ensure database is configured and running")
		return nil
	}
	defer database.Close()

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		fmt.Println("[WARN] Could not create migration runner - showing file listing only")
		return nil
	}
	defer runner.Close()

	// Get status
	cmdCtx := context.Background()
	status, err := runner.Status(cmdCtx)
	if err != nil {
		fmt.Println("[WARN] Could not get database status - showing file listing only")
		return nil
	}
	if status == nil {
		fmt.Println("[WARN] Database returned empty migration status - showing file listing only")
		return nil
	}

	// Try to get detailed status (if available)
	detailedStatus, err := runner.GetDetailedStatus(cmdCtx)
	if err != nil {
		detailedStatus = nil
	}

	renderMigrationStatus(os.Stdout, status, detailedStatus)
	return nil
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
