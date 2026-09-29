package sql

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// baseBuilder contains common SQL building logic
type baseBuilder struct {
	isSQLite   bool
	isPostgres bool
}

// BuildColumnDefinition builds SQL column definition from field
func (b *baseBuilder) BuildColumnDefinition(field generator.FieldDefinition) (string, error) {
	columnName := field.Name
	if dbColumn, ok := field.Options["db_column"].(string); ok && dbColumn != "" {
		columnName = dbColumn
	}
	sqlType := mapFieldTypeToSQL(field, b.isSQLite, b.isPostgres)

	var parts []string
	parts = append(parts, fmt.Sprintf(`"%s"`, columnName))
	parts = append(parts, sqlType)

	isGenerated := false
	if generated, ok := field.Options["generated"].(bool); ok && generated {
		if expr, ok := field.Options["generated_expr"].(string); ok && expr != "" {
			parts = append(parts, fmt.Sprintf("GENERATED ALWAYS AS (%s)", expr))
			if b.isPostgres {
				parts = append(parts, "STORED")
			} else if b.isSQLite {
				if stored, ok := field.Options["generated_stored"].(bool); ok && stored {
					parts = append(parts, "STORED")
				} else {
					parts = append(parts, "VIRTUAL")
				}
			}
			isGenerated = true
		}
	}

	// Handle primary key and auto-increment
	if b.isSQLite {
		if field.PrimaryKey && field.AutoIncrement {
			parts = []string{fmt.Sprintf(`"%s"`, columnName), "INTEGER", "PRIMARY KEY", "AUTOINCREMENT"}
		} else if field.PrimaryKey {
			parts = append(parts, "PRIMARY KEY")
		}
	} else {
		if field.PrimaryKey {
			parts = append(parts, "PRIMARY KEY")
		}
		if field.AutoIncrement {
			parts = append(parts, "GENERATED ALWAYS AS IDENTITY")
		}
	}

	// Handle required/optional
	if field.Required && !field.PrimaryKey {
		parts = append(parts, "NOT NULL")
	}

	// Handle default values
	if !isGenerated {
		if defaultExpr := b.columnDefault(field); defaultExpr != "" {
			parts = append(parts, "DEFAULT "+defaultExpr)
		}
		// Make created_at NOT NULL when AutoNowAdd is set
		if impliesNotNull(field) && !field.Required {
			parts = append(parts, "NOT NULL")
		}
	}

	// Handle unique constraint
	if unique, ok := field.Options["unique"].(bool); ok && unique {
		parts = append(parts, "UNIQUE")
	}

	return strings.Join(parts, " "), nil
}

// columnDefault returns the SQL expression of a column's DEFAULT clause, or ""
// when it has none: a DB default expression verbatim, the current time for
// AutoNowAdd and AutoNow, or the formatted Default value. A generated column
// has no default.
func (b *baseBuilder) columnDefault(field generator.FieldDefinition) string {
	if generated, ok := field.Options["generated"].(bool); ok && generated {
		if expr, ok := field.Options["generated_expr"].(string); ok && expr != "" {
			return ""
		}
	}
	// SQLite accepts no function call as a column default without
	// parentheses; CURRENT_TIMESTAMP is its current-time default.
	currentTime := "now()"
	if b.isSQLite {
		currentTime = "CURRENT_TIMESTAMP"
	}
	if dbDefault, ok := field.Options["db_default"].(string); ok && dbDefault != "" {
		return dbDefault
	}
	if autoNowAdd, ok := field.Options["auto_now_add"].(bool); ok && autoNowAdd {
		return currentTime
	}
	if autoNow, ok := field.Options["auto_now"].(bool); ok && autoNow {
		return currentTime
	}
	if field.Default != nil {
		return formatDefaultValue(field.Default, field.GoType, field.Type, field.Options, b.isSQLite)
	}
	return ""
}

// impliesNotNull reports whether a column that is not a primary key is NOT
// NULL although not declared Required: an AutoNowAdd column without a DB
// default, which is always set on insert.
func impliesNotNull(field generator.FieldDefinition) bool {
	if field.PrimaryKey {
		return false
	}
	if dbDefault, ok := field.Options["db_default"].(string); ok && dbDefault != "" {
		return false
	}
	autoNowAdd, ok := field.Options["auto_now_add"].(bool)
	return ok && autoNowAdd
}

