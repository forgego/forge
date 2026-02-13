// Package dialect provides SQL dialect abstraction for database-agnostic query generation.
// It supports multiple database backends (PostgreSQL, SQLite) with proper placeholder
// formatting, identifier quoting, and dialect-specific SQL features.
package dialect

import (
	"fmt"
	"strings"
)

// Dialect defines the interface for SQL dialect-specific behavior.
// Each database backend implements this interface to provide proper SQL generation.
type Dialect interface {
	// Name returns the dialect name (e.g., "postgres", "sqlite").
	Name() string

	// Placeholder returns the placeholder for the given position (1-indexed).
	// PostgreSQL: $1, $2, $3...
	// SQLite/MySQL: ?, ?, ?
	Placeholder(position int) string

	// BuildPlaceholders generates n placeholders joined by ", ".
	// Example: BuildPlaceholders(3) returns "$1, $2, $3" for PostgreSQL
	// or "?, ?, ?" for SQLite.
	BuildPlaceholders(n int) string

	// QuoteIdentifier quotes a table or column name.
	// PostgreSQL: "table_name"
	// MySQL: `table_name`
	// SQLite: "table_name" or `table_name`
	QuoteIdentifier(name string) string

	// QuoteString quotes a string literal with proper escaping.
	QuoteString(s string) string

	// AutoIncrementType returns the auto-increment type declaration.
	// PostgreSQL: "GENERATED ALWAYS AS IDENTITY" or "SERIAL"
	// SQLite: "AUTOINCREMENT"
	// MySQL: "AUTO_INCREMENT"
	AutoIncrementType() string

	// SupportsReturning returns true if the dialect supports RETURNING clause.
	// PostgreSQL: true
	// SQLite: true (3.35+)
	// MySQL: false (uses LAST_INSERT_ID())
	SupportsReturning() bool

	// CurrentTime returns the current time function.
	// PostgreSQL: "NOW()" or "CURRENT_TIMESTAMP"
	// SQLite: "datetime('now')"
	// MySQL: "NOW()"
	CurrentTime() string

	// CurrentTimestamp returns the current timestamp function name.
	CurrentTimestamp() string

	// BooleanLiteral returns the boolean literal representation.
	// PostgreSQL: TRUE, FALSE
	// SQLite: 1, 0
	// MySQL: TRUE, FALSE
	BooleanLiteral(value bool) string

	// LimitOffset generates LIMIT and OFFSET clause.
	// Some dialects have specific requirements for limit/offset ordering.
	LimitOffset(limit, offset int) string

	// LikeEscape returns the escape character for LIKE patterns.
	// Most databases use backslash, but some may differ.
	LikeEscape() string

	// ConcatOperator returns the string concatenation operator or function.
	// PostgreSQL: "||" or CONCAT()
	// SQLite: "||"
	// MySQL: CONCAT()
	ConcatOperator() string

	// CreateTableOptions returns dialect-specific CREATE TABLE options.
	// Example: "WITHOUT ROWID" for SQLite
	CreateTableOptions() string

	// OnConflictDoNothing returns the dialect-specific ON CONFLICT DO NOTHING clause.
	// PostgreSQL: "ON CONFLICT DO NOTHING"
	// SQLite: "ON CONFLICT DO NOTHING" or "INSERT OR IGNORE"
	// MySQL: "INSERT IGNORE"
	OnConflictDoNothing() string

	// OnConflictDoUpdate returns the dialect-specific ON CONFLICT DO UPDATE clause.
	// PostgreSQL: "ON CONFLICT (column) DO UPDATE SET ..."
	// SQLite: "ON CONFLICT (column) DO UPDATE SET ..."
	OnConflictDoUpdate(column string, updates []string) string
}

// BaseDialect provides common functionality that can be embedded in specific dialects.
type BaseDialect struct {
	name              string
	placeholderStyle  PlaceholderStyle
	quoteChar         string
	supportsReturning bool
}

// PlaceholderStyle defines how placeholders are formatted.
type PlaceholderStyle int

