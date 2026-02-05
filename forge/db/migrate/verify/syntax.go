package verify

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// Validator validates migrations before execution
type Validator struct{}

// NewValidator creates a new migration validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateMigration validates a migration before execution
func (v *Validator) ValidateMigration(plan *core.MigrationPlan) error {
	if plan == nil {
		return fmt.Errorf("migration plan is nil")
	}

	if plan.Version == "" {
		return fmt.Errorf("migration version is empty")
	}

	if plan.Name == "" {
		return fmt.Errorf("migration name is empty")
	}

	if len(plan.Changes) == 0 {
		return fmt.Errorf("migration has no changes")
	}

	// Validate SQL is not empty
	if strings.TrimSpace(plan.UpSQL) == "" {
		return fmt.Errorf("migration up SQL is empty")
	}

	// Validate changes
	for i, change := range plan.Changes {
		if err := v.validateChange(change, i); err != nil {
			return fmt.Errorf("change %d: %w", i, err)
		}
	}

	// Validate dependencies
	if err := v.validateDependencies(plan.Dependencies); err != nil {
		return fmt.Errorf("dependency validation failed: %w", err)
	}

	return nil
}

// validateChange validates a single change
func (v *Validator) validateChange(change core.Change, index int) error {
	if change == nil {
		return fmt.Errorf("change is nil")
	}

	// Data migrations (RunSQL, RunGo) don't have table names
	if change.TableName() == "" && change.Type() != core.ChangeTypeRunSQL && change.Type() != core.ChangeTypeRunGo {
		return fmt.Errorf("change has empty table name")
	}

	// Type-specific validations
	switch c := change.(type) {
	case *core.CreateTable:
		if c.Table == nil {
			return fmt.Errorf("CreateTable has nil table definition")
		}
		if c.Table.Name == "" {
			return fmt.Errorf("CreateTable has empty table name")
		}
		if len(c.Table.Fields) == 0 {
			return fmt.Errorf("CreateTable has no fields")
		}

	case *core.AddColumn:
		if c.Column.Name == "" {
			return fmt.Errorf("AddColumn has empty column name")
		}
		if c.Column.Type == "" {
			return fmt.Errorf("AddColumn has empty column type")
		}

	case *core.DropColumn:
		if c.ColumnName == "" {
			return fmt.Errorf("DropColumn has empty column name")
		}

	case *core.AddIndex:
		if c.Index.Name == "" {
			return fmt.Errorf("AddIndex has empty index name")
		}
		if len(c.Index.Fields) == 0 {
			return fmt.Errorf("AddIndex has no fields")
		}

	case *core.AddForeignKey:
		if c.Relation.Name == "" {
			return fmt.Errorf("AddForeignKey has empty relation name")
		}
		if c.TargetTable == "" {
			return fmt.Errorf("AddForeignKey has empty target table")
		}
		if c.Relation.To == "" {
			return fmt.Errorf("AddForeignKey has empty target model")
		}

	case *core.RunSQL:
		if c.SQL == "" {
			return fmt.Errorf("RunSQL has empty SQL")
		}
		// Data migrations don't need table names
		// Override the table name check for data migrations
		return nil

	case *core.RunGo:
		if c.UpFunc == "" {
			return fmt.Errorf("RunGo has empty UpFunc")
		}
		// Data migrations don't need table names
		// Override the table name check for data migrations
		return nil
	}

	return nil
}

// ValidateSQL validates SQL syntax (basic checks)
func (v *Validator) ValidateSQL(sql string) error {
	if strings.TrimSpace(sql) == "" {
		return fmt.Errorf("SQL is empty")
	}

	// Basic syntax checks
	// Note: Full SQL parsing would require a SQL parser library
	// This is a basic validation that checks for common issues

	// Check for balanced parentheses
	openParens := strings.Count(sql, "(")
	closeParens := strings.Count(sql, ")")
	if openParens != closeParens {
		return fmt.Errorf("unbalanced parentheses in SQL")
	}

	// Check for balanced quotes (basic check)
	singleQuotes := strings.Count(sql, "'")
	if singleQuotes%2 != 0 {
		return fmt.Errorf("unbalanced single quotes in SQL")
	}

	return nil
}

// ValidateReversible checks if all changes in a migration are reversible
func (v *Validator) ValidateReversible(changes []core.Change) error {
	irreversible := []string{}

	for _, change := range changes {
		if !change.Reversible() {
			irreversible = append(irreversible, string(change.Type()))
		}
	}

	if len(irreversible) > 0 {
		return fmt.Errorf("migration contains irreversible changes: %v", irreversible)
	}

	return nil
}

// validateDependencies validates migration dependencies
func (v *Validator) validateDependencies(dependencies []core.Dependency) error {
	if len(dependencies) == 0 {
		return nil // No dependencies is valid
	}

	// Check for circular dependencies
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	for _, dep := range dependencies {
		if err := v.checkCircularDependency(dep, visited, visiting); err != nil {
			return err
		}
	}

	// Validate dependency format
	for _, dep := range dependencies {
		if dep.Version == "" {
			return fmt.Errorf("dependency has empty version")
		}
		// Version should be a valid migration version format (e.g., "000001")
		if len(dep.Version) != 6 {
			return fmt.Errorf("dependency version %s has invalid format (expected 6 digits)", dep.Version)
		}
	}

	return nil
}

// checkCircularDependency checks for circular dependencies (simplified check)
func (v *Validator) checkCircularDependency(dep core.Dependency, visited, visiting map[string]bool) error {
	key := dep.Version
	if dep.App != "" {
		key = dep.App + ":" + dep.Version
	}

	if visiting[key] {
		return fmt.Errorf("circular dependency detected involving %s", key)
	}

	if visited[key] {
		return nil
	}

	visiting[key] = true
	defer func() {
		visiting[key] = false
		visited[key] = true
	}()

	// In a full implementation, we would recursively check dependencies
	// For now, this is a basic check
	return nil
}
