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
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("clean", false, "Mark dirty migration as clean after manual remediation")
	cmd.Flags().Uint("version", 0, "Specific migration version to mark clean (defaults to current dirty version)")
	cmd.Flags().Bool("verify", false, "Validate migration files integrity and compute checksums")
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

	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	rec := execute.NewRecovery(database.DB)
	cmdCtx := context.Background()

	// 1. Verify file checksums if requested
	if verifyFlag {
		fmt.Printf("Verifying migration file integrity in %s...\n", migrationsPath)
		checksums, err := rec.ValidateMigrationIntegrity(migrationsPath)
		if err != nil {
			return fmt.Errorf("integrity check failed: %w", err)
		}
		fmt.Printf("✓ Verified %d migration files. All checksums valid.\n", len(checksums))
	}

	// 2. Check for dirty migration state
	dirtyMigration, err := rec.RecoverDirtyState(cmdCtx, migrationsPath)
	if err != nil {
		// If table doesn't exist or no migrations yet, inform user cleanly
		fmt.Printf("Database check: %v\n", err)
		return nil
	}

	if dirtyMigration == nil {
		fmt.Println("✓ Database migration state is clean (no dirty migrations found).")
		return nil
	}

	fmt.Printf("⚠️  Dirty migration detected at version %d!\n", dirtyMigration.Version)
	fmt.Printf("   Error: %s\n\n", dirtyMigration.ErrorMsg)
	fmt.Println("Recommended recovery steps:")
	steps := rec.GetRecoverySteps(fmt.Sprintf("%d", dirtyMigration.Version), dirtyMigration.ErrorMsg)
	for _, step := range steps {
		fmt.Printf("   %s\n", step)
	}
	fmt.Println()

	if cleanFlag {
		targetVersion := dirtyMigration.Version
		if versionFlag > 0 {
			targetVersion = versionFlag
		}

		if err := rec.MarkMigrationClean(cmdCtx, targetVersion); err != nil {
			return fmt.Errorf("failed to mark migration as clean: %w", err)
		}
		fmt.Printf("✓ Migration %d has been marked as clean.\n", targetVersion)
	} else {
		fmt.Println("To mark this migration as clean after fixing the database manually, run:")
		fmt.Printf("   forge migrate recover --clean --version %d\n", dirtyMigration.Version)
	}

	return nil
}
