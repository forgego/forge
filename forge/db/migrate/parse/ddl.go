package parse

import (
	"regexp"
	"strings"

	"github.com/forgego/forge/codegen"
	"github.com/forgego/forge/db/migrate/core"
)

// DDLParserOptions configures DDL parser behavior
type DDLParserOptions struct {
	Verbose bool // Enable verbose logging
}

// DDLParser parses DDL statements into core.Change objects
// Uses best-effort parsing with soft failure (returns UnknownChange for unparseable statements)
type DDLParser struct {
	dialect     core.Driver
	classifier  *Classifier
	tableParser *TableParser
	verbose     bool
	tableCtx    string // Current table context for index parsing
}

// NewDDLParser creates a new DDL parser
func NewDDLParser() *DDLParser {
	return NewDDLParserWithDialect(core.DriverPostgreSQL)
}

// NewDDLParserWithDialect creates a new DDL parser with a specific dialect
func NewDDLParserWithDialect(dialect core.Driver) *DDLParser {
	return NewDDLParserWithOptions(DDLParserOptions{Verbose: false})
}

// NewDDLParserWithOptions creates a new DDL parser with options
func NewDDLParserWithOptions(opts DDLParserOptions) *DDLParser {
	return &DDLParser{
		dialect:     core.DriverPostgreSQL,
		classifier:  NewClassifier(),
		tableParser: NewTableParser(),
		verbose:     opts.Verbose,
	}
}

// SetTableContext sets the current table context for index parsing
func (p *DDLParser) SetTableContext(tableName string) {
	p.tableCtx = tableName
}

// ParseStatement parses a single SQL statement into core.Change objects
// Returns UnknownChange for unparseable DDL statements (fail-soft)
func (p *DDLParser) ParseStatement(stmt *Statement) ([]core.Change, error) {
	kind := p.classifier.Classify(stmt.Text)

	switch kind {
	case StmtNonDDL:
		// Skip non-DDL statements silently
		return nil, nil
	case StmtCreateTable:
		return p.parseCreateTable(stmt.Text)
	case StmtAlterTable:
		return p.parseAlterTable(stmt.Text)
	case StmtCreateIndex:
		return p.parseCreateIndex(stmt.Text)
	case StmtDropTable:
		return p.parseDropTable(stmt.Text)
	case StmtDropIndex:
		return p.parseDropIndex(stmt.Text)
	case StmtDropColumn:
		return p.parseDropColumn(stmt.Text)
	case StmtUnknown:
		// Unknown statement - might be DDL we don't recognize
		// Fail softly: return as UnknownChange
		return []core.Change{&core.UnknownChange{SQL: stmt.Text}}, nil
	default:
		return []core.Change{&core.UnknownChange{SQL: stmt.Text}}, nil
	}
}

// parseCreateTable parses CREATE TABLE statements
func (p *DDLParser) parseCreateTable(sql string) ([]core.Change, error) {
	// Extract table name and set context for index parsing
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		p.SetTableContext(matches[1])
	}

	createTable, err := p.tableParser.ParseCreateTable(sql)
	if err != nil {
		// Fail softly: return as UnknownChange
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}
	return []core.Change{createTable}, nil
}

