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

// tableForeignKeyRegex matches a table-level foreign key, as SQLite migrations
// declare them inside CREATE TABLE.
var tableForeignKeyRegex = regexp.MustCompile(`(?i)^(?:CONSTRAINT\s+["']?\w+["']?\s+)?FOREIGN\s+KEY\s*\(\s*["']?(\w+)["']?\s*\)\s*REFERENCES\s+["']?(\w+)["']?\s*\(\s*["']?\w+["']?\s*\)(?:\s+ON\s+DELETE\s+(SET\s+NULL|SET\s+DEFAULT|NO\s+ACTION|\w+))?(?:\s+ON\s+UPDATE\s+(SET\s+NULL|SET\s+DEFAULT|NO\s+ACTION|\w+))?`)

// ParseCreateTable parses a CREATE TABLE statement and returns a CreateTable change
func (p *TableParser) ParseCreateTable(sql string) (*core.CreateTable, error) {
	createTable, _, err := p.ParseCreateTableWithForeignKeys(sql)
	return createTable, err
}

// ParseCreateTableWithForeignKeys parses a CREATE TABLE statement into the
// table and the foreign keys it declares as table constraints, which is how
// SQLite migrations add them.
func (p *TableParser) ParseCreateTableWithForeignKeys(sql string) (*core.CreateTable, []*core.AddForeignKey, error) {
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
	columnParts := splitColumnDefinitions(columnsDef)

	for _, colDef := range columnParts {
		if trimmed := strings.TrimSpace(colDef); tableConstraintRegex.MatchString(trimmed) {
			if fk := parseTableForeignKey(tableName, trimmed); fk != nil {
				foreignKeys = append(foreignKeys, fk)
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
			TableName: tableName,
		},
	}

	return &core.CreateTable{Table: def}, foreignKeys, nil
}

// parseTableForeignKey parses a table-level FOREIGN KEY constraint.
func parseTableForeignKey(tableName, constraint string) *core.AddForeignKey {
	matches := tableForeignKeyRegex.FindStringSubmatch(constraint)
	if matches == nil {
		return nil
	}
	onDelete, onUpdate := "NO ACTION", "NO ACTION"
	if matches[3] != "" {
		onDelete = normalizeCascadeAction(matches[3])
	}
	if matches[4] != "" {
		onUpdate = normalizeCascadeAction(matches[4])
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

func extractColumnDefinitions(sql string, searchStart int) (string, error) {
	openIdx := strings.Index(sql[searchStart:], "(")
	if openIdx == -1 {
		return "", fmt.Errorf("could not locate column definitions")
	}
	openIdx += searchStart

	depth := 0
	for i := openIdx; i < len(sql); i++ {
		switch sql[i] {
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

	// Parse column attributes
	field := generator.FieldDefinition{
		Name:    columnName,
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

	// Check for PRIMARY KEY
	if strings.Contains(strings.ToUpper(remaining), "PRIMARY KEY") {
		field.PrimaryKey = true
		field.Required = true
	}

	// Check for AUTOINCREMENT / AUTO_INCREMENT / GENERATED ALWAYS AS IDENTITY
	if strings.Contains(strings.ToUpper(remaining), "AUTOINCREMENT") ||
		strings.Contains(strings.ToUpper(remaining), "AUTO_INCREMENT") ||
		strings.Contains(strings.ToUpper(remaining), "GENERATED ALWAYS AS IDENTITY") {
		field.AutoIncrement = true
	}

	// Check for GENERATED ALWAYS AS (expr)
	upperRemaining := strings.ToUpper(remaining)
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
	if strings.Contains(strings.ToUpper(remaining), "NOT NULL") {
		field.Required = true
	}

	// Check for UNIQUE
	if strings.Contains(strings.ToUpper(remaining), "UNIQUE") {
		field.Options["unique"] = true
	}

	// Check for DEFAULT
	defaultRegex := regexp.MustCompile(`(?i)DEFAULT\s+([^\s,]+)`)
	if defaultMatch := defaultRegex.FindStringSubmatch(remaining); len(defaultMatch) > 1 {
		field.Default = parseDefaultValue(defaultMatch[1])
	}

	return field, nil
}

// splitColumnDefinitions splits column definitions from CREATE TABLE
func splitColumnDefinitions(defs string) []string {
	var columns []string
	var current strings.Builder
	parenDepth := 0

	for _, char := range defs {
		switch char {
		case '(':
			parenDepth++
			current.WriteRune(char)
		case ')':
			parenDepth--
			current.WriteRune(char)
		case ',':
			if parenDepth == 0 {
				col := strings.TrimSpace(current.String())
				if col != "" {
					columns = append(columns, col)
				}
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add last column
	col := strings.TrimSpace(current.String())
	if col != "" {
		columns = append(columns, col)
	}

	return columns
}

// SQLTypeOption is the field option under which the parser records a column's
// SQL type as a migration file declared it (see core.SQLTypeOption).
const SQLTypeOption = core.SQLTypeOption

// columnTypeRegex matches a column type, including the multi-word types the
// SQL builder emits (DOUBLE PRECISION, TIMESTAMP WITH TIME ZONE).
var columnTypeRegex = regexp.MustCompile(`(?i)^\s+((?:DOUBLE\s+PRECISION|CHARACTER\s+VARYING|(?:TIMESTAMP|TIME)\s+WITH(?:OUT)?\s+TIME\s+ZONE|\w+)(?:\s*\([^)]+\))?)`)

// NormalizeSQLType upper-cases a SQL type and collapses its whitespace so that
// equivalent spellings compare equal.
func NormalizeSQLType(sqlType string) string {
	normalized := strings.Join(strings.Fields(strings.ToUpper(sqlType)), " ")
	normalized = strings.ReplaceAll(normalized, " (", "(")
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
