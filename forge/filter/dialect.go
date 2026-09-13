package filter

import "fmt"

// DialectAdapter adapts filter expressions to different SQL dialects
type DialectAdapter interface {
	// TranslateOperator translates a filter operator to SQL
	TranslateOperator(op string) string

	// CaseInsensitiveLike handles case-insensitive LIKE (ILIKE vs LOWER LIKE)
	CaseInsensitiveLike(field, value string) string

	// JSONContains handles JSON containment operators
	JSONContains(field, path, value string) string

	// Similarity handles similarity operators (Postgres similarity, etc.)
	Similarity(field, value string) string
}

// PostgresAdapter adapts filters for PostgreSQL
type PostgresAdapter struct{}

// TranslateOperator translates operator for Postgres
func (a *PostgresAdapter) TranslateOperator(op string) string {
	return op
}

// CaseInsensitiveLike uses ILIKE for Postgres
func (a *PostgresAdapter) CaseInsensitiveLike(field, value string) string {
	return fmt.Sprintf("%s ILIKE %s", field, value)
}

// JSONContains uses Postgres JSON operators
func (a *PostgresAdapter) JSONContains(field, path, value string) string {
	return fmt.Sprintf("%s->>'%s' = %s", field, path, value)
}

// Similarity uses Postgres similarity operator
func (a *PostgresAdapter) Similarity(field, value string) string {
	return fmt.Sprintf("similarity(%s, %s) > 0.3", field, value)
}

// MySQLAdapter adapts filters for MySQL
type MySQLAdapter struct{}

// TranslateOperator translates operator for MySQL
func (a *MySQLAdapter) TranslateOperator(op string) string {
	return op
}

// CaseInsensitiveLike uses LOWER LIKE for MySQL
func (a *MySQLAdapter) CaseInsensitiveLike(field, value string) string {
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", field, value)
}

// JSONContains uses MySQL JSON functions
func (a *MySQLAdapter) JSONContains(field, path, value string) string {
	return fmt.Sprintf("JSON_EXTRACT(%s, '$.%s') = %s", field, path, value)
}

// Similarity is not supported in MySQL
func (a *MySQLAdapter) Similarity(field, value string) string {
	return fmt.Sprintf("%s LIKE %s", field, value)
}

// SQLiteAdapter adapts filters for SQLite
type SQLiteAdapter struct{}

// TranslateOperator translates operator for SQLite
func (a *SQLiteAdapter) TranslateOperator(op string) string {
	return op
}

// CaseInsensitiveLike uses LOWER LIKE for SQLite
func (a *SQLiteAdapter) CaseInsensitiveLike(field, value string) string {
	return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", field, value)
}

// JSONContains uses SQLite JSON functions
func (a *SQLiteAdapter) JSONContains(field, path, value string) string {
	return fmt.Sprintf("json_extract(%s, '$.%s') = %s", field, path, value)
}

// Similarity is not supported in SQLite
func (a *SQLiteAdapter) Similarity(field, value string) string {
	return fmt.Sprintf("%s LIKE %s", field, value)
}

// GetDialectAdapter returns the appropriate dialect adapter
func GetDialectAdapter(dialect string) DialectAdapter {
	switch dialect {
	case "postgres", "postgresql":
		return &PostgresAdapter{}
	case "mysql", "mariadb":
		return &MySQLAdapter{}
	case "sqlite", "sqlite3":
		return &SQLiteAdapter{}
	default:
		// Default to Postgres
		return &PostgresAdapter{}
	}
}