// unalterableColumnChange returns the first part of a column's definition that
// differs between old and new and that ALTER COLUMN cannot change here, or ""
// when there is none: its name, primary key, identity, UNIQUE or generated
// expression. The builder fails on such a change rather than render no SQL,
// which would hide it.
func unalterableColumnChange(old, new generator.FieldDefinition) string {
	switch {
	case columnName(old) != columnName(new):
		return "column name (db_column)"
	case old.PrimaryKey != new.PrimaryKey:
		return "PRIMARY KEY"
	case old.AutoIncrement != new.AutoIncrement:
		return "identity (AutoIncrement)"
	case optionBool(old, "unique") != optionBool(new, "unique"):
		return "UNIQUE"
	case generatedExpr(old) != generatedExpr(new):
		return "generated expression"
	}
	return ""
}

func columnName(field generator.FieldDefinition) string {
	if dbColumn, ok := field.Options["db_column"].(string); ok && dbColumn != "" {
		return dbColumn
	}
	return field.Name
}

func optionBool(field generator.FieldDefinition, key string) bool {
	value, _ := field.Options[key].(bool)
	return value
}

// generatedExpr returns a generated column's expression and storage, or ""
// for an ordinary column.
func generatedExpr(field generator.FieldDefinition) string {
	if !optionBool(field, "generated") {
		return ""
	}
	expr, _ := field.Options["generated_expr"].(string)
	if expr == "" {
		return ""
	}
	return fmt.Sprintf("%s stored=%t", expr, optionBool(field, "generated_stored"))
}

// unalterableColumnError reports a column change that makemigrations cannot
// express as ALTER COLUMN.
func unalterableColumnError(table string, column generator.FieldDefinition, part string) error {
	return core.NewMigrationError(
		core.ErrInvalidChange,
		fmt.Sprintf("changing the %s of existing column %s.%s is not supported by makemigrations; "+
			"revert the model change, or see \"Changes makemigrations cannot generate\" in the migrations guide",
			part, table, column.Name),
		nil,
	)
}

// BuildCreateTable generates CREATE TABLE statement
func (b *baseBuilder) BuildCreateTable(c *core.CreateTable) (string, error) {
	tableName := c.TableName()
	var parts []string

	parts = append(parts, fmt.Sprintf("-- Create table: %s", tableName))
	parts = append(parts, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName))

	var columnDefs []string
	for _, field := range c.Table.Fields {
		colDef, err := b.BuildColumnDefinition(field)
		if err != nil {
			return "", fmt.Errorf("failed to build column %s: %w", field.Name, err)
		}
		columnDefs = append(columnDefs, "    "+colDef)
	}

	parts = append(parts, strings.Join(columnDefs, ",\n"))
	parts = append(parts, ");")

	return strings.Join(parts, "\n"), nil
}

// BuildAddColumn generates ALTER TABLE ADD COLUMN statement
func (b *baseBuilder) BuildAddColumn(c *core.AddColumn) (string, error) {
	colDef, err := b.BuildColumnDefinition(c.Column)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", c.Table, colDef), nil
}

// BuildAddIndex generates CREATE INDEX statement
func (b *baseBuilder) BuildAddIndex(c *core.AddIndex) (string, error) {
	if len(c.Index.Fields) == 0 {
		return "", nil
	}

	escapedFields := make([]string, len(c.Index.Fields))
	for i, field := range c.Index.Fields {
		escapedFields[i] = fmt.Sprintf(`"%s"`, field)
	}

	indexType := "INDEX"
	if c.Index.Unique {
		indexType = "UNIQUE INDEX"
	}

	indexName := c.Index.Name
	if indexName == "" {
		indexName = fmt.Sprintf("idx_%s_%s", c.Table, strings.Join(c.Index.Fields, "_"))
	}

	return fmt.Sprintf("CREATE %s IF NOT EXISTS %s ON %s (%s);",
		indexType, indexName, c.Table, strings.Join(escapedFields, ", ")), nil
}

// BuildModifyIndex generates DROP and CREATE INDEX statements
func (b *baseBuilder) BuildModifyIndex(c *core.ModifyIndex) (string, error) {
	oldIndexName := c.OldIndex.Name
	if oldIndexName == "" {
		oldIndexName = fmt.Sprintf("idx_%s_%s", c.Table, strings.Join(c.OldIndex.Fields, "_"))
	}

	var statements []string
	statements = append(statements, fmt.Sprintf("DROP INDEX IF EXISTS %s;", oldIndexName))

	newIndex := &core.AddIndex{
		Table: c.Table,
		Index: c.NewIndex,
	}
	newIndexSQL, err := b.BuildAddIndex(newIndex)
	if err != nil {
		return "", err
	}
	statements = append(statements, newIndexSQL)

	return strings.Join(statements, "\n"), nil
}

