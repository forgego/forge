// Package migrate provides a public API for migration generation, state management,
// and SQL generation. This is the main entry point for the migration system.
//
// Core types are re-exported from the core sub-package for convenience.
package migrate

import (
	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// ChangeDetector is the interface for detecting changes
type ChangeDetector interface {
	DetectChanges(current, previous []*generator.ModelDefinition) ([]Change, error)
}

// Re-export types for convenience
type (
	Change     = core.Change
	ChangeType = core.ChangeType
	Driver     = core.Driver
)

// Re-export change types
type (
	CreateTable      = core.CreateTable
	DropTable        = core.DropTable
	RenameTable      = core.RenameTable
	AddColumn        = core.AddColumn
	DropColumn       = core.DropColumn
	ModifyColumn     = core.ModifyColumn
	RenameColumn     = core.RenameColumn
	AddIndex         = core.AddIndex
	DropIndex        = core.DropIndex
	ModifyIndex      = core.ModifyIndex
	AddForeignKey    = core.AddForeignKey
	DropForeignKey   = core.DropForeignKey
	ModifyForeignKey = core.ModifyForeignKey
	AddConstraint    = core.AddConstraint
	DropConstraint   = core.DropConstraint
	RunSQL           = core.RunSQL
	RunGo            = core.RunGo
	UnknownChange    = core.UnknownChange
)

// Re-export constants
const (
	ChangeTypeCreateTable      = core.ChangeTypeCreateTable
	ChangeTypeDropTable        = core.ChangeTypeDropTable
	ChangeTypeRenameTable      = core.ChangeTypeRenameTable
	ChangeTypeAddColumn        = core.ChangeTypeAddColumn
	ChangeTypeDropColumn       = core.ChangeTypeDropColumn
	ChangeTypeModifyColumn     = core.ChangeTypeModifyColumn
	ChangeTypeRenameColumn     = core.ChangeTypeRenameColumn
	ChangeTypeAddIndex         = core.ChangeTypeAddIndex
	ChangeTypeDropIndex        = core.ChangeTypeDropIndex
	ChangeTypeModifyIndex      = core.ChangeTypeModifyIndex
	ChangeTypeAddForeignKey    = core.ChangeTypeAddForeignKey
	ChangeTypeDropForeignKey   = core.ChangeTypeDropForeignKey
	ChangeTypeModifyForeignKey = core.ChangeTypeModifyForeignKey
	ChangeTypeAddConstraint    = core.ChangeTypeAddConstraint
	ChangeTypeDropConstraint   = core.ChangeTypeDropConstraint
	ChangeTypeRunSQL           = core.ChangeTypeRunSQL
	ChangeTypeRunGo            = core.ChangeTypeRunGo
	ChangeTypeUnknown          = core.ChangeTypeUnknown
	DriverPostgreSQL           = core.DriverPostgreSQL
	DriverSQLite               = core.DriverSQLite
	DriverSQLite3              = core.DriverSQLite3
)

// Re-export migration types
type (
	Migration     = core.Migration
	MigrationPlan = core.MigrationPlan
	Dependency    = core.Dependency
)

// NewMigrationPlan creates a new migration plan
func NewMigrationPlan(version, name string, changes []Change) *MigrationPlan {
	return core.NewMigrationPlan(version, name, changes)
}
