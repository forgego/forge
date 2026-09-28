package parse

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// TableParser parses CREATE TABLE statements to extract table structure
type TableParser struct {
	createTableRegex *regexp.Regexp
	columnRegex      *regexp.Regexp
}

// NewTableParser creates a new table parser
func NewTableParser() *TableParser {
	return &TableParser{
		createTableRegex: regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?`),
		columnRegex:      regexp.MustCompile(`["']?(\w+)["']?\s+(\w+(?:\([^)]+\))?)\s*(.*?)(?:,|$)`),
	}
}

// tableConstraintRegex matches a table-level constraint in CREATE TABLE,
// which is not a column.
var tableConstraintRegex = regexp.MustCompile(`(?i)^(?:CONSTRAINT\s|FOREIGN\s+KEY|PRIMARY\s+KEY\s*\(|UNIQUE\s*\(|CHECK\s*\()`)

// tableForeignKeyRegex matches the start of a table-level foreign key, as
// SQLite migrations declare them inside CREATE TABLE.
var tableForeignKeyRegex = regexp.MustCompile(`(?i)^(?:CONSTRAINT\s+["']?\w+["']?\s+)?FOREIGN\s+KEY\b`)

// foreignKeyRegex matches FOREIGN KEY (column) REFERENCES [schema.]table
// (column) and captures the column, the target table without its schema, and
// the text after the reference, which holds the referential actions.
var foreignKeyRegex = regexp.MustCompile(`(?is)FOREIGN\s+KEY\s*\(\s*["']?(\w+)["']?\s*\)\s*REFERENCES\s+(?:["']?\w+["']?\s*\.\s*)?["']?(\w+)["']?\s*\(\s*["']?\w+["']?\s*\)([^;]*)`)

// referentialActionRegex matches an ON DELETE or ON UPDATE clause, which may
// come in either order.
var referentialActionRegex = regexp.MustCompile(`(?i)\bON\s+(DELETE|UPDATE)\s+(SET\s+NULL|SET\s+DEFAULT|NO\s+ACTION|\w+)`)

// tableCheckUniqueRegex matches a named CHECK or UNIQUE table constraint, as
// SQLite migrations declare them inside CREATE TABLE.
var tableCheckUniqueRegex = regexp.MustCompile(`(?is)^CONSTRAINT\s+["']?(\w+)["']?\s+(CHECK|UNIQUE)\s*\((.*)\)$`)

// ParseCreateTable parses a CREATE TABLE statement and returns a CreateTable change
func (p *TableParser) ParseCreateTable(sql string) (*core.CreateTable, error) {
	createTable, _, err := p.ParseCreateTableWithForeignKeys(sql)
	return createTable, err
}

// ParseCreateTableWithForeignKeys parses a CREATE TABLE statement into the
// table and the foreign keys it declares as table constraints, which is how
// SQLite migrations add them.
func (p *TableParser) ParseCreateTableWithForeignKeys(sql string) (*core.CreateTable, []*core.AddForeignKey, error) {
	// Comments inside the column list must not split or end it.
	sql = stripComments(sql)
	matches := p.createTableRegex.FindStringSubmatchIndex(sql)
	if len(matches) < 4 {
		return nil, nil, fmt.Errorf("could not parse CREATE TABLE statement")
	}

	tableName := sql[matches[2]:matches[3]]
	columnsDef, err := extractColumnDefinitions(sql, matches[1])
	if err != nil {
		return nil, nil, err
	}

	// Parse columns
	var fields []generator.FieldDefinition
	var foreignKeys []*core.AddForeignKey
	var constraints []generator.ConstraintDefinition
	columnParts := splitColumnDefinitions(columnsDef)

	for _, colDef := range columnParts {
		if trimmed := strings.TrimSpace(colDef); tableConstraintRegex.MatchString(trimmed) {
			if fk := parseTableForeignKey(tableName, trimmed); fk != nil {
				foreignKeys = append(foreignKeys, fk)
			} else if constraint, ok := parseTableCheckUnique(trimmed); ok {
				constraints = append(constraints, constraint)
			}
			continue
		}
		field, err := p.parseColumnDefinition(colDef)
		if err != nil {
			// Skip columns that can't be parsed
			continue
		}
		fields = append(fields, field)
	}

	// Create model definition
	def := &generator.ModelDefinition{
		Name:   toPascalCase(tableName),
		Fields: fields,
		Meta: generator.MetaDefinition{
			TableName:   tableName,
			Constraints: constraints,
		},
	}

	return &core.CreateTable{Table: def}, foreignKeys, nil
}

// parseTableForeignKey parses a table-level FOREIGN KEY constraint.
func parseTableForeignKey(tableName, constraint string) *core.AddForeignKey {
	if !tableForeignKeyRegex.MatchString(constraint) {
		return nil
	}
	return parseForeignKey(tableName, constraint)
}

// parseForeignKey parses the first FOREIGN KEY .. REFERENCES clause in sql as
// a foreign key of tableName.
func parseForeignKey(tableName, sql string) *core.AddForeignKey {
	matches := foreignKeyRegex.FindStringSubmatch(sql)
	if matches == nil {
		return nil
	}
	onDelete, onUpdate := "NO ACTION", "NO ACTION"
	for _, action := range referentialActionRegex.FindAllStringSubmatch(matches[3], -1) {
		if strings.EqualFold(action[1], "DELETE") {
			onDelete = normalizeCascadeAction(action[2])
		} else {
			onUpdate = normalizeCascadeAction(action[2])
		}
	}
	return &core.AddForeignKey{
		Table: tableName,
		Relation: generator.RelationDefinition{
			Name: matches[1],
			Options: map[string]interface{}{
				"on_delete": denormalizeCascadeAction(onDelete),
				"on_update": denormalizeCascadeAction(onUpdate),
			},
		},
		TargetTable: matches[2],
	}
}

// parseTableCheckUnique parses a named CHECK or UNIQUE table constraint.
func parseTableCheckUnique(constraint string) (generator.ConstraintDefinition, bool) {
	matches := tableCheckUniqueRegex.FindStringSubmatch(constraint)
	if matches == nil {
		return generator.ConstraintDefinition{}, false
	}
	return constraintDefinition(matches[1], matches[2], matches[3]), true
}

// constraintDefinition builds a CHECK or UNIQUE constraint from its name, type
// and the text inside its parentheses.
func constraintDefinition(name, constraintType, body string) generator.ConstraintDefinition {
	def := generator.ConstraintDefinition{Name: name, Type: strings.ToUpper(constraintType)}
	body = strings.TrimSpace(body)
	if def.Type == "UNIQUE" {
		def.Fields = extractIndexFieldsFromString(body)
	} else {
		def.Condition = body
	}
	return def
}

func extractColumnDefinitions(sql string, searchStart int) (string, error) {
	openIdx := strings.Index(sql[searchStart:], "(")
	if openIdx == -1 {
		return "", fmt.Errorf("could not locate column definitions")
	}
	openIdx += searchStart

	depth := 0
	for i := openIdx; i < len(sql); i++ {
		switch sql[i] {
		case '\'', '"':
			i = skipQuoted(sql, i) - 1
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return sql[openIdx+1 : i], nil
			}
		}
	}
	return "", fmt.Errorf("could not parse column definitions")
}

