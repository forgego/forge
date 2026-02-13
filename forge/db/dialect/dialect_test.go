package dialect

import (
	"testing"
)

func TestPostgreSQLDialect_Name(t *testing.T) {
	d := NewPostgreSQLDialect()
	if d.Name() != "postgres" {
		t.Errorf("expected name 'postgres', got '%s'", d.Name())
	}
}

func TestPostgreSQLDialect_Placeholder(t *testing.T) {
	d := NewPostgreSQLDialect()

	tests := []struct {
		position  int
		expected  string
	}{
		{1, "$1"},
		{2, "$2"},
		{10, "$10"},
		{100, "$100"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.Placeholder(tt.position)
			if result != tt.expected {
				t.Errorf("Placeholder(%d) = %s, want %s", tt.position, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLDialect_BuildPlaceholders(t *testing.T) {
	d := NewPostgreSQLDialect()

	tests := []struct {
		n        int
		expected string
	}{
		{0, ""},
		{1, "$1"},
		{3, "$1, $2, $3"},
		{5, "$1, $2, $3, $4, $5"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.BuildPlaceholders(tt.n)
			if result != tt.expected {
				t.Errorf("BuildPlaceholders(%d) = %s, want %s", tt.n, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLDialect_QuoteIdentifier(t *testing.T) {
	d := NewPostgreSQLDialect()

	tests := []struct {
		name     string
		expected string
	}{
		{"table", `"table"`},
		{"column_name", `"column_name"`},
		{`table"with"quotes`, `"table""with""quotes"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.QuoteIdentifier(tt.name)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%s) = %s, want %s", tt.name, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLDialect_QuoteString(t *testing.T) {
	d := NewPostgreSQLDialect()

	tests := []struct {
		input    string
		expected string
	}{
		{"hello", `'hello'`},
		{"it's", `'it''s'`},
		{"", `''`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := d.QuoteString(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteString(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLDialect_SupportsReturning(t *testing.T) {
	d := NewPostgreSQLDialect()
	if !d.SupportsReturning() {
		t.Error("PostgreSQL should support RETURNING clause")
	}
}

func TestPostgreSQLDialect_AutoIncrementType(t *testing.T) {
	d := NewPostgreSQLDialect()
	result := d.AutoIncrementType()
	if result != "GENERATED ALWAYS AS IDENTITY" {
		t.Errorf("AutoIncrementType() = %s, want 'GENERATED ALWAYS AS IDENTITY'", result)
	}
}

func TestPostgreSQLDialect_CurrentTime(t *testing.T) {
	d := NewPostgreSQLDialect()
	if d.CurrentTime() != "NOW()" {
		t.Errorf("CurrentTime() = %s, want 'NOW()'", d.CurrentTime())
	}
}

func TestPostgreSQLDialect_BooleanLiteral(t *testing.T) {
	d := NewPostgreSQLDialect()

	if d.BooleanLiteral(true) != "TRUE" {
		t.Errorf("BooleanLiteral(true) = %s, want 'TRUE'", d.BooleanLiteral(true))
	}

	if d.BooleanLiteral(false) != "FALSE" {
		t.Errorf("BooleanLiteral(false) = %s, want 'FALSE'", d.BooleanLiteral(false))
	}
}

func TestPostgreSQLDialect_LimitOffset(t *testing.T) {
	d := NewPostgreSQLDialect()

	tests := []struct {
		limit    int
		offset   int
		expected string
	}{
		{0, 0, ""},
		{10, 0, "LIMIT 10"},
		{0, 5, "OFFSET 5"}, // OFFSET without LIMIT is valid SQL
		{10, 5, "LIMIT 10 OFFSET 5"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.LimitOffset(tt.limit, tt.offset)
			if result != tt.expected {
				t.Errorf("LimitOffset(%d, %d) = %s, want %s", tt.limit, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestPostgreSQLDialect_OnConflictDoNothing(t *testing.T) {
	d := NewPostgreSQLDialect()
	if d.OnConflictDoNothing() != "ON CONFLICT DO NOTHING" {
		t.Errorf("OnConflictDoNothing() = %s, want 'ON CONFLICT DO NOTHING'", d.OnConflictDoNothing())
	}
}

func TestPostgreSQLDialect_OnConflictDoUpdate(t *testing.T) {
	d := NewPostgreSQLDialect()
	result := d.OnConflictDoUpdate("id", []string{"name = EXCLUDED.name", "updated_at = NOW()"})
	expected := `ON CONFLICT ("id") DO UPDATE SET name = EXCLUDED.name, updated_at = NOW()`
	if result != expected {
		t.Errorf("OnConflictDoUpdate() = %s, want %s", result, expected)
	}
}

// SQLite Dialect Tests

func TestSQLiteDialect_Name(t *testing.T) {
	d := NewSQLiteDialect()
	if d.Name() != "sqlite" {
		t.Errorf("expected name 'sqlite', got '%s'", d.Name())
	}
}

func TestSQLiteDialect_Placeholder(t *testing.T) {
	d := NewSQLiteDialect()

	tests := []struct {
		position  int
		expected  string
	}{
		{1, "?"},
		{2, "?"},
		{10, "?"},
		{100, "?"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.Placeholder(tt.position)
			if result != tt.expected {
				t.Errorf("Placeholder(%d) = %s, want %s", tt.position, result, tt.expected)
			}
		})
	}
}

func TestSQLiteDialect_BuildPlaceholders(t *testing.T) {
	d := NewSQLiteDialect()

	tests := []struct {
		n        int
		expected string
	}{
		{0, ""},
		{1, "?"},
		{3, "?, ?, ?"},
		{5, "?, ?, ?, ?, ?"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.BuildPlaceholders(tt.n)
			if result != tt.expected {
				t.Errorf("BuildPlaceholders(%d) = %s, want %s", tt.n, result, tt.expected)
			}
		})
	}
}

func TestSQLiteDialect_QuoteIdentifier(t *testing.T) {
	d := NewSQLiteDialect()

	tests := []struct {
		name     string
		expected string
	}{
		{"table", `"table"`},
		{"column_name", `"column_name"`},
		{`table"with"quotes`, `"table""with""quotes"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.QuoteIdentifier(tt.name)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%s) = %s, want %s", tt.name, result, tt.expected)
			}
		})
	}
}

func TestSQLiteDialect_SupportsReturning(t *testing.T) {
	d := NewSQLiteDialect()
	if !d.SupportsReturning() {
		t.Error("SQLite 3.35+ should support RETURNING clause")
	}
}

func TestSQLiteDialect_AutoIncrementType(t *testing.T) {
	d := NewSQLiteDialect()
	result := d.AutoIncrementType()
	if result != "AUTOINCREMENT" {
		t.Errorf("AutoIncrementType() = %s, want 'AUTOINCREMENT'", result)
	}
}

func TestSQLiteDialect_CurrentTime(t *testing.T) {
	d := NewSQLiteDialect()
	if d.CurrentTime() != "datetime('now')" {
		t.Errorf("CurrentTime() = %s, want 'datetime('now')'", d.CurrentTime())
	}
}

func TestSQLiteDialect_BooleanLiteral(t *testing.T) {
	d := NewSQLiteDialect()

	if d.BooleanLiteral(true) != "1" {
		t.Errorf("BooleanLiteral(true) = %s, want '1'", d.BooleanLiteral(true))
	}

	if d.BooleanLiteral(false) != "0" {
		t.Errorf("BooleanLiteral(false) = %s, want '0'", d.BooleanLiteral(false))
	}
}

func TestSQLiteDialect_LimitOffset(t *testing.T) {
	d := NewSQLiteDialect()

	tests := []struct {
		limit    int
		offset   int
		expected string
	}{
		{0, 0, ""},
		{10, 0, "LIMIT 10"},
		{0, 5, "OFFSET 5"}, // OFFSET without LIMIT is valid SQL
		{10, 5, "LIMIT 10 OFFSET 5"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := d.LimitOffset(tt.limit, tt.offset)
			if result != tt.expected {
				t.Errorf("LimitOffset(%d, %d) = %s, want %s", tt.limit, tt.offset, result, tt.expected)
			}
		})
	}
}

func TestSQLiteDialect_OnConflictDoNothing(t *testing.T) {
	d := NewSQLiteDialect()
	if d.OnConflictDoNothing() != "ON CONFLICT DO NOTHING" {
		t.Errorf("OnConflictDoNothing() = %s, want 'ON CONFLICT DO NOTHING'", d.OnConflictDoNothing())
	}
}

func TestSQLiteDialect_OnConflictDoUpdate(t *testing.T) {
	d := NewSQLiteDialect()
	result := d.OnConflictDoUpdate("id", []string{"name = EXCLUDED.name", "updated_at = datetime('now')"})
	expected := `ON CONFLICT ("id") DO UPDATE SET name = EXCLUDED.name, updated_at = datetime('now')`
	if result != expected {
		t.Errorf("OnConflictDoUpdate() = %s, want %s", result, expected)
	}
}

// Dialect Detection Tests

func TestDetectDialectFromDSN(t *testing.T) {
	tests := []struct {
		dsn      string
		expected string
	}{
		{"postgres://user:pass@localhost/db", "postgres"},
		{"postgresql://user:pass@localhost/db", "postgres"},
		{"host=localhost port=5432 user=postgres dbname=test", "postgres"},
		{"host=localhost sslmode=disable", "postgres"},
		{"file:test.db", "sqlite"},
		{"test.db", "sqlite"},
		{"test.sqlite", "sqlite"},
		{"test.sqlite3", "sqlite"},
		{":memory:", "sqlite"},
		{"/path/to/database.db", "sqlite"},
	}

	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			d := DetectDialectFromDSN(tt.dsn)
			if d.Name() != tt.expected {
				t.Errorf("DetectDialectFromDSN(%s) = %s, want %s", tt.dsn, d.Name(), tt.expected)
			}
		})
	}
}

func TestDetectDialectFromDriver(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{"postgres", "postgres"},
		{"postgresql", "postgres"},
		{"pgx", "postgres"},
		{"sqlite", "sqlite"},
		{"sqlite3", "sqlite"},
		{"unknown", "postgres"}, // defaults to postgres
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			d := DetectDialectFromDriver(tt.driver)
			if d.Name() != tt.expected {
				t.Errorf("DetectDialectFromDriver(%s) = %s, want %s", tt.driver, d.Name(), tt.expected)
			}
		})
	}
}

// Interface compliance tests

func TestPostgreSQLDialect_ImplementsDialect(t *testing.T) {
	var _ Dialect = NewPostgreSQLDialect()
}

func TestSQLiteDialect_ImplementsDialect(t *testing.T) {
	var _ Dialect = NewSQLiteDialect()
}
