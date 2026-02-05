package sql

import (
	"github.com/forgego/forge/db/migrate/core"
)

// SQLBuilder generates SQL DDL statements from changes
type SQLBuilder interface {
	// BuildUpSQL generates the up migration SQL for a list of changes
	BuildUpSQL(changes []core.Change) (string, error)

	// BuildDownSQL generates the down migration SQL for a list of changes
	BuildDownSQL(changes []core.Change) (string, error)

	// BuildCreateTable generates CREATE TABLE SQL
	BuildCreateTable(change *core.CreateTable) (string, error)

	// BuildAddColumn generates ALTER TABLE ADD COLUMN SQL
	BuildAddColumn(change *core.AddColumn) (string, error)

	// BuildModifyColumn generates ALTER TABLE ALTER COLUMN SQL
	BuildModifyColumn(change *core.ModifyColumn) (string, error)

	// BuildAddIndex generates CREATE INDEX SQL
	BuildAddIndex(change *core.AddIndex) (string, error)

	// BuildModifyIndex generates DROP and CREATE INDEX SQL
	BuildModifyIndex(change *core.ModifyIndex) (string, error)

	// BuildAddForeignKey generates ALTER TABLE ADD CONSTRAINT FOREIGN KEY SQL
	BuildAddForeignKey(change *core.AddForeignKey) (string, error)

	// BuildModifyForeignKey generates DROP and ADD FOREIGN KEY SQL
	BuildModifyForeignKey(change *core.ModifyForeignKey) (string, error)

	// BuildAddConstraint generates ALTER TABLE ADD CONSTRAINT SQL
	BuildAddConstraint(change *core.AddConstraint) (string, error)
}

// NewSQLBuilder creates a new SQL builder for the given driver
func NewSQLBuilder(driver core.Driver) (SQLBuilder, error) {
	switch {
	case driver.IsPostgreSQL():
		return NewPostgreSQLBuilder(), nil
	case driver.IsSQLite():
		return NewSQLiteBuilder(), nil
	default:
		return nil, core.NewMigrationError(
			core.ErrInvalidChange,
			"unsupported database driver: "+driver.String(),
			nil,
		)
	}
}
