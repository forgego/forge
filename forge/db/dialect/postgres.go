package dialect

import (
	"fmt"
	"strings"
)

// PostgreSQLDialect implements Dialect for PostgreSQL databases.
type PostgreSQLDialect struct {
	BaseDialect
}

// NewPostgreSQLDialect creates a new PostgreSQL dialect instance.
func NewPostgreSQLDialect() *PostgreSQLDialect {
	return &PostgreSQLDialect{
		BaseDialect: NewBaseDialect("postgres", PositionalStyle, `"`, true),
	}
}

// AutoIncrementType returns the auto-increment type declaration for PostgreSQL.
// PostgreSQL uses "GENERATED ALWAYS AS IDENTITY" for modern tables (PostgreSQL 10+)
// or "SERIAL" for legacy compatibility.
func (d *PostgreSQLDialect) AutoIncrementType() string {
	return "GENERATED ALWAYS AS IDENTITY"
}

// CurrentTime returns the current time function for PostgreSQL.
func (d *PostgreSQLDialect) CurrentTime() string {
	return "NOW()"
}

// CurrentTimestamp returns the current timestamp function for PostgreSQL.
func (d *PostgreSQLDialect) CurrentTimestamp() string {
	return "CURRENT_TIMESTAMP"
}

// BooleanLiteral returns the boolean literal representation for PostgreSQL.
// PostgreSQL supports TRUE and FALSE keywords directly.
func (d *PostgreSQLDialect) BooleanLiteral(value bool) string {
	if value {
		return "TRUE"
	}
	return "FALSE"
}

// LimitOffset generates LIMIT and OFFSET clause for PostgreSQL.
// PostgreSQL supports standard LIMIT/OFFSET syntax.
func (d *PostgreSQLDialect) LimitOffset(limit, offset int) string {
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

// CreateTableOptions returns dialect-specific CREATE TABLE options for PostgreSQL.
// PostgreSQL doesn't have special table options like SQLite's WITHOUT ROWID.
func (d *PostgreSQLDialect) CreateTableOptions() string {
	return ""
}

// OnConflictDoNothing returns the ON CONFLICT DO NOTHING clause for PostgreSQL.
func (d *PostgreSQLDialect) OnConflictDoNothing() string {
	return "ON CONFLICT DO NOTHING"
}

// OnConflictDoUpdate returns the ON CONFLICT DO UPDATE clause for PostgreSQL.
func (d *PostgreSQLDialect) OnConflictDoUpdate(column string, updates []string) string {
	return fmt.Sprintf("ON CONFLICT (%s) DO UPDATE SET %s", d.QuoteIdentifier(column), strings.Join(updates, ", "))
}

// SupportsReturning returns true as PostgreSQL supports RETURNING clause.
func (d *PostgreSQLDialect) SupportsReturning() bool {
	return true
}

// ConcatOperator returns the string concatenation operator for PostgreSQL.
// PostgreSQL supports both || operator and CONCAT function.
func (d *PostgreSQLDialect) ConcatOperator() string {
	return "||"
}

// LikeEscape returns the escape character for LIKE patterns in PostgreSQL.
func (d *PostgreSQLDialect) LikeEscape() string {
	return "\\"
}

// Ensure PostgreSQLDialect implements Dialect interface
var _ Dialect = (*PostgreSQLDialect)(nil)
