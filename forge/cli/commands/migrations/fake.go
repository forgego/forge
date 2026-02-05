package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/forgego/forge/cli/core"
	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/migrate/verify"
	"github.com/spf13/cobra"
)

// FakeCommand creates the fake command for marking migrations as applied
type FakeCommand struct{}

// NewFakeCommand creates a new instance of FakeCommand
func NewFakeCommand() *FakeCommand {
	return &FakeCommand{}
}

// Definition returns the cobra command definition
func (c *FakeCommand) Definition() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fake [version]",
		Short: "Mark migrations as applied without running them",
		Long:  "Mark migrations as applied without actually executing them. Use --fake-initial to mark initial migrations as applied if tables already exist.",
		Args:  cobra.MaximumNArgs(1),
	}
	cmd.Flags().String("path", "./migrations", "Path to migrations directory")
	cmd.Flags().Bool("fake-initial", false, "Mark initial migrations as applied if tables already exist")
	return cmd
}

// Execute runs the command logic
func (c *FakeCommand) Execute(ctx *core.Context, args []string) error {
	migrationsPath, err := ctx.Cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("failed to get migrations path flag: %w", err)
	}
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	fakeInitial, err := ctx.Cmd.Flags().GetBool("fake-initial")
	if err != nil {
		return fmt.Errorf("failed to get fake-initial flag: %w", err)
	}

	database, err := db.NewDBFromConfig(ctx.Config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	runner, err := db.NewMigrationRunner(database, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migration runner: %w", err)
	}
	defer runner.Close()

	cmdCtx := context.Background()

	if fakeInitial {
		// Mark initial migrations as applied if tables already exist
		return c.fakeInitialMigrations(cmdCtx, database, migrationsPath, runner)
	}

	if len(args) == 0 {
		// Mark all pending migrations as applied
		return c.fakeAllPending(cmdCtx, database, migrationsPath, runner)
	}

	// Mark specific version as applied
	versionStr := args[0]
	version, err := strconv.ParseUint(versionStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version number: %w", err)
	}

	// Validate migration exists
	if err := c.validateMigrationExists(migrationsPath, versionStr); err != nil {
		return err
	}

	// Use Force to mark as applied (similar to Django's --fake)
	if err := runner.Force(cmdCtx, uint(version)); err != nil {
		return fmt.Errorf("failed to fake migration version: %w", err)
	}

	fmt.Printf("✓ Marked migration version %s as applied (without running it)\n", versionStr)
	return nil
}

// fakeInitialMigrations marks initial migrations as applied if tables already exist
func (c *FakeCommand) fakeInitialMigrations(ctx context.Context, db *db.DB, migrationsPath string, runner *db.MigrationRunner) error {
	fmt.Println("Checking for existing tables to determine initial migrations...")

	// Get existing tables from database
	existingTables, err := c.getExistingTables(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get existing tables: %w", err)
	}

	if len(existingTables) == 0 {
		fmt.Println("No existing tables found - nothing to fake")
		return nil
	}

	fmt.Printf("Found %d existing table(s): %s\n", len(existingTables), strings.Join(existingTables, ", "))

	// Find migrations that create these tables
	migrationsToFake, err := c.findMigrationsForTables(migrationsPath, existingTables)
	if err != nil {
		return fmt.Errorf("failed to find migrations for tables: %w", err)
	}

	if len(migrationsToFake) == 0 {
		fmt.Println("No migrations found that create these tables")
		return nil
	}

	// Sort by version
	sort.Slice(migrationsToFake, func(i, j int) bool {
		vi, _ := strconv.ParseUint(migrationsToFake[i].Version, 10, 64)
		vj, _ := strconv.ParseUint(migrationsToFake[j].Version, 10, 64)
		return vi < vj
	})

	// Mark each migration as applied
	checksumValidator := verify.NewChecksumValidator(migrationsPath)
	for _, mig := range migrationsToFake {
		version, _ := strconv.ParseUint(mig.Version, 10, 64)
		
		// Calculate checksum
		upPath := filepath.Join(migrationsPath, fmt.Sprintf("%s_%s.up.sql", mig.Version, mig.Name))
		checksum, err := checksumValidator.CalculateChecksum(upPath)
		if err != nil {
			checksum = "" // Continue without checksum if calculation fails
		}

		// Mark as applied using Force
		if err := runner.Force(ctx, uint(version)); err != nil {
			return fmt.Errorf("failed to fake migration %s: %w", mig.Version, err)
		}

		// Optionally store checksum in schema_migrations (if custom table structure)
		_ = checksum // Checksum stored by golang-migrate internally

		fmt.Printf("  ✓ Marked %s_%s as applied\n", mig.Version, mig.Name)
	}

	fmt.Printf("\n✓ Marked %d initial migration(s) as applied\n", len(migrationsToFake))
	return nil
}