// parseAlterTable parses ALTER TABLE statements
func (p *DDLParser) parseAlterTable(sql string) ([]core.Change, error) {
	upper := strings.ToUpper(sql)
	var changes []core.Change

	// Extract table name and set context
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) >= 2 {
		p.SetTableContext(matches[1])
	}

	// ALTER TABLE ADD COLUMN
	// Check for "ADD COLUMN" specifically to avoid matching "ADD CONSTRAINT"
	if strings.Contains(upper, "ADD COLUMN") {
		change := p.parseAddColumn(sql)
		if change != nil {
			changes = append(changes, change)
		}
	}

	// ALTER TABLE ADD CONSTRAINT FOREIGN KEY
	if strings.Contains(upper, "ADD CONSTRAINT") && strings.Contains(upper, "FOREIGN KEY") {
		change := p.parseAddForeignKey(sql)
		if change != nil {
			changes = append(changes, change)
		}
	}

	// ALTER TABLE CREATE INDEX (indexes created within ALTER TABLE)
	if strings.Contains(upper, "CREATE") && strings.Contains(upper, "INDEX") {
		// Extract index creation from ALTER TABLE
		indexRe := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?\s*\(([^)]+)\)`)
		indexMatches := indexRe.FindStringSubmatch(sql)
		if len(indexMatches) >= 3 && p.tableCtx != "" {
			indexName := indexMatches[1]
			fieldsStr := indexMatches[2]
			isUnique := strings.Contains(upper, "UNIQUE")
			fields := extractIndexFieldsFromString(fieldsStr)
			changes = append(changes, &core.AddIndex{
				Table: p.tableCtx,
				Index: generator.IndexDefinition{
					Name:   indexName,
					Fields: fields,
					Unique: isUnique,
				},
			})
		}
	}

	// ALTER TABLE DROP CONSTRAINT
	if strings.Contains(upper, "DROP CONSTRAINT") {
		change := p.parseDropConstraint(sql)
		if change != nil {
			changes = append(changes, change)
		}
	}

	// If no changes were parsed, return as UnknownChange
	if len(changes) == 0 {
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	return changes, nil
}

// parseAddColumn parses ALTER TABLE ADD COLUMN statements
func (p *DDLParser) parseAddColumn(sql string) *core.AddColumn {
	// Pattern: ALTER TABLE table_name ADD [COLUMN] column_name type [constraints]
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?\s+ADD\s+(?:COLUMN\s+)?["']?(\w+)["']?\s+(\w+(?:\([^)]+\))?)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 4 {
		return nil
	}

	tableName := matches[1]
	columnName := matches[2]
	typeStr := matches[3] // Keep type as opaque string

	// Parse basic constraints from remaining SQL
	remaining := sql[len(matches[0]):]
	required := strings.Contains(strings.ToUpper(remaining), "NOT NULL")

	field := generator.FieldDefinition{
		Name:     columnName,
		Type:     mapSQLTypeToFieldType(typeStr),
		GoType:   mapSQLTypeToGoType(typeStr),
		Required: required,
		Options:  make(map[string]interface{}),
	}

	// Check for UNIQUE
	if strings.Contains(strings.ToUpper(remaining), "UNIQUE") {
		field.Options["unique"] = true
	}

	// Check for DEFAULT (simplified - just extract value)
	defaultRe := regexp.MustCompile(`(?i)DEFAULT\s+([^\s,;]+)`)
	if defaultMatch := defaultRe.FindStringSubmatch(remaining); len(defaultMatch) > 1 {
		field.Default = parseDefaultValue(defaultMatch[1])
	}

	return &core.AddColumn{
		Table:  tableName,
		Column: field,
	}
}

// parseAddForeignKey parses ALTER TABLE ADD CONSTRAINT FOREIGN KEY statements
func (p *DDLParser) parseAddForeignKey(sql string) *core.AddForeignKey {
	// Pattern: ALTER TABLE table ADD CONSTRAINT name FOREIGN KEY (column) REFERENCES target (id) [ON DELETE action] [ON UPDATE action]
	// Also handles DO $$ BEGIN ... ALTER TABLE ... END $$; blocks
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?\s+ADD\s+CONSTRAINT\s+["']?(\w+)["']?\s+FOREIGN\s+KEY\s+\(["']?(\w+)["']?\)\s+REFERENCES\s+["']?(\w+)["']?\s+\(["']?(\w+)["']?\)(?:\s+ON\s+DELETE\s+(\w+))?(?:\s+ON\s+UPDATE\s+(\w+))?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil
	}

	tableName := matches[1]
	columnName := matches[3]
	targetTable := matches[4]

	onDelete := "NO ACTION"
	if len(matches) > 6 && matches[6] != "" {
		onDelete = normalizeCascadeAction(matches[6])
	}

	onUpdate := "NO ACTION"
	if len(matches) > 7 && matches[7] != "" {
		onUpdate = normalizeCascadeAction(matches[7])
	}

	relation := generator.RelationDefinition{
		Name: columnName,
		Options: map[string]interface{}{
			"on_delete": denormalizeCascadeAction(onDelete),
			"on_update": denormalizeCascadeAction(onUpdate),
		},
	}

	return &core.AddForeignKey{
		Table:       tableName,
		Relation:    relation,
		TargetTable: targetTable,
	}
}

// parseDropConstraint parses ALTER TABLE DROP CONSTRAINT statements
func (p *DDLParser) parseDropConstraint(sql string) core.Change {
	// Pattern: ALTER TABLE table DROP CONSTRAINT [IF EXISTS] constraint_name
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?\s+DROP\s+CONSTRAINT\s+(?:IF\s+EXISTS\s+)?["']?(\w+)["']?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return &core.UnknownChange{SQL: sql}
	}

	tableName := matches[1]
	constraintName := matches[2]

	// Check if it's a foreign key constraint
	upper := strings.ToUpper(sql)
	if strings.Contains(upper, "FOREIGN KEY") || strings.HasPrefix(constraintName, "fk_") {
		return &core.DropForeignKey{
			Table:  tableName,
			FKName: constraintName,
		}
	}

	return &core.DropConstraint{
		Table:          tableName,
		ConstraintName: constraintName,
	}
}

// parseCreateIndex parses CREATE INDEX statements
func (p *DDLParser) parseCreateIndex(sql string) ([]core.Change, error) {
	// Pattern: CREATE [UNIQUE] INDEX [IF NOT EXISTS] index_name ON table_name (columns)
	re := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?\s+ON\s+["']?(\w+)["']?\s*\(([^)]+)\)`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 4 {
		// Try to infer table name from context if available
		if p.tableCtx != "" {
			// Pattern: CREATE [UNIQUE] INDEX [IF NOT EXISTS] index_name (columns)
			re2 := regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?\s*\(([^)]+)\)`)
			matches2 := re2.FindStringSubmatch(sql)
			if len(matches2) >= 3 {
				indexName := matches2[1]
				fieldsStr := matches2[2]
				isUnique := strings.Contains(strings.ToUpper(sql), "UNIQUE")
				fields := extractIndexFieldsFromString(fieldsStr)
				return []core.Change{&core.AddIndex{
					Table: p.tableCtx,
					Index: generator.IndexDefinition{
						Name:   indexName,
						Fields: fields,
						Unique: isUnique,
					},
				}}, nil
			}
		}
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	indexName := matches[1]
	tableName := matches[2]
	fieldsStr := matches[3]
	isUnique := strings.Contains(strings.ToUpper(sql), "UNIQUE")

	// Extract field names
	fields := extractIndexFieldsFromString(fieldsStr)

	return []core.Change{&core.AddIndex{
		Table: tableName,
		Index: generator.IndexDefinition{
			Name:   indexName,
			Fields: fields,
			Unique: isUnique,
		},
	}}, nil
}

// parseDropTable parses DROP TABLE statements
func (p *DDLParser) parseDropTable(sql string) ([]core.Change, error) {
	// Pattern: DROP TABLE [IF EXISTS] table_name [CASCADE]
	re := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?["']?(\w+)["']?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	tableName := matches[1]
	return []core.Change{&core.DropTable{Table: tableName}}, nil
}

// parseDropIndex parses DROP INDEX statements
func (p *DDLParser) parseDropIndex(sql string) ([]core.Change, error) {
	// Pattern: DROP INDEX [IF EXISTS] index_name [ON table_name] (PostgreSQL)
	// Pattern: DROP INDEX [IF EXISTS] table_name.index_name (SQLite)
	re := regexp.MustCompile(`(?i)DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?(?:["']?(\w+)["']?\.)?["']?(\w+)["']?(?:\s+ON\s+["']?(\w+)["']?)?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 2 {
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	indexName := matches[2]
	tableName := ""
	
	// Try to get table name from various patterns
	if len(matches) > 3 && matches[3] != "" {
		// ON table_name pattern (PostgreSQL)
		tableName = matches[3]
	} else if len(matches) > 1 && matches[1] != "" {
		// table_name.index_name pattern (SQLite)
		tableName = matches[1]
	} else if p.tableCtx != "" {
		// Use context if available
		tableName = p.tableCtx
	}

	return []core.Change{&core.DropIndex{
		Table:     tableName,
		IndexName: indexName,
	}}, nil
}

// parseDropColumn parses ALTER TABLE DROP COLUMN statements
func (p *DDLParser) parseDropColumn(sql string) ([]core.Change, error) {
	// Pattern: ALTER TABLE table_name DROP COLUMN [IF EXISTS] column_name
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?\s+DROP\s+COLUMN\s+(?:IF\s+EXISTS\s+)?["']?(\w+)["']?`)
	matches := re.FindStringSubmatch(sql)
	if len(matches) < 3 {
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	tableName := matches[1]
	columnName := matches[2]
	return []core.Change{&core.DropColumn{
		Table:      tableName,
		ColumnName: columnName,
	}}, nil
}

// extractIndexFieldsFromString extracts field names from index definition string
func extractIndexFieldsFromString(fieldsStr string) []string {
	fields := strings.Split(fieldsStr, ",")
	var result []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		// Remove quotes and direction (ASC/DESC)
		field = strings.Trim(field, `"'`)
		// Remove direction keywords
		field = strings.TrimSuffix(strings.TrimSuffix(field, " ASC"), " DESC")
		field = strings.TrimSpace(field)
		if field != "" {
			result = append(result, field)
		}
	}
	return result
}

// normalizeCascadeAction normalizes cascade action strings to SQL format
func normalizeCascadeAction(action string) string {
	action = strings.ToUpper(strings.TrimSpace(action))
	switch action {
	case "CASCADE", "RESTRICT", "SET NULL", "NO ACTION":
		return action
	case "PROTECT":
		return "RESTRICT"
	default:
		return "NO ACTION"
	}
}

// denormalizeCascadeAction converts SQL cascade action back to model format
func denormalizeCascadeAction(action string) string {
	action = strings.ToUpper(strings.TrimSpace(action))
	switch action {
	case "RESTRICT":
		return "PROTECT"
	case "CASCADE", "SET NULL", "NO ACTION":
		return action
	default:
		return "NO ACTION"
	}
}
