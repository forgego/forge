package sql

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// orderChanges orders changes by dependency
func orderChanges(changes []core.Change) []core.Change {
	ordered := make([]core.Change, 0, len(changes))

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
			ct != core.ChangeTypeAddIndex && ct != core.ChangeTypeModifyIndex && ct != core.ChangeTypeAddConstraint {
			ordered = append(ordered, change)
		}
	}

	return ordered
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

// formatDefaultValue formats default value for SQL
func formatDefaultValue(value interface{}, goType string, fieldType string, fieldOptions map[string]interface{}, isSQLite bool) string {
	// Check if value is a SQL expression (string that looks like a function call)
	if str, ok := value.(string); ok {
		// Common SQL functions that should not be quoted
		sqlFunctions := []string{"now()", "CURRENT_TIMESTAMP", "CURRENT_DATE", "CURRENT_TIME",
			"uuid_generate_v4()", "gen_random_uuid()", "random()"}
		for _, fn := range sqlFunctions {
			if strings.EqualFold(str, fn) {
				return str // Return unquoted SQL expression
			}
		}
		// Check if it looks like a function call (contains parentheses)
		if strings.Contains(str, "(") && strings.Contains(str, ")") {
			return str // Likely a SQL function, return unquoted
		}
		// Regular string, quote it
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