// parseColumnDefinition parses a single column definition
func (p *TableParser) parseColumnDefinition(colDef string) (generator.FieldDefinition, error) {
	colDef = strings.TrimSpace(colDef)

	// Extract column name (first word, possibly quoted)
	nameRegex := regexp.MustCompile(`^["']?(\w+)["']?`)
	nameMatch := nameRegex.FindStringSubmatch(colDef)
	if len(nameMatch) < 2 {
		return generator.FieldDefinition{}, fmt.Errorf("could not parse column name")
	}

	columnName := nameMatch[1]
	remaining := colDef[len(nameMatch[0]):]

	// Extract SQL type
	typeMatch := columnTypeRegex.FindStringSubmatch(remaining)
	if len(typeMatch) < 2 {
		return generator.FieldDefinition{}, fmt.Errorf("could not parse column type")
	}

	sqlType := typeMatch[1]
	remaining = remaining[len(typeMatch[0]):]

	field := fieldForSQLType(columnName, sqlType)
	// Keywords are matched outside quoted literals, so a default such as
	// 'NOT NULL' does not set them.
	masked := maskQuoted(remaining)

	// Check for PRIMARY KEY
	if strings.Contains(strings.ToUpper(masked), "PRIMARY KEY") {
		field.PrimaryKey = true
		field.Required = true
	}

	// Check for AUTOINCREMENT / AUTO_INCREMENT / GENERATED ALWAYS AS IDENTITY
	if strings.Contains(strings.ToUpper(masked), "AUTOINCREMENT") ||
		strings.Contains(strings.ToUpper(masked), "AUTO_INCREMENT") ||
		strings.Contains(strings.ToUpper(masked), "GENERATED ALWAYS AS IDENTITY") {
		field.AutoIncrement = true
	}

	// Check for GENERATED ALWAYS AS (expr)
	upperRemaining := strings.ToUpper(masked)
	if strings.Contains(upperRemaining, "GENERATED ALWAYS AS") && !strings.Contains(upperRemaining, "IDENTITY") {
		field.Options["generated"] = true
		startIdx := strings.Index(remaining, "(")
		endIdx := strings.LastIndex(remaining, ")")
		if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
			field.Options["generated_expr"] = strings.TrimSpace(remaining[startIdx+1 : endIdx])
		}
		if strings.Contains(upperRemaining, "STORED") {
			field.Options["generated_stored"] = true
		}
	}

	// Check for NOT NULL
	if strings.Contains(strings.ToUpper(masked), "NOT NULL") {
		field.Required = true
	}

	// Check for UNIQUE
	if strings.Contains(strings.ToUpper(masked), "UNIQUE") {
		field.Options["unique"] = true
	}

	// Check for DEFAULT
	if expr, ok := defaultExpression(remaining); ok {
		field.Default = parseDefaultValue(expr)
	}

	return field, nil
}

