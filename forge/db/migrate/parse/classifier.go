package parse

import (
	"strings"
)

// StatementKind represents the type of SQL statement
type StatementKind int

const (
	StmtUnknown StatementKind = iota
	StmtCreateTable
	StmtAlterTable
	StmtCreateIndex
	StmtDropTable
	StmtDropIndex
	StmtDropColumn
	StmtNonDDL // SELECT, INSERT, UPDATE, DELETE, etc.
)

// Classifier classifies SQL statements by type
type Classifier struct{}

// NewClassifier creates a new statement classifier
func NewClassifier() *Classifier {
	return &Classifier{}
}

// Classify determines the kind of SQL statement
// Uses fast case-insensitive prefix matching to skip non-DDL statements
func (c *Classifier) Classify(stmt string) StatementKind {
	// Trim whitespace and normalize
	stmt = strings.TrimSpace(stmt)
	if stmt == "" {
		return StmtUnknown
	}

	// Convert to uppercase for comparison
	upper := strings.ToUpper(stmt)

	// Fast prefix matching for DDL statements
	switch {
	case strings.HasPrefix(upper, "CREATE TABLE"):
		return StmtCreateTable
	case strings.HasPrefix(upper, "ALTER TABLE"):
		return StmtAlterTable
	case strings.HasPrefix(upper, "CREATE INDEX") || strings.HasPrefix(upper, "CREATE UNIQUE INDEX"):
		return StmtCreateIndex
	case strings.HasPrefix(upper, "DROP TABLE"):
		return StmtDropTable
	case strings.HasPrefix(upper, "DROP INDEX"):
		return StmtDropIndex
	case strings.Contains(upper, "ALTER TABLE") && strings.Contains(upper, "DROP COLUMN"):
		return StmtDropColumn
	}

	// Check for non-DDL statements (skip these)
	nonDDLPrefixes := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE",
		"WITH", "EXPLAIN", "ANALYZE",
		"GRANT", "REVOKE", "COMMIT", "ROLLBACK",
		"SET ", "SHOW", "DESCRIBE", "DESC",
	}

	for _, prefix := range nonDDLPrefixes {
		if strings.HasPrefix(upper, prefix) {
			return StmtNonDDL
		}
	}

	// Unknown statement - might be DDL we don't recognize yet
	return StmtUnknown
}

// IsDDL returns true if the statement kind is a DDL statement
func (k StatementKind) IsDDL() bool {
	switch k {
	case StmtCreateTable, StmtAlterTable, StmtCreateIndex,
		StmtDropTable, StmtDropIndex, StmtDropColumn:
		return true
	default:
		return false
	}
}

// String returns a string representation of the statement kind
func (k StatementKind) String() string {
	switch k {
	case StmtCreateTable:
		return "CREATE TABLE"
	case StmtAlterTable:
		return "ALTER TABLE"
	case StmtCreateIndex:
		return "CREATE INDEX"
	case StmtDropTable:
		return "DROP TABLE"
	case StmtDropIndex:
		return "DROP INDEX"
	case StmtDropColumn:
		return "DROP COLUMN"
	case StmtNonDDL:
		return "NON-DDL"
	default:
		return "UNKNOWN"
	}
}
