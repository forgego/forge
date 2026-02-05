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
		createTableRegex: regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?\s*\(([^)]+)\)`),
		columnRegex:      regexp.MustCompile(`["']?(\w+)["']?\s+(\w+(?:\([^)]+\))?)\s*(.*?)(?:,|$)`),
	}
}

// ParseCreateTable parses a CREATE TABLE statement and returns a CreateTable change
func (p *TableParser) ParseCreateTable(sql string) (*core.CreateTable, error) {
	matches := p.createTableRegex.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return nil, fmt.Errorf("could not parse CREATE TABLE statement")
	}

	tableName := matches[1]
	columnsDef := matches[2]

	// Parse columns
	var fields []generator.FieldDefinition
	columnParts := splitColumnDefinitions(columnsDef)

	for _, colDef := range columnParts {
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

	return &core.CreateTable{Table: def}, nil
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
	typeRegex := regexp.MustCompile(`^\s+(\w+(?:\([^)]+\))?)`)
	typeMatch := typeRegex.FindStringSubmatch(remaining)
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
		Options: make(map[string]interface{}),
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

// mapSQLTypeToFieldType maps SQL types to field types
func mapSQLTypeToFieldType(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)

	// Remove size/precision
	if idx := strings.Index(sqlType, "("); idx > 0 {
		sqlType = sqlType[:idx]
	}

	switch sqlType {
	case "BIGINT", "INTEGER", "INT":
		return "Int64"
	case "SMALLINT":
		return "Int32"
	case "TEXT", "VARCHAR", "CHAR":
		return "String"
	case "BOOLEAN", "BOOL":
		return "Bool"
	case "REAL", "DOUBLE PRECISION", "FLOAT":
		return "Float64"
	case "NUMERIC", "DECIMAL":
		return "Decimal"
	case "DATE":
		return "Date"
	case "TIMESTAMP", "TIMESTAMP WITH TIME ZONE":
		return "DateTime"
	case "TIME":
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