// fieldForSQLType returns a field of the given SQL type, with the options its
// size or precision implies, as for VARCHAR(255) or NUMERIC(10, 2).
func fieldForSQLType(name, sqlType string) generator.FieldDefinition {
	field := generator.FieldDefinition{
		Name:    name,
		Type:    mapSQLTypeToFieldType(sqlType),
		GoType:  mapSQLTypeToGoType(sqlType),
		Options: map[string]interface{}{SQLTypeOption: NormalizeSQLType(sqlType)},
	}

	// Extract options from sqlType like VARCHAR(255), NUMERIC(10, 2)
	upperSQLType := strings.ToUpper(sqlType)
	if openIdx := strings.Index(upperSQLType, "("); openIdx > 0 {
		if closeIdx := strings.Index(upperSQLType, ")"); closeIdx > openIdx {
			inner := upperSQLType[openIdx+1 : closeIdx]
			if strings.HasPrefix(upperSQLType, "VARCHAR") || strings.HasPrefix(upperSQLType, "CHAR") {
				if maxLen, err := strconv.Atoi(strings.TrimSpace(inner)); err == nil {
					field.Options["max_length"] = maxLen
				}
			} else if strings.HasPrefix(upperSQLType, "NUMERIC") || strings.HasPrefix(upperSQLType, "DECIMAL") {
				parts := strings.Split(inner, ",")
				if len(parts) >= 1 {
					if maxDigits, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
						field.Options["max_digits"] = maxDigits
					}
				}
				if len(parts) >= 2 {
					if decimalPlaces, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						field.Options["decimal_places"] = decimalPlaces
					}
				}
			}
		}
	}
	return field
}

// splitColumnDefinitions splits column definitions from CREATE TABLE
func splitColumnDefinitions(defs string) []string {
	var columns []string
	start := 0
	parenDepth := 0

	for i := 0; i < len(defs); i++ {
		switch defs[i] {
		case '\'', '"':
			i = skipQuoted(defs, i) - 1
		case '(':
			parenDepth++
		case ')':
			parenDepth--
		case ',':
			if parenDepth == 0 {
				if col := strings.TrimSpace(defs[start:i]); col != "" {
					columns = append(columns, col)
				}
				start = i + 1
			}
		}
	}

	// Add last column
	col := strings.TrimSpace(defs[start:])
	if col != "" {
		columns = append(columns, col)
	}

	return columns
}

