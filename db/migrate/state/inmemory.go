package state

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// InMemoryState is an in-memory implementation of StateManager
type InMemoryState struct {
	state *SchemaState
}

// NewInMemoryState creates a new in-memory state manager
func NewInMemoryState() StateManager {
	return &InMemoryState{
		state: &SchemaState{
			Tables: make(map[string]*TableState),
		},
	}
}

// Load returns the current state (no-op for in-memory)
func (s *InMemoryState) Load() (*SchemaState, error) {
	return s.state, nil
}

// Apply applies changes to the state
func (s *InMemoryState) Apply(changes []core.Change) error {
	for _, change := range changes {
		switch c := change.(type) {
		case *core.CreateTable:
			if err := s.applyCreateTable(c); err != nil {
				return err
			}
		case *core.DropTable:
			delete(s.state.Tables, c.Table)
		case *core.RenameTable:
			s.applyRenameTable(c)
		case *core.AddColumn:
			if err := s.applyAddColumn(c); err != nil {
				return err
			}
		case *core.DropColumn:
			if table, exists := s.state.Tables[c.Table]; exists {
				delete(table.Columns, c.ColumnName)
			}
		case *core.ModifyColumn:
			if err := s.applyModifyColumn(c); err != nil {
				return err
			}
		case *core.AlterColumn:
			if err := s.applyAlterColumn(c); err != nil {
				return err
			}
		case *core.RenameColumn:
			s.applyRenameColumn(c)
		case *core.AddIndex:
			if err := s.applyAddIndex(c); err != nil {
				return err
			}
		case *core.DropIndex:
			if table, exists := s.state.Tables[c.Table]; exists {
				delete(table.Indexes, c.IndexName)
			}
		case *core.AddForeignKey:
			if err := s.applyAddForeignKey(c); err != nil {
				return err
			}
		case *core.DropForeignKey:
			if table, exists := s.state.Tables[c.Table]; exists {
				delete(table.ForeignKeys, c.FKName)
			}
		case *core.AddConstraint:
			if err := s.applyAddConstraint(c); err != nil {
				return err
			}
		case *core.DropConstraint:
			if table, exists := s.state.Tables[c.Table]; exists {
				delete(table.Constraints, c.ConstraintName)
			}
		}
	}
	return nil
}

// GetState returns the current state
func (s *InMemoryState) GetState() *SchemaState {
	return s.state
}

func (s *InMemoryState) applyCreateTable(c *core.CreateTable) error {
	tableName := c.TableName()
	tableState := &TableState{
		Name:        tableName,
		Columns:     make(map[string]*ColumnState),
		Indexes:     make(map[string]*IndexState),
		ForeignKeys: make(map[string]*ForeignKeyState),
		Constraints: make(map[string]*ConstraintState),
	}

	// Add columns
	for _, field := range c.Table.Fields {
		colState := &ColumnState{
			Name:          field.Name,
			Type:          field.Type,
			GoType:        field.GoType,
			Required:      field.Required,
			PrimaryKey:    field.PrimaryKey,
			AutoIncrement: field.AutoIncrement,
			Unique:        false,
			Default:       field.Default,
			Options:       field.Options,
		}
		if unique, ok := field.Options["unique"].(bool); ok && unique {
			colState.Unique = unique
		}
		tableState.Columns[field.Name] = colState
	}

	// Add indexes
	for _, idx := range c.Table.Meta.Indexes {
		idxState := &IndexState{
			Name:   idx.Name,
			Fields: idx.Fields,
			Unique: idx.Unique,
		}
		if idxState.Name == "" {
			idxState.Name = fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Fields, "_"))
		}
		tableState.Indexes[idxState.Name] = idxState
	}

	// Add constraints
	for _, constr := range c.Table.Meta.Constraints {
		constrState := &ConstraintState{
			Name:      constr.Name,
			Type:      constr.Type,
			Condition: constr.Condition,
			Fields:    constr.Fields,
		}
		tableState.Constraints[constr.Name] = constrState
	}

	s.state.Tables[tableName] = tableState
	return nil
}

func (s *InMemoryState) applyAddColumn(c *core.AddColumn) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}

	colState := &ColumnState{
		Name:          c.Column.Name,
		Type:          c.Column.Type,
		GoType:        c.Column.GoType,
		Required:      c.Column.Required,
		PrimaryKey:    c.Column.PrimaryKey,
		AutoIncrement: c.Column.AutoIncrement,
		Unique:        false,
		Default:       c.Column.Default,
		Options:       c.Column.Options,
	}
	if unique, ok := c.Column.Options["unique"].(bool); ok && unique {
		colState.Unique = unique
	}
	table.Columns[c.Column.Name] = colState
	return nil
}

