package state

import (
	"github.com/forgego/forge/db/migrate/core"
)

// SchemaState represents the current database schema state
type SchemaState struct {
	Tables map[string]*TableState
}

// TableState represents the state of a single table
type TableState struct {
	Name        string
	Columns     map[string]*ColumnState
	Indexes     map[string]*IndexState
	ForeignKeys map[string]*ForeignKeyState
	Constraints map[string]*ConstraintState
}

// ColumnState represents the state of a column
type ColumnState struct {
	Name          string
	Type          string
	GoType        string
	Required      bool
	PrimaryKey    bool
	AutoIncrement bool
	Unique        bool
	Default       interface{}
	Options       map[string]interface{}
}

// IndexState represents the state of an index
type IndexState struct {
	Name   string
	Fields []string
	Unique bool
}

// ForeignKeyState represents the state of a foreign key
type ForeignKeyState struct {
	Name         string
	Column       string
	TargetTable  string
	TargetColumn string
	OnDelete     string
	OnUpdate     string
}

// ConstraintState represents the state of a constraint
type ConstraintState struct {
	Name      string
	Type      string
	Condition string
	Fields    []string
}

// StateManager manages schema state
type StateManager interface {
	// Load loads the current schema state
	Load() (*SchemaState, error)

	// Apply applies changes to the state
	Apply(changes []core.Change) error

	// GetState returns the current state
	GetState() *SchemaState
}
