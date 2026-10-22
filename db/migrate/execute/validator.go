package execute

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
	"github.com/forgego/forge/db/migrate/parse"
	"github.com/forgego/forge/db/migrate/verify"
)

// MigrationValidator validates migrations before execution
type MigrationValidator struct {
	validator *verify.Validator
}

type allowDestructiveKey struct{}

// WithDestructiveOperationsAllowed marks an explicitly confirmed migration run.
func WithDestructiveOperationsAllowed(ctx context.Context) context.Context {
	return context.WithValue(ctx, allowDestructiveKey{}, true)
}

// NewMigrationValidator creates a new migration validator
func NewMigrationValidator() *MigrationValidator {
	return &MigrationValidator{
		validator: verify.NewValidator(),
	}
}

// ValidateBeforeExecution validates a migration file before execution
func (v *MigrationValidator) ValidateBeforeExecution(ctx context.Context, migrationPath string) error {
	// Read migration file
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Validate SQL syntax
	if err := v.validator.ValidateSQL(string(content)); err != nil {
		return fmt.Errorf("SQL validation failed: %w", err)
	}

	// Check for dangerous operations
	if ctx.Value(allowDestructiveKey{}) != true {
		operations, err := parse.DestructiveOperations(string(content))
		if err != nil {
			return fmt.Errorf("failed to inspect operations: %w", err)
		}
		if len(operations) > 0 {
			return fmt.Errorf("dangerous operation detected: %s", strings.Join(operations, ", "))
		}
	}

	return nil
}

// ValidateMigrationPlan validates a migration plan before generation
func (v *MigrationValidator) ValidateMigrationPlan(plan *core.MigrationPlan) error {
	return v.validator.ValidateMigration(plan)
}

// ValidateMigrationPair validates that up and down migrations are consistent
func (v *MigrationValidator) ValidateMigrationPair(upPath, downPath string) error {
	// Check both files exist
	if _, err := os.Stat(upPath); os.IsNotExist(err) {
		return fmt.Errorf("up migration file does not exist: %s", upPath)
	}

	if _, err := os.Stat(downPath); os.IsNotExist(err) {
		return fmt.Errorf("down migration file does not exist: %s", downPath)
	}

	// Read both files
	upContent, err := os.ReadFile(upPath)
	if err != nil {
		return fmt.Errorf("failed to read up migration: %w", err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		return fmt.Errorf("failed to read down migration: %w", err)
	}

	// Basic validation: down should not be empty if up is not empty
	if len(strings.TrimSpace(string(upContent))) > 0 && len(strings.TrimSpace(string(downContent))) == 0 {
		return fmt.Errorf("down migration is empty but up migration is not")
	}

	return nil
}

// ValidateMigrationsDir validates all migrations in a directory
func (v *MigrationValidator) ValidateMigrationsDir(migrationsDir string) error {
	pattern := filepath.Join(migrationsDir, "*_*.up.sql")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	var errors []error
	for _, upPath := range matches {
		downPath := strings.Replace(upPath, ".up.sql", ".down.sql", 1)

		if err := v.ValidateMigrationPair(upPath, downPath); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", filepath.Base(upPath), err))
		}

		if err := v.ValidateBeforeExecution(context.Background(), upPath); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", filepath.Base(upPath), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed for %d migration(s): %v", len(errors), errors)
	}

	return nil
}
