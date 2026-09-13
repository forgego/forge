package core

import (
	"fmt"

	"github.com/forgego/forge/codegen"
)

// ChangeType represents the type of a migration change
type ChangeType string

const (
	ChangeTypeCreateTable      ChangeType = "CreateTable"
	ChangeTypeDropTable        ChangeType = "DropTable"
	ChangeTypeRenameTable      ChangeType = "RenameTable"
	ChangeTypeAddColumn        ChangeType = "AddColumn"
	ChangeTypeDropColumn       ChangeType = "DropColumn"
	ChangeTypeModifyColumn     ChangeType = "ModifyColumn"
	ChangeTypeRenameColumn     ChangeType = "RenameColumn"
	ChangeTypeAddIndex         ChangeType = "AddIndex"
	ChangeTypeDropIndex        ChangeType = "DropIndex"
	ChangeTypeModifyIndex      ChangeType = "ModifyIndex"
	ChangeTypeAddForeignKey    ChangeType = "AddForeignKey"
	ChangeTypeDropForeignKey   ChangeType = "DropForeignKey"
	ChangeTypeModifyForeignKey ChangeType = "ModifyForeignKey"
	ChangeTypeAddConstraint    ChangeType = "AddConstraint"
	ChangeTypeDropConstraint   ChangeType = "DropConstraint"
	ChangeTypeRunSQL           ChangeType = "RunSQL"
	ChangeTypeRunGo            ChangeType = "RunGo"
	ChangeTypeUnknown          ChangeType = "Unknown"
)

// Change represents a single migration change
type Change interface {
	Type() ChangeType
	TableName() string
	Reversible() bool
}

// CreateTable represents creating a new table
type CreateTable struct {
	Table *generator.ModelDefinition
}

func (c *CreateTable) Type() ChangeType { return ChangeTypeCreateTable }
func (c *CreateTable) TableName() string {
	if c.Table.Meta.TableName != "" {
		return c.Table.Meta.TableName
	}
	return fmt.Sprintf("%ss", toSnakeCase(c.Table.Name))
}
func (c *CreateTable) Reversible() bool { return true }

// DropTable represents dropping a table
type DropTable struct {
	Table string
}

func (c *DropTable) Type() ChangeType  { return ChangeTypeDropTable }
func (c *DropTable) TableName() string { return c.Table }
func (c *DropTable) Reversible() bool  { return true }

// RenameTable represents renaming a table
type RenameTable struct {
	OldName string
	NewName string
}

func (c *RenameTable) Type() ChangeType  { return ChangeTypeRenameTable }
func (c *RenameTable) TableName() string { return c.NewName }
func (c *RenameTable) Reversible() bool  { return true }

// AddColumn represents adding a new column
type AddColumn struct {
	Table  string
	Column generator.FieldDefinition
}

func (c *AddColumn) Type() ChangeType  { return ChangeTypeAddColumn }
func (c *AddColumn) TableName() string { return c.Table }
func (c *AddColumn) Reversible() bool  { return true }

// DropColumn represents dropping a column
type DropColumn struct {
	Table      string
	ColumnName string
}

func (c *DropColumn) Type() ChangeType  { return ChangeTypeDropColumn }
func (c *DropColumn) TableName() string { return c.Table }
func (c *DropColumn) Reversible() bool  { return true }

// ModifyColumn represents modifying a column
type ModifyColumn struct {
	Table     string
	OldColumn generator.FieldDefinition
	NewColumn generator.FieldDefinition
}

func (c *ModifyColumn) Type() ChangeType  { return ChangeTypeModifyColumn }
func (c *ModifyColumn) TableName() string { return c.Table }
func (c *ModifyColumn) Reversible() bool  { return true }

// RenameColumn represents renaming a column
type RenameColumn struct {
	Table   string
	OldName string
	NewName string
}

func (c *RenameColumn) Type() ChangeType  { return ChangeTypeRenameColumn }
func (c *RenameColumn) TableName() string { return c.Table }
func (c *RenameColumn) Reversible() bool  { return true }

// AddIndex represents adding an index
type AddIndex struct {
	Table string
	Index generator.IndexDefinition
}

