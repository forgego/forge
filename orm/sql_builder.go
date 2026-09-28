package orm

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/db/dialect"
)

// JoinResolver resolves a relation path to an alias and target database column.
// parts contains all segments of the path (e.g. ["customer", "name"]).
type JoinResolver func(parts []string) (alias string, column string, err error)

// SQLBuilder provides safe SQL building with proper escaping and parameter binding
type SQLBuilder struct {
	paramIndex   int
	args         []interface{}
	dialect      dialect.Dialect
	placeholder  func(int) string
	joinResolver JoinResolver
}

func defaultPlaceholder(position int) string {
	return fmt.Sprintf("$%d", position)
}

// NewSQLBuilder creates a new SQL builder
func NewSQLBuilder() *SQLBuilder {
	return &SQLBuilder{
		paramIndex:  1,
		args:        []interface{}{},
		placeholder: defaultPlaceholder,
	}
}

// NewSQLBuilderWithDialect creates a new SQL builder with the specified dialect
func NewSQLBuilderWithDialect(d dialect.Dialect) *SQLBuilder {
	b := &SQLBuilder{
		paramIndex: 1,
		args:       []interface{}{},
		dialect:    d,
	}
	if d != nil {
		b.placeholder = d.Placeholder
	} else {
		b.placeholder = defaultPlaceholder
	}
	return b
}

// Placeholder returns the placeholder for the given position
func (b *SQLBuilder) Placeholder(position int) string {
	if b != nil && b.placeholder != nil {
		return b.placeholder(position)
	}
	return defaultPlaceholder(position)
}

// SetPlaceholder sets a custom placeholder function
func (b *SQLBuilder) SetPlaceholder(fn func(int) string) {
	if b != nil {
		b.placeholder = fn
	}
}

// CaseInsensitiveLike returns a case-insensitive LIKE comparison expression.
func (b *SQLBuilder) CaseInsensitiveLike(field, placeholder string) string {
	if b != nil && b.dialect != nil {
		if ciLiker, ok := b.dialect.(dialect.CaseInsensitiveLiker); ok {
			return ciLiker.CaseInsensitiveLike(field, placeholder)
		}
	}
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", field, placeholder)
}

// SetJoinResolver sets an optional join resolver for resolving relation paths.
func (b *SQLBuilder) SetJoinResolver(r JoinResolver) {
	if b != nil {
		b.joinResolver = r
	}
}

// resolveColumn resolves a field path to an escaped column identifier, resolving relation joins if a resolver is set.
func (b *SQLBuilder) resolveColumn(fieldPath string) (string, error) {
	if b == nil || b.joinResolver == nil {
		return EscapeIdentifier(fieldPath), nil
	}
	parts := splitFieldPath(fieldPath)
	if len(parts) <= 1 {
		return EscapeIdentifier(fieldPath), nil
	}
	alias, column, err := b.joinResolver(parts)
	if err != nil {
		return "", err
	}
	return EscapeIdentifier(alias) + "." + EscapeIdentifier(column), nil
}

func (b *SQLBuilder) isSQLite() bool {
	if b == nil || b.dialect == nil {
		return false
	}
	name := strings.ToLower(b.dialect.Name())
	return name == "sqlite" || name == "sqlite3"
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
	placeholder := b.Placeholder(b.paramIndex)
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
	if b == nil {
		return "", nil
	}

	var whereParts []string
	var allArgs []interface{}

	// Add conditions
	for _, cond := range conditions {
		adapter := newQueryExprAdapter(cond)
		sql, condArgs, _ := adapter.ToSQL(b)
		whereParts = append(whereParts, sql)
		allArgs = append(allArgs, condArgs...)
	}

	// Add excludes with NOT
	for _, exclude := range excludes {
		adapter := newQueryExprAdapter(exclude)
		sql, excludeArgs, _ := adapter.ToSQL(b)
		whereParts = append(whereParts, "NOT ("+sql+")")
		allArgs = append(allArgs, excludeArgs...)
	}

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
			col, err := b.resolveColumn(fieldName)
			if err != nil {
				col = EscapeIdentifier(fieldName)
			}
			parts = append(parts, col+" DESC")
		} else {
			col, err := b.resolveColumn(field)
			if err != nil {
				col = EscapeIdentifier(field)
			}
			parts = append(parts, col+" ASC")
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
		placeholder := b.Placeholder(paramIndex)
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
		placeholder := b.Placeholder(paramIndex)
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
