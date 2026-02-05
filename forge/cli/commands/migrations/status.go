package migrations

import (
	"context"
	"fmt"
	"path/filepath"

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
	if err == nil && len(matches) > 0 {
		fmt.Printf("Migration Files (%d):\n", len(matches))
		for _, match := range matches {
			fmt.Printf("  - %s\n", filepath.Base(match))
		}
		fmt.Println()
	}

	// Try to connect to database for detailed status
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		fmt.Println("⚠️  Could not connect to database - showing file listing only")
		fmt.Println("   To see database status, ensure database is configured and running")
		return nil
	}
	defer database.Close()

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		fmt.Println("⚠️  Could not create migration runner - showing file listing only")
		return nil
	}
	defer runner.Close()

	// Get status
	cmdCtx := context.Background()
	status, err := runner.Status(cmdCtx)
	if err != nil {
		fmt.Println("⚠️  Could not get database status - showing file listing only")
		return nil
	}

	// Display status
	fmt.Println("Database Migration Status:")
	fmt.Printf("  Current Version: %d\n", status.Version)
	if status.Dirty {
		fmt.Println("  Status: DIRTY (migration failed, manual intervention required)")
		fmt.Println("\n  ⚠️  WARNING: Database is in a dirty state!")
		fmt.Println("     Manual intervention required before running migrations.")
	} else if status.Version == 0 {
		fmt.Println("  Status: No migrations applied")
	} else {
		fmt.Println("  Status: OK")
	}

	// Try to get detailed status (if available)
	detailedStatus, err := runner.GetDetailedStatus(cmdCtx)
	if err == nil && detailedStatus != nil {
		// Show applied migrations
		if len(detailedStatus.Applied) > 0 {
			fmt.Printf("\n  Applied Migrations (%d):\n", len(detailedStatus.Applied))
			for _, mig := range detailedStatus.Applied {
				fmt.Printf("    ✓ %s\n", mig)
			}
		}

		// Show pending migrations
		if len(detailedStatus.Pending) > 0 {
			fmt.Printf("\n  Pending Migrations (%d):\n", len(detailedStatus.Pending))
			for _, mig := range detailedStatus.Pending {
				fmt.Printf("    ○ %s\n", mig)
			}
		}

		// Show out-of-order migrations
		if len(detailedStatus.OutOfOrder) > 0 {
			fmt.Printf("\n  ⚠️  Out-of-Order Migrations (%d):\n", len(detailedStatus.OutOfOrder))
			for _, mig := range detailedStatus.OutOfOrder {
				fmt.Printf("    ⚠  %s (applied before current version)\n", mig)
			}
		}

		if detailedStatus.Next != "" && detailedStatus.Next != "Already at latest version" {
			fmt.Printf("\n  Next Migration: [%s]\n", detailedStatus.Next)
		} else if detailedStatus.Next == "Already at latest version" {
			fmt.Printf("\n  Next Migration: Already at latest version\n")
		}
	} else {
		// Fallback: show basic file listing
		fmt.Println("\n  (Detailed status not available)")
	}

	return nil
}