func (c *AddIndex) Type() ChangeType  { return ChangeTypeAddIndex }
func (c *AddIndex) TableName() string { return c.Table }
func (c *AddIndex) Reversible() bool  { return true }

// DropIndex represents dropping an index
type DropIndex struct {
	Table     string
	IndexName string
}

func (c *DropIndex) Type() ChangeType  { return ChangeTypeDropIndex }
func (c *DropIndex) TableName() string { return c.Table }
func (c *DropIndex) Reversible() bool  { return true }

// ModifyIndex represents modifying an index
type ModifyIndex struct {
	Table    string
	OldIndex generator.IndexDefinition
	NewIndex generator.IndexDefinition
}

func (c *ModifyIndex) Type() ChangeType  { return ChangeTypeModifyIndex }
func (c *ModifyIndex) TableName() string { return c.Table }
func (c *ModifyIndex) Reversible() bool  { return true }

// AddForeignKey represents adding a foreign key
type AddForeignKey struct {
	Table       string
	Relation    generator.RelationDefinition
	TargetTable string
}

func (c *AddForeignKey) Type() ChangeType  { return ChangeTypeAddForeignKey }
func (c *AddForeignKey) TableName() string { return c.Table }
func (c *AddForeignKey) Reversible() bool  { return true }

// DropForeignKey represents dropping a foreign key
type DropForeignKey struct {
	Table  string
	FKName string
}

func (c *DropForeignKey) Type() ChangeType  { return ChangeTypeDropForeignKey }
func (c *DropForeignKey) TableName() string { return c.Table }
func (c *DropForeignKey) Reversible() bool  { return true }

// ModifyForeignKey represents modifying a foreign key
type ModifyForeignKey struct {
	Table       string
	OldFK       generator.RelationDefinition
	NewFK       generator.RelationDefinition
	TargetTable string
}

func (c *ModifyForeignKey) Type() ChangeType  { return ChangeTypeModifyForeignKey }
func (c *ModifyForeignKey) TableName() string { return c.Table }
func (c *ModifyForeignKey) Reversible() bool  { return true }

// AddConstraint represents adding a constraint
type AddConstraint struct {
	Table      string
	Constraint generator.ConstraintDefinition
}

func (c *AddConstraint) Type() ChangeType  { return ChangeTypeAddConstraint }
func (c *AddConstraint) TableName() string { return c.Table }
func (c *AddConstraint) Reversible() bool  { return true }

// DropConstraint represents dropping a constraint
type DropConstraint struct {
	Table          string
	ConstraintName string
}

func (c *DropConstraint) Type() ChangeType  { return ChangeTypeDropConstraint }
func (c *DropConstraint) TableName() string { return c.Table }
func (c *DropConstraint) Reversible() bool  { return true }

// RunSQL represents a raw SQL data migration operation
type RunSQL struct {
	SQL        string // SQL to execute in up migration
	ReverseSQL string // SQL to execute in down migration (optional)
	CanReverse bool   // Whether this operation can be reversed
}

func (c *RunSQL) Type() ChangeType  { return ChangeTypeRunSQL }
func (c *RunSQL) TableName() string { return "" } // Data migrations don't target a specific table
func (c *RunSQL) Reversible() bool  { return c.CanReverse }

// RunGo represents a Go code execution data migration operation
type RunGo struct {
	UpFunc     string // Function name or code to execute in up migration
	DownFunc   string // Function name or code to execute in down migration (optional)
	CanReverse bool   // Whether this operation can be reversed
}

func (c *RunGo) Type() ChangeType  { return ChangeTypeRunGo }
func (c *RunGo) TableName() string { return "" } // Data migrations don't target a specific table
func (c *RunGo) Reversible() bool  { return c.CanReverse }

// UnknownChange represents a SQL statement that couldn't be parsed
// but should be preserved for state reconstruction
type UnknownChange struct {
	SQL string // Raw SQL text
}

func (c *UnknownChange) Type() ChangeType  { return ChangeTypeUnknown }
func (c *UnknownChange) TableName() string { return "" }
func (c *UnknownChange) Reversible() bool  { return false }

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return string(result)
}
