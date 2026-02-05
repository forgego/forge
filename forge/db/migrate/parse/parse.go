package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/forgego/forge/db/migrate/core"
)

// ParseError represents a parsing error with context
type ParseError struct {
	File    string
	Line    int
	Column  int
	SQL     string
	Message string
}

// Error implements error interface
func (e *ParseError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.File, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%d:%d: %s", e.Line, e.Column, e.Message)
}

// SQLParser parses SQL migration files to extract schema changes
// Uses lexer-based statement splitting and limited DDL parsing
type SQLParser struct {
	lexer      *Lexer
	classifier *Classifier
	ddlParser  *DDLParser
	verbose    bool
	errors     []*ParseError
	tableCtx   string // Current table context for index parsing
}

// ParserOptions configures parser behavior
type ParserOptions struct {
	Verbose bool // Enable verbose logging of UnknownChange statements
}

// NewSQLParser creates a new SQL parser using lexer-based approach
func NewSQLParser() *SQLParser {
	return NewSQLParserWithOptions(ParserOptions{Verbose: false})
}

// NewSQLParserWithOptions creates a new SQL parser with options
func NewSQLParserWithOptions(opts ParserOptions) *SQLParser {
	return &SQLParser{
		lexer:      NewLexer(),
		classifier: NewClassifier(),
		ddlParser:  NewDDLParserWithOptions(DDLParserOptions{Verbose: opts.Verbose}),
		verbose:    opts.Verbose,
		errors:     []*ParseError{},
	}
}

// GetErrors returns all parse errors collected during parsing
func (p *SQLParser) GetErrors() []*ParseError {
	return p.errors
}

// ClearErrors clears collected parse errors
func (p *SQLParser) ClearErrors() {
	p.errors = nil
}

// ParseUpSQL parses up migration SQL and extracts changes
// Uses lexer-based statement splitting and limited DDL parsing
// Returns UnknownChange for unparseable statements (fail-soft)
// Uses two-pass parsing: first CREATE TABLE, then constraints/indexes
func (p *SQLParser) ParseUpSQL(sql string) ([]core.Change, error) {
	// 1. Split SQL into statements using lexer
	statements, err := p.lexer.Scan(sql)
	if err != nil {
		// If lexer fails, try to continue with best-effort
		// This shouldn't happen often, but fail softly
		line, col := p.getLineColumn(sql, 0)
		if p.verbose {
			p.errors = append(p.errors, &ParseError{
				Line:    line,
				Column:  col,
				SQL:     sql,
				Message: fmt.Sprintf("lexer error: %v", err),
			})
		}
		return []core.Change{&core.UnknownChange{SQL: sql}}, nil
	}

	// 2. Two-pass parsing: first CREATE TABLE, then constraints/indexes
	// First pass: collect CREATE TABLE statements and track table context
	var createTableChanges []core.Change
	var otherChanges []core.Change
	var tableContexts = make(map[string]string) // Track table names for index parsing

	for _, stmt := range statements {
		// Skip empty statements
		if strings.TrimSpace(stmt.Text) == "" {
			continue
		}

		// Update table context for index parsing
		p.updateTableContext(stmt.Text, tableContexts)

		// Parse statement using DDL parser
		parsed, err := p.ddlParser.ParseStatement(stmt)
		if err != nil {
			// Fail softly: add as UnknownChange
			line, col := p.getLineColumn(sql, stmt.Pos)
			if p.verbose {
				p.errors = append(p.errors, &ParseError{
					Line:    line,
					Column:  col,
					SQL:     stmt.Text,
					Message: fmt.Sprintf("parse error: %v", err),
				})
			}
			changes := []core.Change{&core.UnknownChange{SQL: stmt.Text}}
			otherChanges = append(otherChanges, changes...)
			continue
		}

		// Separate CREATE TABLE from other changes
		for _, change := range parsed {
			if _, ok := change.(*core.CreateTable); ok {
				createTableChanges = append(createTableChanges, change)
			} else {
				otherChanges = append(otherChanges, change)
			}
		}
	}

	// Combine: CREATE TABLE first, then everything else
	changes := append(createTableChanges, otherChanges...)

	return changes, nil
}

// updateTableContext updates table context for index parsing
func (p *SQLParser) updateTableContext(sql string, contexts map[string]string) {
	upper := strings.ToUpper(sql)
	// Track CREATE TABLE statements
	if strings.Contains(upper, "CREATE TABLE") {
		re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?["']?(\w+)["']?`)
		matches := re.FindStringSubmatch(sql)
		if len(matches) >= 2 {
			tableName := matches[1]
			p.tableCtx = tableName
			contexts[tableName] = tableName
		}
	}
	// Track ALTER TABLE statements
	if strings.Contains(upper, "ALTER TABLE") {
		re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+["']?(\w+)["']?`)
		matches := re.FindStringSubmatch(sql)
		if len(matches) >= 2 {
			p.tableCtx = matches[1]
		}
	}
}

// getLineColumn calculates line and column from position in SQL string
func (p *SQLParser) getLineColumn(sql string, pos int) (line, col int) {
	if pos < 0 || pos > len(sql) {
		return 1, 1
	}
	line = 1
	col = 1
	for i := 0; i < pos && i < len(sql); i++ {
		if sql[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}
