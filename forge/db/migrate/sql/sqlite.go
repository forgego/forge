package sql

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// SQLiteBuilder implements SQLBuilder for SQLite
type SQLiteBuilder struct {
	*baseBuilder
}

// NewSQLiteBuilder creates a new SQLite SQL builder
func NewSQLiteBuilder() SQLBuilder {
	return &SQLiteBuilder{
		baseBuilder: &baseBuilder{
			isSQLite:   true,
			isPostgres: false,
		},
	}
}

// BuildUpSQL generates the up migration SQL for a list of changes
func (b *SQLiteBuilder) BuildUpSQL(changes []core.Change) (string, error) {
	var statements []string

	// Order changes by dependency
	orderedChanges := orderChanges(changes)

	for _, change := range orderedChanges {
		sql, err := b.buildChangeUpSQL(change)
		if err != nil {
			return "", core.NewMigrationError(
				core.ErrInvalidChange,
				fmt.Sprintf("failed to generate SQL for %s", change.Type()),
				err,
			)
		}
		if sql != "" {
			statements = append(statements, sql)
		}
	}

	return strings.Join(statements, "\n\n"), nil
}

// BuildDownSQL generates the down migration SQL for a list of changes
func (b *SQLiteBuilder) BuildDownSQL(changes []core.Change) (string, error) {
	var statements []string

	// Reverse order for down migrations
	orderedChanges := orderChanges(changes)
	for i := len(orderedChanges) - 1; i >= 0; i-- {
		change := orderedChanges[i]
		sql, err := b.buildChangeDownSQL(change)
		if err != nil {
			return "", core.NewMigrationError(
				core.ErrInvalidChange,
				fmt.Sprintf("failed to generate down SQL for %s", change.Type()),
				err,
			)
		}
		if sql != "" {
			statements = append(statements, sql)
		}
	}

	return strings.Join(statements, "\n\n"), nil
}

// buildChangeUpSQL generates up SQL for a single change
func (b *SQLiteBuilder) buildChangeUpSQL(change core.Change) (string, error) {
	switch c := change.(type) {
	case *core.CreateTable:
		return b.BuildCreateTable(c)
	case *core.DropTable:
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;", c.Table), nil
	case *core.RenameTable:
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", c.OldName, c.NewName), nil
	case *core.AddColumn:
		return b.BuildAddColumn(c)
	case *core.DropColumn:
		// SQLite doesn't support DROP COLUMN directly
		return fmt.Sprintf("-- SQLite does not support DROP COLUMN directly\n-- Column %s.%s should be dropped manually", c.Table, c.ColumnName), nil
	case *core.ModifyColumn:
		return b.BuildModifyColumn(c)
	case *core.RenameColumn:
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s;", c.Table, c.OldName, c.NewName), nil
	case *core.AddIndex:
		return b.BuildAddIndex(c)
	case *core.DropIndex:
		return fmt.Sprintf("DROP INDEX IF EXISTS %s;", c.IndexName), nil
	case *core.ModifyIndex:
		return b.BuildModifyIndex(c)
	case *core.AddForeignKey:
		return b.BuildAddForeignKey(c)
	case *core.DropForeignKey:
		return fmt.Sprintf("-- SQLite does not support DROP FOREIGN KEY directly\n-- Foreign key %s should be dropped manually", c.FKName), nil
	case *core.ModifyForeignKey:
		return b.BuildModifyForeignKey(c)
	case *core.AddConstraint:
		return b.BuildAddConstraint(c)
	case *core.DropConstraint:
		return fmt.Sprintf("-- SQLite has limited constraint support\n-- Constraint %s should be dropped manually", c.ConstraintName), nil
	case *core.RunSQL:
		// For RunSQL, return the SQL directly
		return c.SQL, nil
	case *core.RunGo:
		// For RunGo, embed as a comment with special marker for execution
		// The actual execution will be handled by the runner
		return fmt.Sprintf("-- RUNGO: %s\n-- This migration requires Go code execution", c.UpFunc), nil
	default:
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			fmt.Sprintf("unknown change type: %T", change),
			nil,
		)
	}
}

