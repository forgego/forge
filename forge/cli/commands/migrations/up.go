package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
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

	// Check for dry-run flag first (doesn't need DB connection)
	dryRun, _ := ctx.Cmd.Flags().GetBool("dry-run")

	if dryRun {
		// Dry-run mode: show what would be executed
		fmt.Println("Dry-run mode: Preview of migrations that would be applied:")
		fmt.Println()

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
			runner, err := db.NewMigrationRunner(database, migrationsPath)
			if err == nil {
				defer runner.Close()
				cmdCtx := context.Background()
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
			if err == nil && uint(version) > currentVersion {
				pendingMigrations = append(pendingMigrations, match)
			}
		}

		// Sort by version
		sort.Strings(pendingMigrations)

		if len(pendingMigrations) == 0 {
			if currentVersion == 0 {
				fmt.Println("  All migration files would be applied (database not connected to check current version)")
			} else {
				fmt.Println("  No pending migrations")
			}
			return nil
		}

		fmt.Printf("  Would apply %d migration(s):\n\n", len(pendingMigrations))
		for i, migFile := range pendingMigrations {
			content, err := os.ReadFile(migFile)
			if err != nil {
				continue
			}

			basename := filepath.Base(migFile)
			fmt.Printf("  Migration %d: %s\n", i+1, basename)
			fmt.Println("  SQL Preview (first 20 lines):")
			lines := strings.Split(string(content), "\n")
			previewLines := lines
			if len(lines) > 20 {
				previewLines = lines[:20]
			}
			for _, line := range previewLines {
				if strings.TrimSpace(line) != "" {
					fmt.Printf("    %s\n", line)
				}
			}
			if len(lines) > 20 {
				fmt.Printf("    ... (%d more lines)\n", len(lines)-20)
			}
			fmt.Println()
		}

		fmt.Println("(No changes were made to the database)")
		return nil
	}

	// Connect to database (only needed for actual migration)
	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Create migration runner
	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}
	defer runner.Close()

	// Apply migrations
	cmdCtx := context.Background()
	if err := runner.Migrate(cmdCtx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("✓ Migrations applied successfully")
	return nil
}