func (s *InMemoryState) applyModifyColumn(c *core.ModifyColumn) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}

	// Delete old column if name changed
	if c.OldColumn.Name != c.NewColumn.Name {
		delete(table.Columns, c.OldColumn.Name)
	}

	colState := &ColumnState{
		Name:          c.NewColumn.Name,
		Type:          c.NewColumn.Type,
		GoType:        c.NewColumn.GoType,
		Required:      c.NewColumn.Required,
		PrimaryKey:    c.NewColumn.PrimaryKey,
		AutoIncrement: c.NewColumn.AutoIncrement,
		Unique:        false,
		Default:       c.NewColumn.Default,
		Options:       c.NewColumn.Options,
	}
	if unique, ok := c.NewColumn.Options["unique"].(bool); ok && unique {
		colState.Unique = unique
	}
	table.Columns[c.NewColumn.Name] = colState
	return nil
}

// typeOptionKeys are the column options derived from its SQL type, which a
// TYPE clause replaces.
var typeOptionKeys = []string{core.SQLTypeOption, "db_type", "max_length", "max_digits", "decimal_places"}

// defaultOptionKeys are the column options that render its default, which a
// SET DEFAULT or DROP DEFAULT clause replaces.
var defaultOptionKeys = []string{"db_default", "auto_now", "auto_now_add"}

func (s *InMemoryState) applyAlterColumn(c *core.AlterColumn) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}
	col, exists := table.Columns[c.Column]
	if !exists {
		return fmt.Errorf("column %s.%s does not exist", c.Table, c.Column)
	}
	options := make(map[string]interface{}, len(col.Options))
	for key, value := range col.Options {
		options[key] = value
	}
	if c.NewType != nil {
		col.Type = c.NewType.Type
		col.GoType = c.NewType.GoType
		for _, key := range typeOptionKeys {
			delete(options, key)
		}
		for key, value := range c.NewType.Options {
			options[key] = value
		}
	}
	if c.SetDefault || c.DropDefault {
		// A default replaces whatever set the previous one, literal,
		// expression or auto timestamp.
		for _, key := range defaultOptionKeys {
			delete(options, key)
		}
		col.Default = nil
	}
	if c.SetDefault {
		col.Default = c.Default
		if c.DBDefault != "" {
			options["db_default"] = c.DBDefault
		}
	}
	col.Options = options
	if c.NotNull != nil {
		col.Required = *c.NotNull
	}
	return nil
}

func (s *InMemoryState) applyAddIndex(c *core.AddIndex) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}

	idxState := &IndexState{
		Name:   c.Index.Name,
		Fields: c.Index.Fields,
		Unique: c.Index.Unique,
	}
	if idxState.Name == "" {
		idxState.Name = fmt.Sprintf("idx_%s_%s", c.Table, strings.Join(c.Index.Fields, "_"))
	}
	table.Indexes[idxState.Name] = idxState
	return nil
}

func (s *InMemoryState) applyAddForeignKey(c *core.AddForeignKey) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}

	fkName := fmt.Sprintf("fk_%s_%s", c.Table, c.Relation.Name)
	fkState := &ForeignKeyState{
		Name:         fkName,
		Column:       c.Relation.Name,
		TargetTable:  c.TargetTable,
		TargetColumn: "id", // Default to id
		OnDelete:     "NO ACTION",
		OnUpdate:     "NO ACTION",
	}

	if onDelete, ok := c.Relation.Options["on_delete"].(string); ok {
		fkState.OnDelete = mapCascadeType(onDelete)
	}
	if onUpdate, ok := c.Relation.Options["on_update"].(string); ok {
		fkState.OnUpdate = mapCascadeType(onUpdate)
	}

	table.ForeignKeys[fkName] = fkState
	return nil
}

func (s *InMemoryState) applyAddConstraint(c *core.AddConstraint) error {
	table, exists := s.state.Tables[c.Table]
	if !exists {
		return fmt.Errorf("table %s does not exist", c.Table)
	}

	constrState := &ConstraintState{
		Name:      c.Constraint.Name,
		Type:      c.Constraint.Type,
		Condition: c.Constraint.Condition,
		Fields:    c.Constraint.Fields,
	}
	table.Constraints[c.Constraint.Name] = constrState
	return nil
}

// mapCascadeType maps cascade type strings to SQL
func mapCascadeType(cascade string) string {
	switch cascade {
	case "CASCADE":
		return "CASCADE"
	case "SET_NULL", "SET NULL":
		return "SET NULL"
	case "PROTECT":
		return "RESTRICT"
	case "SET_DEFAULT", "SET DEFAULT":
		return "SET DEFAULT"
	case "DO_NOTHING", "NO ACTION":
		return "NO ACTION"
	default:
		return "NO ACTION"
	}
}