// BuildAddForeignKey generates ALTER TABLE ADD CONSTRAINT FOREIGN KEY statement
func (b *baseBuilder) BuildAddForeignKey(c *core.AddForeignKey) (string, error) {
	// Validate target table is not empty
	if c.TargetTable == "" {
		return "", core.NewMigrationError(
			core.ErrInvalidChange,
			fmt.Sprintf("foreign key %s.%s has empty target table", c.Table, c.Relation.Name),
			nil,
		)
	}

	onDelete := "NO ACTION"
	if onDeleteVal, ok := c.Relation.Options["on_delete"].(string); ok {
		onDelete = mapCascadeType(onDeleteVal)
	}

	onUpdate := "NO ACTION"
	if onUpdateVal, ok := c.Relation.Options["on_update"].(string); ok {
		onUpdate = mapCascadeType(onUpdateVal)
	}

	fkName := fmt.Sprintf("fk_%s_%s", c.Table, c.Relation.Name)

	// Use DO block to check if constraint exists before adding (PostgreSQL-specific)
	// This prevents errors when constraint already exists (e.g., in incremental migrations)
	return fmt.Sprintf(`DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = '%s' 
        AND conrelid = '%s'::regclass
    ) THEN
        ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY ("%s") REFERENCES %s (id) ON DELETE %s ON UPDATE %s;
    END IF;
END $$;`,
		fkName, c.Table, c.Table, fkName, c.Relation.Name, c.TargetTable, onDelete, onUpdate), nil
}

// BuildModifyForeignKey generates DROP and ADD FOREIGN KEY statements
func (b *baseBuilder) BuildModifyForeignKey(c *core.ModifyForeignKey) (string, error) {
	oldFKName := fmt.Sprintf("fk_%s_%s", c.Table, c.OldFK.Name)

	var statements []string
	statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;", c.Table, oldFKName))

	newFK := &core.AddForeignKey{
		Table:       c.Table,
		Relation:    c.NewFK,
		TargetTable: c.TargetTable,
	}
	newFKSQL, err := b.BuildAddForeignKey(newFK)
	if err != nil {
		return "", err
	}
	statements = append(statements, newFKSQL)

	return strings.Join(statements, "\n"), nil
}

// BuildAddConstraint generates ALTER TABLE ADD CONSTRAINT statement
func (b *baseBuilder) BuildAddConstraint(c *core.AddConstraint) (string, error) {
	constraintSQL, err := constraintBody(c.Constraint)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;",
		c.Table, c.Constraint.Name, constraintSQL), nil
}

// constraintBody renders a table constraint without its name, as
// CHECK (condition) or UNIQUE ("field", ...).
func constraintBody(constraint generator.ConstraintDefinition) (string, error) {
	var constraintSQL string
	switch strings.ToUpper(constraint.Type) {
	case "CHECK":
		if constraint.Condition != "" {
			constraintSQL = fmt.Sprintf("CHECK (%s)", constraint.Condition)
		} else if len(constraint.Fields) > 0 {
			escapedFields := make([]string, len(constraint.Fields))
			for i, field := range constraint.Fields {
				escapedFields[i] = fmt.Sprintf(`"%s"`, field)
			}
			constraintSQL = fmt.Sprintf("CHECK (%s)", strings.Join(escapedFields, ", "))
		} else {
			return "", core.NewMigrationError(
				core.ErrInvalidChange,
				"CHECK constraint requires condition or fields",
				nil,
			)
		}
	case "UNIQUE":
		if len(constraint.Fields) > 0 {
			escapedFields := make([]string, len(constraint.Fields))
			for i, field := range constraint.Fields {
				escapedFields[i] = fmt.Sprintf(`"%s"`, field)
			}
			constraintSQL = fmt.Sprintf("UNIQUE (%s)", strings.Join(escapedFields, ", "))
		} else {
			return "", core.NewMigrationError(
				core.ErrInvalidChange,
				"UNIQUE constraint requires fields",
				nil,
			)
		}
	default:
		if constraint.Condition != "" {
			constraintSQL = constraint.Condition
		} else {
			return "", core.NewMigrationError(
				core.ErrInvalidChange,
				fmt.Sprintf("constraint type %s requires condition", constraint.Type),
				nil,
			)
		}
	}
	return constraintSQL, nil
}