const (
	// PositionalStyle uses positional placeholders like $1, $2, $3 (PostgreSQL)
	PositionalStyle PlaceholderStyle = iota
	// QuestionMarkStyle uses question mark placeholders like ?, ?, ? (SQLite, MySQL)
	QuestionMarkStyle
)

// NewBaseDialect creates a new base dialect with common settings.
func NewBaseDialect(name string, style PlaceholderStyle, quoteChar string, supportsReturning bool) BaseDialect {
	return BaseDialect{
		name:              name,
		placeholderStyle:  style,
		quoteChar:         quoteChar,
		supportsReturning: supportsReturning,
	}
}

// Name returns the dialect name.
func (d BaseDialect) Name() string {
	return d.name
}

// Placeholder returns the placeholder for the given position.
func (d BaseDialect) Placeholder(position int) string {
	switch d.placeholderStyle {
	case PositionalStyle:
		return fmt.Sprintf("$%d", position)
	case QuestionMarkStyle:
		return "?"
	default:
		return "?"
	}
}

// BuildPlaceholders generates n placeholders joined by ", ".
func (d BaseDialect) BuildPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = d.Placeholder(i + 1)
	}
	return strings.Join(parts, ", ")
}

// QuoteIdentifier quotes a table or column name.
func (d BaseDialect) QuoteIdentifier(name string) string {
	// Escape any existing quote characters
	escaped := strings.ReplaceAll(name, d.quoteChar, d.quoteChar+d.quoteChar)
	return d.quoteChar + escaped + d.quoteChar
}

// QuoteString quotes a string literal with proper escaping.
func (d BaseDialect) QuoteString(s string) string {
	// Standard SQL escaping - double single quotes
	escaped := strings.ReplaceAll(s, "'", "''")
	return "'" + escaped + "'"
}

// SupportsReturning returns whether the dialect supports RETURNING clause.
func (d BaseDialect) SupportsReturning() bool {
	return d.supportsReturning
}

// LikeEscape returns the escape character for LIKE patterns.
func (d BaseDialect) LikeEscape() string {
	return "\\"
}

// ConcatOperator returns the string concatenation operator.
func (d BaseDialect) ConcatOperator() string {
	return "||"
}

// CreateTableOptions returns dialect-specific CREATE TABLE options.
func (d BaseDialect) CreateTableOptions() string {
	return ""
}

// OnConflictDoNothing returns the dialect-specific ON CONFLICT DO NOTHING clause.
func (d BaseDialect) OnConflictDoNothing() string {
	return "ON CONFLICT DO NOTHING"
}

// OnConflictDoUpdate returns the dialect-specific ON CONFLICT DO UPDATE clause.
func (d BaseDialect) OnConflictDoUpdate(column string, updates []string) string {
	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", column, strings.Join(updates, ", "))
}

// DetectDialectFromDSN detects the appropriate dialect from a DSN string.
// It examines the DSN prefix and structure to determine the database type.
func DetectDialectFromDSN(dsn string) Dialect {
	dsn = strings.ToLower(dsn)

	// Check for explicit driver prefixes
	if strings.HasPrefix(dsn, "postgres://") ||
		strings.HasPrefix(dsn, "postgresql://") ||
		strings.Contains(dsn, "host=") ||
		strings.Contains(dsn, "sslmode=") {
		return NewPostgreSQLDialect()
	}

	// SQLite typically uses file paths
	if strings.HasPrefix(dsn, "file:") ||
		strings.HasSuffix(dsn, ".db") ||
		strings.HasSuffix(dsn, ".sqlite") ||
		strings.HasSuffix(dsn, ".sqlite3") ||
		dsn == ":memory:" {
		return NewSQLiteDialect()
	}

	// Default to PostgreSQL for unknown DSNs
	return NewPostgreSQLDialect()
}

// DetectDialectFromDriver detects the appropriate dialect from a driver name.
func DetectDialectFromDriver(driver string) Dialect {
	switch strings.ToLower(driver) {
	case "postgres", "postgresql", "pgx":
		return NewPostgreSQLDialect()
	case "sqlite", "sqlite3":
		return NewSQLiteDialect()
	default:
		// Default to PostgreSQL
		return NewPostgreSQLDialect()
	}
}