// SQLTypeOption is the field option under which the parser records a column's
// SQL type as a migration file declared it (see core.SQLTypeOption).
const SQLTypeOption = core.SQLTypeOption

// columnTypeRegex matches a column type, including the multi-word types the
// SQL builder emits (DOUBLE PRECISION, TIMESTAMP WITH TIME ZONE) and
// PostgreSQL array suffixes (TEXT[], INTEGER[3][]).
var columnTypeRegex = regexp.MustCompile(`(?i)^\s+((?:DOUBLE\s+PRECISION|CHARACTER\s+VARYING|(?:TIMESTAMP|TIME)\s+WITH(?:OUT)?\s+TIME\s+ZONE|\w+)(?:\s*\([^)]+\))?(?:\s*\[\s*\d*\s*\])*)`)

// NormalizeSQLType upper-cases a SQL type and collapses its whitespace so that
// equivalent spellings compare equal.
func NormalizeSQLType(sqlType string) string {
	normalized := strings.Join(strings.Fields(strings.ToUpper(sqlType)), " ")
	normalized = strings.ReplaceAll(normalized, " (", "(")
	normalized = strings.ReplaceAll(normalized, " [", "[")
	return strings.ReplaceAll(normalized, ", ", ",")
}

// mapSQLTypeToFieldType maps SQL types to field types
func mapSQLTypeToFieldType(sqlType string) string {
	sqlType = NormalizeSQLType(sqlType)

	// Remove size/precision
	if idx := strings.Index(sqlType, "("); idx > 0 {
		sqlType = sqlType[:idx]
	}

	switch sqlType {
	case "BIGINT", "INTEGER", "INT":
		return "Int64"
	case "SMALLINT":
		return "Int32"
	case "TEXT", "VARCHAR", "CHAR", "CHARACTER VARYING":
		return "String"
	case "BOOLEAN", "BOOL":
		return "Bool"
	case "REAL", "DOUBLE PRECISION", "FLOAT":
		return "Float64"
	case "NUMERIC", "DECIMAL":
		return "Decimal"
	case "DATE":
		return "Date"
	case "TIMESTAMP", "TIMESTAMP WITH TIME ZONE", "TIMESTAMP WITHOUT TIME ZONE":
		return "DateTime"
	case "TIME", "TIME WITH TIME ZONE", "TIME WITHOUT TIME ZONE":
		return "Time"
	case "UUID":
		return "UUID"
	case "JSON", "JSONB":
		return "JSON"
	case "BYTEA", "BLOB":
		return "Bytes"
	default:
		return "String"
	}
}

// mapSQLTypeToGoType maps SQL types to Go types
func mapSQLTypeToGoType(sqlType string) string {
	fieldType := mapSQLTypeToFieldType(sqlType)

	switch fieldType {
	case "Int64", "Int":
		return "int64"
	case "Int32":
		return "int32"
	case "String", "Text", "Email", "URL":
		return "string"
	case "Bool":
		return "bool"
	case "Float64":
		return "float64"
	case "Decimal":
		return "float64"
	case "Date", "DateTime", "Time":
		return "time.Time"
	case "UUID":
		return "string"
	case "JSON":
		return "map[string]interface{}"
	case "Bytes":
		return "[]byte"
	default:
		return "string"
	}
}

// parseDefaultValue parses a default value from SQL
func parseDefaultValue(value string) interface{} {
	value = strings.TrimSpace(value)
	// A string literal is a string, whatever it spells.
	if literal, ok := unquoteLiteral(value); ok {
		return literal
	}
	value = strings.Trim(value, `"'`)

	// Check for SQL functions
	if strings.Contains(value, "(") {
		return value // Return as-is for functions like now()
	}

	// Try to parse as number
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal
	}
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Check for boolean
	if strings.EqualFold(value, "true") {
		return true
	}
	if strings.EqualFold(value, "false") {
		return false
	}

	// Return as string
	return value
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]) + strings.ToLower(part[1:]))
		}
	}
	return result.String()
}
