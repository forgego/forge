package execute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitSQL(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected []string
	}{
		{
			name:     "basic multiple statements",
			sql:      "CREATE TABLE a(x int); CREATE TABLE b(y int);",
			expected: []string{"CREATE TABLE a(x int)", "CREATE TABLE b(y int)"},
		},
		{
			name:     "string with semicolon",
			sql:      "INSERT INTO t VALUES ('a;b'); SELECT 1",
			expected: []string{"INSERT INTO t VALUES ('a;b')", "SELECT 1"},
		},
		{
			name:     "string with escaped quote and semicolon",
			sql:      "INSERT INTO t VALUES ('it''s;ok');",
			expected: []string{"INSERT INTO t VALUES ('it''s;ok')"},
		},
		{
			name:     "quoted identifier with semicolon",
			sql:      `SELECT "we;ird" FROM t;`,
			expected: []string{`SELECT "we;ird" FROM t`},
		},
		{
			name:     "trailing line comment with semicolon",
			sql:      "SELECT 1; -- trailing; comment\nSELECT 2;",
			expected: []string{"SELECT 1", "-- trailing; comment\nSELECT 2"},
		},
		{
			name:     "block comment with semicolon",
			sql:      "/* a; b */ SELECT 1;",
			expected: []string{"/* a; b */ SELECT 1"},
		},
		{
			name:     "dollar-quoted body with semicolons",
			sql:      "CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql; SELECT 2;",
			expected: []string{"CREATE FUNCTION f() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql", "SELECT 2"},
		},
		{
			name:     "tagged dollar-quoted body with semicolons",
			sql:      "DO $body$ BEGIN PERFORM 1; END $body$; SELECT 3",
			expected: []string{"DO $body$ BEGIN PERFORM 1; END $body$", "SELECT 3"},
		},
		{
			name:     "only whitespace and semicolons",
			sql:      "  ;  ; ",
			expected: nil,
		},
		{
			name:     "positional parameter not dollar quote",
			sql:      "SELECT $1; SELECT 2",
			expected: []string{"SELECT $1", "SELECT 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitSQL(tt.sql)
			if len(tt.expected) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
