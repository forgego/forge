package orm

import (
	"fmt"
	"strings"
)

// SQLBuilder provides safe SQL building with proper escaping and parameter binding
type SQLBuilder struct {
	paramIndex int
	args       []interface{}
}

// NewSQLBuilder creates a new SQL builder
func NewSQLBuilder() *SQLBuilder {
	return &SQLBuilder{
		paramIndex: 1,
		args:       []interface{}{},
	}
}

// EscapeIdentifier escapes SQL identifiers (table/column names) to prevent SQL injection
// PostgreSQL uses double quotes for identifiers
func EscapeIdentifier(identifier string) string {
	// Replace any double quotes with escaped double quotes
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	// Wrap in double quotes
	return `"` + escaped + `"`
}

// EscapeIdentifierList escapes a list of identifiers
func EscapeIdentifierList(identifiers []string) []string {
	escaped := make([]string, len(identifiers))
	for i, id := range identifiers {
		escaped[i] = EscapeIdentifier(id)
	}
	return escaped
}

// AddArg adds an argument and returns a placeholder
func (b *SQLBuilder) AddArg(value interface{}) string {
	placeholder := fmt.Sprintf("$%d", b.paramIndex)
	b.args = append(b.args, value)
	b.paramIndex++
	return placeholder
}

// AddArgs adds multiple arguments and returns placeholders
func (b *SQLBuilder) AddArgs(values []interface{}) []string {
	placeholders := make([]string, len(values))
	for i, value := range values {
		placeholders[i] = b.AddArg(value)
	}
	return placeholders
}

// Args returns all collected arguments
func (b *SQLBuilder) Args() []interface{} {
	return b.args
}

// Reset resets the builder (useful for reusing)
func (b *SQLBuilder) Reset() {
	b.paramIndex = 1
	b.args = []interface{}{}
}

// BuildSelect builds a SELECT query
func (b *SQLBuilder) BuildSelect(table string, fields []string, distinct bool) string {
	var selectClause string
	if distinct {
		selectClause = "SELECT DISTINCT "
	} else {
		selectClause = "SELECT "
	}

	if len(fields) == 0 {
		selectClause += "*"
	} else {
		escapedFields := EscapeIdentifierList(fields)
		selectClause += strings.Join(escapedFields, ", ")
	}

	escapedTable := EscapeIdentifier(table)
	return selectClause + " FROM " + escapedTable
}

// BuildWhere builds a WHERE clause from conditions
func (b *SQLBuilder) BuildWhere(conditions []QueryExpr, excludes []QueryExpr) (string, []interface{}) {
	var whereParts []string
	var allArgs []interface{}
	paramIndex := b.paramIndex

	// Add conditions
	for _, cond := range conditions {
		sql, condArgs, nextIndex := cond.ToSQL(paramIndex)
		whereParts = append(whereParts, sql)
		allArgs = append(allArgs, condArgs...)
		paramIndex = nextIndex
	}

	// Add excludes with NOT
	for _, exclude := range excludes {
		sql, excludeArgs, nextIndex := exclude.ToSQL(paramIndex)
		whereParts = append(whereParts, "NOT ("+sql+")")
		allArgs = append(allArgs, excludeArgs...)
		paramIndex = nextIndex
	}

	// Update builder's param index
	b.paramIndex = paramIndex
	b.args = append(b.args, allArgs...)

	if len(whereParts) == 0 {
		return "", allArgs
	}

	return "WHERE " + strings.Join(whereParts, " AND "), allArgs
}

// BuildOrderBy builds an ORDER BY clause
func (b *SQLBuilder) BuildOrderBy(orderBy []string) string {
	if len(orderBy) == 0 {
		return ""
	}

	var parts []string
	for _, field := range orderBy {
		// Handle descending order (fields starting with "-")
		if strings.HasPrefix(field, "-") {
			fieldName := strings.TrimPrefix(field, "-")
			parts = append(parts, EscapeIdentifier(fieldName)+" DESC")
		} else {
			parts = append(parts, EscapeIdentifier(field)+" ASC")
		}
	}

	return "ORDER BY " + strings.Join(parts, ", ")
}

// BuildLimit builds a LIMIT clause
func (b *SQLBuilder) BuildLimit(limit *int) string {
	if limit == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d", *limit)
}

// BuildOffset builds an OFFSET clause
func (b *SQLBuilder) BuildOffset(offset *int) string {
	if offset == nil {
		return ""
	}
	return fmt.Sprintf("OFFSET %d", *offset)
}

// BuildUpdate builds an UPDATE query
func (b *SQLBuilder) BuildUpdate(table string, fields map[string]interface{}) (string, []interface{}) {
	if len(fields) == 0 {
		return "", nil
	}

	escapedTable := EscapeIdentifier(table)
	var setParts []string
	var updateArgs []interface{}
	paramIndex := b.paramIndex

	for field, value := range fields {
		escapedField := EscapeIdentifier(field)
		placeholder := fmt.Sprintf("$%d", paramIndex)
		setParts = append(setParts, escapedField+" = "+placeholder)
		updateArgs = append(updateArgs, value)
		b.args = append(b.args, value)
		paramIndex++
	}

	b.paramIndex = paramIndex

	query := fmt.Sprintf("UPDATE %s SET %s", escapedTable, strings.Join(setParts, ", "))
	return query, updateArgs
}

// BuildInsert builds an INSERT query
func (b *SQLBuilder) BuildInsert(table string, fields map[string]interface{}) (string, []interface{}) {
	if len(fields) == 0 {
		return "", nil
	}

	escapedTable := EscapeIdentifier(table)
	var fieldNames []string
	var placeholders []string
	paramIndex := b.paramIndex

	for field, value := range fields {
		escapedField := EscapeIdentifier(field)
		fieldNames = append(fieldNames, escapedField)
		placeholder := fmt.Sprintf("$%d", paramIndex)
		placeholders = append(placeholders, placeholder)
		b.args = append(b.args, value)
		paramIndex++
	}

	b.paramIndex = paramIndex

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		escapedTable,
		strings.Join(fieldNames, ", "),
		strings.Join(placeholders, ", "))

	return query, b.args
}

// BuildDelete builds a DELETE query
func (b *SQLBuilder) BuildDelete(table string) string {
	escapedTable := EscapeIdentifier(table)
	return "DELETE FROM " + escapedTable
}
