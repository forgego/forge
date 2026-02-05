package orm

import (
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example unit test for FieldExpression
func TestFieldExpression_ToSQL(t *testing.T) {
	tests := []struct {
		name     string
		field    orm.FieldExpression[string]
		expected string
	}{
		{
			name:     "simple field",
			field:    orm.NewField[string]("name", "users"),
			expected: `"users"."name"`,
		},
		{
			name:     "field without table",
			field:    orm.NewField[string]("name", ""),
			expected: `"name"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := orm.NewSQLBuilder()
			sql, _, err := tt.field.ToSQL(builder)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, sql)
		})
	}
}

// Example unit test for ComparisonExpression
func TestComparisonExpression_ToSQL(t *testing.T) {
	tests := []struct {
		name     string
		field    orm.FieldExpression[float64]
		op       orm.Operator
		value    interface{}
		expected string
	}{
		{
			name:     "equals",
			field:    orm.NewField[float64]("price", "books"),
			op:       orm.OpEquals,
			value:    10.0,
			expected: `"books"."price" = $1`,
		},
		{
			name:     "greater than",
			field:    orm.NewField[float64]("price", "books"),
			op:       orm.OpGreater,
			value:    10.0,
			expected: `"books"."price" > $1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := orm.ComparisonExpression[float64]{
				Field: tt.field,
				Op:    tt.op,
				Value: tt.value,
			}

			builder := orm.NewSQLBuilder()
			sql, args, err := expr.ToSQL(builder)
			require.NoError(t, err)
			assert.Contains(t, sql, tt.expected)
			assert.Len(t, args, 1)
			assert.Equal(t, tt.value, args[0])
		})
	}
}

// Example unit test for Q object
func TestQ_And(t *testing.T) {
	priceField := orm.NewField[float64]("price", "books")
	availableField := orm.NewField[bool]("available", "books")

	q1 := orm.NewQ(priceField.Gt(10.0))
	q2 := orm.NewQ(availableField.Eq(true))
	combined := q1.And(q2)

	builder := orm.NewSQLBuilder()
	sql, args, err := combined.ToSQL(builder)
	require.NoError(t, err)

	// Should contain both conditions
	assert.Contains(t, sql, "price")
	assert.Contains(t, sql, "available")
	assert.Contains(t, sql, "AND")
	assert.Len(t, args, 2)
}

// Example unit test for SQLBuilder
func TestSQLBuilder_AddArg(t *testing.T) {
	builder := orm.NewSQLBuilder()

	placeholder1 := builder.AddArg("value1")
	placeholder2 := builder.AddArg("value2")

	assert.Equal(t, "$1", placeholder1)
	assert.Equal(t, "$2", placeholder2)
	assert.Equal(t, []interface{}{"value1", "value2"}, builder.Args())
}

// Example unit test for EscapeIdentifier
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := orm.EscapeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
