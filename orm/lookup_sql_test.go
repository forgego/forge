package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupSQL_QueryExpr(t *testing.T) {
	t.Run("IN with []string", func(t *testing.T) {
		q := NewFieldQueryExpr("status", OpIn, []string{"a", "b"})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "status IN ($1, $2)", sql)
		assert.Equal(t, []interface{}{"a", "b"}, args)
		assert.Equal(t, 3, nextIdx)
	})

	t.Run("IN with []int", func(t *testing.T) {
		q := NewFieldQueryExpr("id", OpIn, []int{1, 2, 3})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "id IN ($1, $2, $3)", sql)
		assert.Equal(t, []interface{}{1, 2, 3}, args)
		assert.Equal(t, 4, nextIdx)
	})

	t.Run("IN with empty []string", func(t *testing.T) {
		q := NewFieldQueryExpr("status", OpIn, []string{})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "1=0", sql)
		assert.Empty(t, args)
		assert.Equal(t, 1, nextIdx)
	})

	t.Run("NOT IN with empty []string", func(t *testing.T) {
		q := NewFieldQueryExpr("status", OpNotIn, []string{})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "1=1", sql)
		assert.Empty(t, args)
		assert.Equal(t, 1, nextIdx)
	})

	t.Run("range with []string", func(t *testing.T) {
		q := NewFieldQueryExpr("price", OpRange, []string{"1", "9"})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "price BETWEEN $1 AND $2", sql)
		assert.Equal(t, []interface{}{"1", "9"}, args)
		assert.Equal(t, 3, nextIdx)
	})

	t.Run("range with []int of len 1", func(t *testing.T) {
		q := NewFieldQueryExpr("price", OpRange, []int{1})
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "1=0", sql)
		assert.Empty(t, args)
		assert.Equal(t, 1, nextIdx)
	})

	t.Run("isnull true", func(t *testing.T) {
		q := NewFieldQueryExpr("deleted_at", OpIsNull, true)
		sql, args, _ := q.ToSQL(1)
		assert.Equal(t, "deleted_at IS NULL", sql)
		assert.Empty(t, args)
	})

	t.Run("isnull false", func(t *testing.T) {
		q := NewFieldQueryExpr("deleted_at", OpIsNull, false)
		sql, args, _ := q.ToSQL(1)
		assert.Equal(t, "deleted_at IS NOT NULL", sql)
		assert.Empty(t, args)
	})

	t.Run("iexact", func(t *testing.T) {
		q := NewFieldQueryExpr("name", OpIExact, "Bob")
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "LOWER(name) LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"Bob"}, args)
		assert.Equal(t, 2, nextIdx)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("icontains", func(t *testing.T) {
		q := NewFieldQueryExpr("name", OpIContains, "Bob")
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "LOWER(name) LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"%Bob%"}, args)
		assert.Equal(t, 2, nextIdx)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("istartswith", func(t *testing.T) {
		q := NewFieldQueryExpr("name", OpIStartsWith, "Bob")
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "LOWER(name) LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"Bob%"}, args)
		assert.Equal(t, 2, nextIdx)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("iendswith", func(t *testing.T) {
		q := NewFieldQueryExpr("name", OpIEndsWith, "Bob")
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "LOWER(name) LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"%Bob"}, args)
		assert.Equal(t, 2, nextIdx)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("startswith stays LIKE without LOWER", func(t *testing.T) {
		q := NewFieldQueryExpr("name", OpStartsWith, "Bob")
		sql, args, nextIdx := q.ToSQL(1)
		assert.Equal(t, "name LIKE $1", sql)
		assert.Equal(t, []interface{}{"Bob%"}, args)
		assert.Equal(t, 2, nextIdx)
		assert.False(t, strings.Contains(sql, "LOWER"))
	})
}

func TestLookupSQL_ComparisonExpression(t *testing.T) {
	t.Run("IN with []string", func(t *testing.T) {
		expr := Where("status", OpIn, []string{"a", "b"})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"status\" IN ($1, $2)", sql)
		assert.Equal(t, []interface{}{"a", "b"}, args)
	})

	t.Run("IN with []int", func(t *testing.T) {
		expr := Where("id", OpIn, []int{1, 2, 3})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"id\" IN ($1, $2, $3)", sql)
		assert.Equal(t, []interface{}{1, 2, 3}, args)
	})

	t.Run("IN with empty []string", func(t *testing.T) {
		expr := Where("status", OpIn, []string{})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "1=0", sql)
		assert.Empty(t, args)
	})

	t.Run("NOT IN with empty []string", func(t *testing.T) {
		expr := Where("status", OpNotIn, []string{})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "1=1", sql)
		assert.Empty(t, args)
	})

	t.Run("range with []string", func(t *testing.T) {
		expr := Where("price", OpRange, []string{"1", "9"})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"price\" BETWEEN $1 AND $2", sql)
		assert.Equal(t, []interface{}{"1", "9"}, args)
	})

	t.Run("range with []int of len 1", func(t *testing.T) {
		expr := Where("price", OpRange, []int{1})
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "1=0", sql)
		assert.Empty(t, args)
	})

	t.Run("isnull true", func(t *testing.T) {
		expr := Where("deleted_at", OpIsNull, true)
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"deleted_at\" IS NULL", sql)
		assert.Empty(t, args)
	})

	t.Run("isnull false", func(t *testing.T) {
		expr := Where("deleted_at", OpIsNull, false)
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"deleted_at\" IS NOT NULL", sql)
		assert.Empty(t, args)
	})

	t.Run("iexact", func(t *testing.T) {
		expr := Where("name", OpIExact, "Bob")
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "LOWER(\"name\") LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"Bob"}, args)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("icontains", func(t *testing.T) {
		expr := Where("name", OpIContains, "Bob")
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "LOWER(\"name\") LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"%Bob%"}, args)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("istartswith", func(t *testing.T) {
		expr := Where("name", OpIStartsWith, "Bob")
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "LOWER(\"name\") LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"Bob%"}, args)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("iendswith", func(t *testing.T) {
		expr := Where("name", OpIEndsWith, "Bob")
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "LOWER(\"name\") LIKE LOWER($1)", sql)
		assert.Equal(t, []interface{}{"%Bob"}, args)
		assert.False(t, strings.Contains(sql, "ILIKE"))
	})

	t.Run("startswith stays LIKE without LOWER", func(t *testing.T) {
		expr := Where("name", OpStartsWith, "Bob")
		builder := NewSQLBuilder()
		sql, args, err := expr.ToSQL(builder)
		require.NoError(t, err)
		assert.Equal(t, "\"name\" LIKE $1", sql)
		assert.Equal(t, []interface{}{"Bob%"}, args)
		assert.False(t, strings.Contains(sql, "LOWER"))
	})
}
