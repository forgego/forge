package migrations

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/execute"
	"github.com/spf13/cobra"
)

// UpCommand creates the migrate up command
type UpCommand struct{}

// NewUpCommand creates a new instance of UpCommand
func NewUpCommand() *UpCommand {
	return &UpCommand{}
}

// Definition returns the cobra command definition
func (c *UpCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up",
		Short:   "Apply migrations",
		Long:    "Apply pending migrations to the database",
		Aliases: []string{"apply"},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := core.NewContext()
			ctx.Cmd = cmd
			return c.Execute(ctx, args)
		},
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("dry-run", false, "Preview migrations without applying them")
	return cmd
}

// Execute runs the command logic
func (c *UpCommand) Execute(ctx *core.Context, args []string) error {
	// Get migrations path
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	out := ctx.Cmd.OutOrStdout()

	// Check for dry-run flag first (doesn't need DB connection)
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")

	if dryRun {
		// Dry-run mode: show what would be executed
		fmt.Fprintln(out, "Dry-run mode: Preview of migrations that would be applied:")
		fmt.Fprintln(out)

		// Find all migration files
		pattern := filepath.Join(migrationsPath, "*_*.up.sql")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("failed to scan migrations directory: %w", err)
		}

		// Try to get current version, but don't fail if DB is not available
		currentVersion := uint(0)
		// Try to connect to get current version (optional)
		database, err := db.NewDBFromConfig(ctx.Config)
		if err == nil {
			defer database.Close()
			cmdCtx := context.Background()
			rec := execute.NewRecovery(database.DB)
			report, err := rec.VerifyAgainstBaseline(cmdCtx, migrationsPath)
			if err == nil && report != nil {
				if report.HasBlocking() {
					var blocking []string
					for _, e := range report.Entries {
						if e.Status == execute.BaselineStatusMismatched || e.Status == execute.BaselineStatusMissing {
							blocking = append(blocking, fmt.Sprintf("%d (%s)", e.Version, e.Status))
						}
					}
					return fmt.Errorf("migration baseline verification failed for version(s): %s; run 'forge migrations recover --verify' to inspect", strings.Join(blocking, ", "))
				}
				if report.Counts()[execute.BaselineStatusUnverified] > 0 {
					fmt.Fprintln(out, "Hint: unverified migrations exist; run 'forge migrations baseline --adopt' to record baseline checksums")
				}
			}
			runner, err := db.NewMigrationRunner(database, migrationsPath)
			if err == nil {
				defer runner.Close()
				ver, _, err := runner.Version(cmdCtx)
				if err == nil {
					currentVersion = ver
				}
			}
		}

		// Sort and filter migrations
		var pendingMigrations []string
		for _, match := range matches {
			basename := filepath.Base(match)
			versionStr := strings.Split(basename, "_")[0]
			version, err := strconv.ParseUint(versionStr, 10, 64)
			if err == nil && version <= math.MaxUint && uint(version) > currentVersion {
				pendingMigrations = append(pendingMigrations, match)
			}
		}

		// Sort by version
		sort.Strings(pendingMigrations)

		if len(pendingMigrations) == 0 {
			if currentVersion == 0 {
				fmt.Fprintln(out, "  All migration files would be applied (database not connected to check current version)")
			} else {
				fmt.Fprintln(out, "  No pending migrations")
			}
			return nil
		}

		fmt.Fprintf(out, "  Would apply %d migration(s):\n\n", len(pendingMigrations))
		for i, migFile := range pendingMigrations {
			content, err := os.ReadFile(migFile)
			if err != nil {
				continue
			}

			basename := filepath.Base(migFile)
			fmt.Fprintf(out, "  Migration %d: %s\n", i+1, basename)
			fmt.Fprintln(out, "  SQL Preview (first 20 lines):")
			lines := strings.Split(string(content), "\n")
			previewLines := lines
			if len(lines) > 20 {
				previewLines = lines[:20]
			}
			for _, line := range previewLines {
				if strings.TrimSpace(line) != "" {
					fmt.Fprintf(out, "    %s\n", line)
				}
			}
			if len(lines) > 20 {
				fmt.Fprintf(out, "    ... (%d more lines)\n", len(lines)-20)
			}
			fmt.Fprintln(out)
		}

		fmt.Fprintln(out, "(No changes were made to the database)")
		return nil
	}

	// Connect to database (only needed for actual migration)
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	cmdCtx := context.Background()

	// Before applying, verify against baseline
	rec := execute.NewRecovery(database.DB)
	report, err := rec.VerifyAgainstBaseline(cmdCtx, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to verify migration baseline: %w", err)
	}
	if report != nil {
		if report.HasBlocking() {
			var blocking []string
			for _, e := range report.Entries {
				if e.Status == execute.BaselineStatusMismatched || e.Status == execute.BaselineStatusMissing {
					blocking = append(blocking, fmt.Sprintf("%d (%s)", e.Version, e.Status))
				}
			}
			return fmt.Errorf("migration baseline verification failed for version(s): %s; run 'forge migrations recover --verify' to inspect", strings.Join(blocking, ", "))
		}
		if report.Counts()[execute.BaselineStatusUnverified] > 0 {
			fmt.Fprintln(out, "Hint: unverified migrations exist; run 'forge migrations baseline --adopt' to record baseline checksums")
		}
	}

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}
	defer runner.Close()

	// Apply migrations
	if err := runner.Migrate(cmdCtx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Fprintln(out, "✓ Migrations applied successfully")
	return nil
}
