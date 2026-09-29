package sql

import (
	"fmt"
	"github.com/forgego/forge/schema"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// orderChanges orders changes by dependency
func orderChanges(changes []core.Change) []core.Change {
	ordered := make([]core.Change, 0, len(changes))

	// Foreign keys and constraints are dropped before anything else, so a
	// constraint re-added under the same name, or a dropped column, follows
	// its drop, and the down migration re-adds them after everything they
	// depend on has been restored.
	for _, change := range changes {
		if isConstraintDrop(change) {
			ordered = append(ordered, change)
		}
	}

	// First pass: CreateTable
	for _, change := range changes {
		if change.Type() == core.ChangeTypeCreateTable {
			ordered = append(ordered, change)
		}
	}

	// Second pass: AddColumn, RenameColumn, ModifyColumn
	for _, change := range changes {
		ct := change.Type()
		if ct == core.ChangeTypeAddColumn || ct == core.ChangeTypeRenameColumn || ct == core.ChangeTypeModifyColumn {
			ordered = append(ordered, change)
		}
	}

	// Third pass: AddForeignKey, ModifyForeignKey
	for _, change := range changes {
		if change.Type() == core.ChangeTypeAddForeignKey || change.Type() == core.ChangeTypeModifyForeignKey {
			ordered = append(ordered, change)
		}
	}

	// Fourth pass: AddIndex, ModifyIndex, AddConstraint
	for _, change := range changes {
		ct := change.Type()
		if ct == core.ChangeTypeAddIndex || ct == core.ChangeTypeModifyIndex || ct == core.ChangeTypeAddConstraint {
			ordered = append(ordered, change)
		}
	}

	// Fifth pass: Everything else
	for _, change := range changes {
		ct := change.Type()
		if ct != core.ChangeTypeCreateTable && ct != core.ChangeTypeAddColumn && ct != core.ChangeTypeRenameColumn &&
			ct != core.ChangeTypeModifyColumn && ct != core.ChangeTypeAddForeignKey && ct != core.ChangeTypeModifyForeignKey &&
			ct != core.ChangeTypeAddIndex && ct != core.ChangeTypeModifyIndex && ct != core.ChangeTypeAddConstraint &&
			!isConstraintDrop(change) {
			ordered = append(ordered, change)
		}
	}

	return ordered
}

func isConstraintDrop(change core.Change) bool {
	ct := change.Type()
	return ct == core.ChangeTypeDropForeignKey || ct == core.ChangeTypeDropConstraint
}

// CascadeAction returns the SQL referential action for a model or SQL cascade name.
func CascadeAction(cascade string) string {
	return mapCascadeType(cascade)
}

// ColumnDefinition renders the column definition the driver's builder emits for field.
func ColumnDefinition(driver core.Driver, field generator.FieldDefinition) (string, error) {
	b := &baseBuilder{isSQLite: driver.IsSQLite(), isPostgres: driver.IsPostgreSQL()}
	return b.BuildColumnDefinition(field)
}

// mapCascadeType maps cascade type strings to SQL
func mapCascadeType(cascade string) string {
	cascade = strings.TrimPrefix(cascade, "Cascade")
	switch strings.ToUpper(cascade) {
	case "CASCADE":
		return "CASCADE"
	case "SET_NULL", "SET NULL", "SETNULL":
		return "SET NULL"
	case "PROTECT", "RESTRICT":
		return "RESTRICT"
	case "SET_DEFAULT", "SET DEFAULT", "SETDEFAULT":
		return "SET DEFAULT"
	case "DO_NOTHING", "NO ACTION", "NOACTION":
		return "NO ACTION"
	default:
		return "NO ACTION"
	}
}

// formatDefaultValue formats default value for SQL
func formatDefaultValue(value interface{}, goType string, fieldType string, fieldOptions map[string]interface{}, isSQLite bool) string {
	// A string is a literal and is quoted. A SQL expression default belongs in
	// DBDefault, which the builder renders as is; only these common
	// current-time and random functions are still recognized in Default.
	if str, ok := value.(string); ok {
		if schema.IsDatabaseFunctionDefault(str) {
			return str // Return unquoted SQL expression
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
	}

	switch v := value.(type) {
	case bool:
		if isSQLite {
			if v {
				return "1"
			}
			return "0"
		}
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		// For Decimal fields, format to match decimal places
		if fieldType == "Decimal" {
			decimalPlaces := 2
			if dp, ok := fieldOptions["decimal_places"].(int); ok && dp >= 0 {
				decimalPlaces = dp
			}
			return fmt.Sprintf("%.*f", decimalPlaces, v)
		}
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// mapFieldTypeToSQL maps field types to SQL types
func mapFieldTypeToSQL(field generator.FieldDefinition, isSQLite, isPostgres bool) string {
	if dbType, ok := field.Options["db_type"].(string); ok && dbType != "" {
		return dbType
	}
	if sqlType, ok := field.Options[core.SQLTypeOption].(string); ok && sqlType != "" {
		return sqlType
	}

	switch field.Type {
	case "Decimal":
		maxDigits := 10
		decimalPlaces := 2
		if md, ok := field.Options["max_digits"].(int); ok && md > 0 {
			maxDigits = md
		}
		if dp, ok := field.Options["decimal_places"].(int); ok && dp >= 0 {
			decimalPlaces = dp
		}
		return fmt.Sprintf("NUMERIC(%d, %d)", maxDigits, decimalPlaces)

	case "JSON":
		if isPostgres {
			return "JSONB"
		}
		if isSQLite {
			return "TEXT"
		}
		return "JSON"

	case "Date":
		return "DATE"

	case "DateTime":
		if isPostgres {
			return "TIMESTAMP WITH TIME ZONE"
		}
		return "TIMESTAMP"

	case "Time":
		return "TIME"

	case "String":
		if maxLen, ok := field.Options["max_length"].(int); ok && maxLen > 0 {
			return fmt.Sprintf("VARCHAR(%d)", maxLen)
		}
		return "TEXT"

	case "Text":
		return "TEXT"

	case "UUID":
		if isPostgres {
			return "UUID"
		}
		return "CHAR(36)"

	case "Bytes":
		if isSQLite {
			return "BLOB"
		}
		if isPostgres {
			return "BYTEA"
		}
		return "BLOB"

	case "Float64":
		if isSQLite {
			return "REAL"
		}
		return "DOUBLE PRECISION"

	case "Int64", "Int":
		if isSQLite {
			return "INTEGER"
		}
		return "BIGINT"

	case "Int32":
		return "INTEGER"

	case "Bool":
		if isSQLite {
			return "INTEGER"
		}
		return "BOOLEAN"

	case "Email", "URL":
		if maxLen, ok := field.Options["max_length"].(int); ok && maxLen > 0 {
			return fmt.Sprintf("VARCHAR(%d)", maxLen)
		}
		return "TEXT"
	}

	// Fallback to Go type
	return mapGoTypeToSQL(field.GoType, isSQLite, isPostgres)
}

// mapGoTypeToSQL maps Go types to SQL types (fallback)
func mapGoTypeToSQL(goType string, isSQLite, isPostgres bool) string {
	switch goType {
	case "int64", "int":
		if isSQLite {
			return "INTEGER"
		}
		return "BIGINT"
	case "int32":
		return "INTEGER"
	case "string":
		return "TEXT"
	case "bool":
		if isSQLite {
			return "INTEGER"
		}
		return "BOOLEAN"
	case "time.Time", "*time.Time":
		return "TIMESTAMP"
	case "float64":
		if isSQLite {
			return "REAL"
		}
		return "DOUBLE PRECISION"
	case "float32":
		return "REAL"
	case "[]byte":
		if isSQLite {
			return "BLOB"
		}
		return "BYTEA"
	default:
		return "TEXT"
	}
}
