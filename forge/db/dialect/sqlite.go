package dialect

import (
	"fmt"
	"strings"
)

// SQLiteDialect implements Dialect for SQLite databases.
type SQLiteDialect struct {
	BaseDialect
}

// NewSQLiteDialect creates a new SQLite dialect instance.
func NewSQLiteDialect() *SQLiteDialect {
	return &SQLiteDialect{
		BaseDialect: NewBaseDialect("sqlite", QuestionMarkStyle, `"`, true),
	}
}

// AutoIncrementType returns the auto-increment type declaration for SQLite.
// SQLite uses "AUTOINCREMENT" keyword for auto-incrementing primary keys.
func (d *SQLiteDialect) AutoIncrementType() string {
	return "AUTOINCREMENT"
}

// CurrentTime returns the current time function for SQLite.
// SQLite uses datetime('now') for current timestamp.
func (d *SQLiteDialect) CurrentTime() string {
	return "datetime('now')"
}

// CurrentTimestamp returns the current timestamp function for SQLite.
func (d *SQLiteDialect) CurrentTimestamp() string {
	return "datetime('now')"
}

// BooleanLiteral returns the boolean literal representation for SQLite.
// SQLite stores booleans as integers (0 and 1).
func (d *SQLiteDialect) BooleanLiteral(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// LimitOffset generates LIMIT and OFFSET clause for SQLite.
// SQLite supports standard LIMIT/OFFSET syntax.
func (d *SQLiteDialect) LimitOffset(limit, offset int) string {
	if limit <= 0 && offset <= 0 {
		return ""
	}
	var parts []string
	if limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", limit))
	}
	if offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", offset))
	}
	return strings.Join(parts, " ")
}

// CreateTableOptions returns dialect-specific CREATE TABLE options for SQLite.
// SQLite supports options like "WITHOUT ROWID" for optimized tables.
func (d *SQLiteDialect) CreateTableOptions() string {
	return "" // Can be overridden to "WITHOUT ROWID" for specific use cases
}

// OnConflictDoNothing returns the ON CONFLICT DO NOTHING clause for SQLite.
// SQLite supports both "INSERT OR IGNORE" and "ON CONFLICT DO NOTHING" (3.24+).
func (d *SQLiteDialect) OnConflictDoNothing() string {
	return "ON CONFLICT DO NOTHING"
}

// OnConflictDoUpdate returns the ON CONFLICT DO UPDATE clause for SQLite.
// SQLite supports UPSERT syntax (3.24+).
func (d *SQLiteDialect) OnConflictDoUpdate(column string, updates []string) string {
	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", d.QuoteIdentifier(column), strings.Join(updates, ", "))
}

// SupportsReturning returns true as SQLite 3.35+ supports RETURNING clause.
// Note: For older SQLite versions, this should return false.
func (d *SQLiteDialect) SupportsReturning() bool {
	return true // SQLite 3.35+ supports RETURNING
}

// ConcatOperator returns the string concatenation operator for SQLite.
// SQLite uses the || operator for string concatenation.
func (d *SQLiteDialect) ConcatOperator() string {
	return "||"
}

// LikeEscape returns the escape character for LIKE patterns in SQLite.
func (d *SQLiteDialect) LikeEscape() string {
	return "\\"
}

// Ensure SQLiteDialect implements Dialect interface
var _ Dialect = (*SQLiteDialect)(nil)
