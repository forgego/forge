package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLBuilder_AddArg(t *testing.T) {
	builder := NewSQLBuilder()

	placeholder1 := builder.AddArg("value1")
	placeholder2 := builder.AddArg("value2")
	placeholder3 := builder.AddArg(42)

	assert.Equal(t, "$1", placeholder1)
	assert.Equal(t, "$2", placeholder2)
	assert.Equal(t, "$3", placeholder3)
	assert.Equal(t, []interface{}{"value1", "value2", 42}, builder.Args())
}

func TestSQLBuilder_AddArgs(t *testing.T) {
	builder := NewSQLBuilder()

	values := []interface{}{"a", "b", "c"}
	placeholders := builder.AddArgs(values)

	assert.Len(t, placeholders, 3)
	assert.Equal(t, "$1", placeholders[0])
	assert.Equal(t, "$2", placeholders[1])
	assert.Equal(t, "$3", placeholders[2])
	assert.Equal(t, values, builder.Args())
}

func TestSQLBuilder_Reset(t *testing.T) {
	builder := NewSQLBuilder()
	builder.AddArg("value1")
	builder.AddArg("value2")

	builder.Reset()

	assert.Equal(t, 1, builder.paramIndex)
	assert.Len(t, builder.Args(), 0)
}

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple identifier",
			input:    "users",
			expected: `"users"`,
		},
		{
			name:     "identifier with quotes",
			input:    `user"name`,
			expected: `"user""name"`,
		},
		{
			name:     "identifier with underscore",
			input:    "user_name",
			expected: `"user_name"`,
		},
		{
			name:     "mixed case",
			input:    "UserName",
			expected: `"UserName"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeIdentifierList(t *testing.T) {
	identifiers := []string{"users", "books", "authors"}
	escaped := EscapeIdentifierList(identifiers)

	assert.Len(t, escaped, 3)
	assert.Equal(t, `"users"`, escaped[0])
	assert.Equal(t, `"books"`, escaped[1])
	assert.Equal(t, `"authors"`, escaped[2])
}

func TestSQLBuilder_BuildSelect(t *testing.T) {
	builder := NewSQLBuilder()

	t.Run("select all", func(t *testing.T) {
		sql := builder.BuildSelect("users", []string{}, false)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, "*")
		assert.Contains(t, sql, "FROM")
		assert.Contains(t, sql, `"users"`)
	})

	t.Run("select specific fields", func(t *testing.T) {
		sql := builder.BuildSelect("users", []string{"id", "name"}, false)
		assert.Contains(t, sql, "SELECT")
		assert.Contains(t, sql, `"id"`)
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, "FROM")
	})

	t.Run("select distinct", func(t *testing.T) {
		sql := builder.BuildSelect("users", []string{"email"}, true)
		assert.Contains(t, sql, "SELECT DISTINCT")
	})
}

func TestSQLBuilder_BuildWhere(t *testing.T) {
	builder := NewSQLBuilder()

	t.Run("single condition", func(t *testing.T) {
		conditions := []QueryExpr{
			NewFieldQueryExpr("price", OpGreater, 10.0),
		}
		where, args := builder.BuildWhere(conditions, []QueryExpr{})
		assert.Contains(t, where, "WHERE")
		assert.Contains(t, where, "price")
		assert.Len(t, args, 1)
	})

	t.Run("multiple conditions", func(t *testing.T) {
		conditions := []QueryExpr{
			NewFieldQueryExpr("price", OpGreater, 10.0),
			NewFieldQueryExpr("available", OpEquals, true),
		}
		where, args := builder.BuildWhere(conditions, []QueryExpr{})
		assert.Contains(t, where, "WHERE")
		assert.Contains(t, where, "AND")
		assert.Len(t, args, 2)
	})

	t.Run("with excludes", func(t *testing.T) {
		conditions := []QueryExpr{
			NewFieldQueryExpr("price", OpGreater, 10.0),
		}
		excludes := []QueryExpr{
			NewFieldQueryExpr("deleted", OpEquals, true),
		}
		where, args := builder.BuildWhere(conditions, excludes)
		assert.Contains(t, where, "WHERE")
		assert.Contains(t, where, "NOT")
		assert.Len(t, args, 2)
	})
}

func TestSQLBuilder_BuildOrderBy(t *testing.T) {
	builder := NewSQLBuilder()

	t.Run("single field ascending", func(t *testing.T) {
		sql := builder.BuildOrderBy([]string{"name"})
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, "ASC")
	})

	t.Run("single field descending", func(t *testing.T) {
		sql := builder.BuildOrderBy([]string{"-name"})
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, "DESC")
	})

	t.Run("multiple fields", func(t *testing.T) {
		sql := builder.BuildOrderBy([]string{"-price", "name"})
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, "price")
		assert.Contains(t, sql, "name")
	})
}

func TestSQLBuilder_BuildLimit(t *testing.T) {
	builder := NewSQLBuilder()

	t.Run("with limit", func(t *testing.T) {
		limit := 10
		sql := builder.BuildLimit(&limit)
		assert.Contains(t, sql, "LIMIT")
		assert.Contains(t, sql, "10")
	})

	t.Run("without limit", func(t *testing.T) {
		sql := builder.BuildLimit(nil)
		assert.Empty(t, sql)
	})
}

func TestSQLBuilder_BuildOffset(t *testing.T) {
	builder := NewSQLBuilder()

	t.Run("with offset", func(t *testing.T) {
		offset := 20
		sql := builder.BuildOffset(&offset)
		assert.Contains(t, sql, "OFFSET")
		assert.Contains(t, sql, "20")
	})

	t.Run("without offset", func(t *testing.T) {
		sql := builder.BuildOffset(nil)
		assert.Empty(t, sql)
	})
}

func TestSQLBuilder_BuildUpdate(t *testing.T) {
	builder := NewSQLBuilder()

	fields := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	sql, args := builder.BuildUpdate("users", fields)

	assert.Contains(t, sql, "UPDATE")
	assert.Contains(t, sql, `"users"`)
	assert.Contains(t, sql, "SET")
	assert.Contains(t, sql, `"name"`)
	assert.Contains(t, sql, `"email"`)
	assert.Len(t, args, 2)
}

func TestSQLBuilder_BuildInsert(t *testing.T) {
	builder := NewSQLBuilder()

	fields := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	sql, args := builder.BuildInsert("users", fields)

	assert.Contains(t, sql, "INSERT INTO")
	assert.Contains(t, sql, `"users"`)
	assert.Contains(t, sql, "VALUES")
	assert.Contains(t, sql, `"name"`)
	assert.Contains(t, sql, `"email"`)
	assert.Len(t, args, 2)
}

func TestSQLBuilder_BuildDelete(t *testing.T) {
	builder := NewSQLBuilder()

	sql := builder.BuildDelete("users")

	assert.Contains(t, sql, "DELETE FROM")
	assert.Contains(t, sql, `"users"`)
}