// fakeAllPending marks all pending migrations as applied
func (c *FakeCommand) fakeAllPending(ctx context.Context, db *db.DB, migrationsPath string, runner *db.MigrationRunner) error {
	// Get current version
	currentVersion, _, err := runner.Version(ctx)
	if err != nil {
		// If there's an error getting version, assume no migrations applied yet
		// Version() should return 0, false, nil for no migrations, but handle errors gracefully
		currentVersion = 0
	}

	// Get all migration files
	pattern := filepath.Join(migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Parse and filter pending migrations
	var pendingMigrations []migrationInfo
	checksumValidator := verify.NewChecksumValidator(migrationsPath)

	for _, match := range matches {
		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) < 2 {
			continue
		}

		versionStr := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")
		version, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			continue
		}

		// Check if this migration is pending
		if uint(version) > currentVersion {
			// Calculate checksum
			checksum, err := checksumValidator.CalculateChecksum(match)
			if err != nil {
				checksum = ""
			}

			pendingMigrations = append(pendingMigrations, migrationInfo{
				Version:  versionStr,
				Name:     name,
				VersionU: uint(version),
				Checksum: checksum,
			})
		}
	}

	if len(pendingMigrations) == 0 {
		fmt.Println("No pending migrations to fake")
		return nil
	}

	// Sort by version
	sort.Slice(pendingMigrations, func(i, j int) bool {
		return pendingMigrations[i].VersionU < pendingMigrations[j].VersionU
	})

	// Confirm if more than 5 migrations
	if len(pendingMigrations) > 5 {
		fmt.Printf("Warning: About to mark %d migrations as applied without running them.\n", len(pendingMigrations))
		fmt.Print("Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Canceled")
			return nil
		}
	}

	// Mark each pending migration as applied
	for _, mig := range pendingMigrations {
		if err := runner.Force(ctx, mig.VersionU); err != nil {
			return fmt.Errorf("failed to fake migration %s: %w", mig.Version, err)
		}
		fmt.Printf("  ✓ Marked %s_%s as applied\n", mig.Version, mig.Name)
	}

	fmt.Printf("\n✓ Marked %d pending migration(s) as applied\n", len(pendingMigrations))
	return nil
}

// getExistingTables gets list of existing tables from database
func (c *FakeCommand) getExistingTables(ctx context.Context, db *db.DB) ([]string, error) {
	var tables []string
	var query string

	// Use driver-specific queries
	if db.Driver == "postgres" || db.Driver == "postgresql" {
		query = `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' AND table_name != 'schema_migrations' ORDER BY table_name`
	} else if db.Driver == "sqlite" || db.Driver == "sqlite3" {
		query = `SELECT name FROM sqlite_master WHERE type='table' AND name != 'sqlite_sequence' AND name != 'schema_migrations' ORDER BY name`
	} else {
		return nil, fmt.Errorf("unsupported database driver: %s", db.Driver)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, rows.Err()
}

// migrationInfo represents a migration file
type migrationInfo struct {
	Version  string
	Name     string
	VersionU uint
	Checksum string
}

// findMigrationsForTables finds migrations that create the specified tables
func (c *FakeCommand) findMigrationsForTables(migrationsPath string, tables []string) ([]migrationInfo, error) {
	tableSet := make(map[string]bool)
	for _, table := range tables {
		tableSet[strings.ToLower(table)] = true
	}

	// Find all migration files
	pattern := filepath.Join(migrationsPath, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var migrations []migrationInfo
	for _, match := range matches {
		content, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		basename := filepath.Base(match)
		parts := strings.Split(basename, "_")
		if len(parts) < 2 {
			continue
		}

		versionStr := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		// Check if this migration creates any of the tables
		sqlContent := strings.ToUpper(string(content))
		for table := range tableSet {
			// Look for CREATE TABLE statements
			createPattern := fmt.Sprintf(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?%s["']?`, regexp.QuoteMeta(strings.ToUpper(table)))
			matched, _ := regexp.MatchString(createPattern, sqlContent)
			if matched {
				version, _ := strconv.ParseUint(versionStr, 10, 64)
				migrations = append(migrations, migrationInfo{
					Version:  versionStr,
					Name:     name,
					VersionU: uint(version),
				})
				break // Found a table, move to next migration
			}
		}
	}

	return migrations, nil
}

// validateMigrationExists validates that a migration file exists
func (c *FakeCommand) validateMigrationExists(migrationsPath string, versionStr string) error {
	pattern := filepath.Join(migrationsPath, versionStr+"_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Errorf("migration version %s does not exist", versionStr)
	}

	return nil
}