// buildChangeDownSQL generates down SQL for a single change
func (b *SQLiteBuilder) buildChangeDownSQL(change core.Change) (string, error) {
	switch c := change.(type) {
	case *core.CreateTable:
		return fmt.Sprintf("DROP TABLE IF EXISTS %s;", c.TableName()), nil
	case *core.DropTable:
		// Cannot generate down SQL for DropTable without original table definition
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"cannot generate down SQL for DropTable without table definition",
			nil,
		)
	case *core.RenameTable:
		return fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", c.NewName, c.OldName), nil
	case *core.AddColumn:
		return fmt.Sprintf("-- SQLite does not support DROP COLUMN directly\n-- Column %s.%s should be dropped manually", c.Table, c.Column.Name), nil
	case *core.DropColumn:
		// Cannot generate down SQL for DropColumn without original column definition
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"cannot generate down SQL for DropColumn without column definition",
			nil,
		)
	case *core.ModifyColumn:
		// Reverse the modification
		return b.buildModifyColumnDown(c)
	case *core.RenameColumn:
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s;", c.Table, c.NewName, c.OldName), nil
	case *core.AddIndex:
		indexName := c.Index.Name
		if indexName == "" {
			indexName = fmt.Sprintf("idx_%s_%s", c.Table, strings.Join(c.Index.Fields, "_"))
		}
		return fmt.Sprintf("DROP INDEX IF EXISTS %s;", indexName), nil
	case *core.DropIndex:
		// Cannot generate down SQL for DropIndex without original index definition
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"cannot generate down SQL for DropIndex without index definition",
			nil,
		)
	case *core.ModifyIndex:
		// Reverse the modification
		reversed := &core.ModifyIndex{
			Table:    c.Table,
			OldIndex: c.NewIndex,
			NewIndex: c.OldIndex,
		}
		return b.BuildModifyIndex(reversed)
	case *core.AddForeignKey:
		fkName := fmt.Sprintf("fk_%s_%s", c.Table, c.Relation.Name)
		return fmt.Sprintf("-- SQLite does not support DROP FOREIGN KEY directly\n-- Foreign key %s should be dropped manually", fkName), nil
	case *core.DropForeignKey:
		// Cannot generate down SQL for DropForeignKey without original FK definition
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"cannot generate down SQL for DropForeignKey without FK definition",
			nil,
		)
	case *core.ModifyForeignKey:
		// Reverse the modification
		reversed := &core.ModifyForeignKey{
			Table:       c.Table,
			OldFK:       c.NewFK,
			NewFK:       c.OldFK,
			TargetTable: c.TargetTable,
		}
		return b.BuildModifyForeignKey(reversed)
	case *core.AddConstraint:
		return fmt.Sprintf("-- SQLite has limited constraint support\n-- Constraint %s should be dropped manually", c.Constraint.Name), nil
	case *core.DropConstraint:
		// Cannot generate down SQL for DropConstraint without original constraint definition
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"cannot generate down SQL for DropConstraint without constraint definition",
			nil,
		)
	case *core.RunSQL:
		// For RunSQL, return the reverse SQL if provided
		if c.ReverseSQL != "" {
			return c.ReverseSQL, nil
		}
		if !c.CanReverse {
			return "-- This data migration is not reversible", nil
		}
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"RunSQL migration is marked as reversible but no ReverseSQL provided",
			nil,
		)
	case *core.RunGo:
		// For RunGo, embed as a comment with special marker for execution
		if c.DownFunc != "" {
			return fmt.Sprintf("-- RUNGO: %s\n-- This migration requires Go code execution for rollback", c.DownFunc), nil
		}
		if !c.CanReverse {
			return "-- This data migration is not reversible", nil
		}
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			"RunGo migration is marked as reversible but no DownFunc provided",
			nil,
		)
	default:
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			fmt.Sprintf("unknown change type: %T", change),
			nil,
		)
	}
}

// BuildModifyColumn generates ALTER TABLE ALTER COLUMN statement for SQLite
// SQLite has very limited ALTER TABLE support, so this generates a comment
func (b *SQLiteBuilder) BuildModifyColumn(c *core.ModifyColumn) (string, error) {
	newType := mapFieldTypeToSQL(c.NewColumn, true, false)
	oldType := mapFieldTypeToSQL(c.OldColumn, true, false)

	// SQLite has limited ALTER TABLE support
	// This would require table recreation in practice
	return fmt.Sprintf("-- SQLite does not support ALTER COLUMN directly\n-- Column %s.%s type changed from %s to %s\n-- This requires table recreation in practice",
		c.Table, c.NewColumn.Name, oldType, newType), nil
}

// buildModifyColumnDown generates the reverse of ModifyColumn
func (b *SQLiteBuilder) buildModifyColumnDown(c *core.ModifyColumn) (string, error) {
	reversed := &core.ModifyColumn{
		Table:     c.Table,
		OldColumn: c.NewColumn,
		NewColumn: c.OldColumn,
	}
	return b.BuildModifyColumn(reversed)
}

// BuildCreateTable delegates to baseBuilder
func (b *SQLiteBuilder) BuildCreateTable(c *core.CreateTable) (string, error) {
	return b.baseBuilder.BuildCreateTable(c)
}

// BuildAddColumn delegates to baseBuilder
func (b *SQLiteBuilder) BuildAddColumn(c *core.AddColumn) (string, error) {
	return b.baseBuilder.BuildAddColumn(c)
}

// BuildAddIndex delegates to baseBuilder
func (b *SQLiteBuilder) BuildAddIndex(c *core.AddIndex) (string, error) {
	return b.baseBuilder.BuildAddIndex(c)
}

// BuildModifyIndex delegates to baseBuilder
func (b *SQLiteBuilder) BuildModifyIndex(c *core.ModifyIndex) (string, error) {
	return b.baseBuilder.BuildModifyIndex(c)
}

// BuildAddForeignKey delegates to baseBuilder
func (b *SQLiteBuilder) BuildAddForeignKey(c *core.AddForeignKey) (string, error) {
	return b.baseBuilder.BuildAddForeignKey(c)
}

// BuildModifyForeignKey delegates to baseBuilder
func (b *SQLiteBuilder) BuildModifyForeignKey(c *core.ModifyForeignKey) (string, error) {
	return b.baseBuilder.BuildModifyForeignKey(c)
}

// BuildAddConstraint delegates to baseBuilder
func (b *SQLiteBuilder) BuildAddConstraint(c *core.AddConstraint) (string, error) {
	return b.baseBuilder.BuildAddConstraint(c)
}
