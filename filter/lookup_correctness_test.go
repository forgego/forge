package filter

import (
	"reflect"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseValue_EmptyLookups(t *testing.T) {
	p := NewParser(WithAllowAllSecurity())

	t.Run("empty in returns empty non-nil slice", func(t *testing.T) {
		val, err := p.parseValue("", "in")
		require.NoError(t, err)
		require.NotNil(t, val, "parseValue(\"\", \"in\") must not return nil")
		slice, ok := val.([]string)
		require.True(t, ok, "val must be []string")
		assert.Empty(t, slice, "slice must be empty")
	})

	t.Run("empty exact returns nil", func(t *testing.T) {
		val, err := p.parseValue("", "exact")
		require.NoError(t, err)
		assert.Nil(t, val, "parseValue(\"\", \"exact\") must return nil")
	})

	t.Run("in with values returns slice", func(t *testing.T) {
		val, err := p.parseValue("foo, bar", "in")
		require.NoError(t, err)
		assert.Equal(t, []string{"foo", "bar"}, val)
	})
}

func TestExpressionConverter_LookupToOperator(t *testing.T) {
	ec := &ExpressionConverter[any]{}
	stringType := reflect.TypeOf("")

	t.Run("istartswith maps to OpIStartsWith", func(t *testing.T) {
		op, err := ec.lookupToOperator("istartswith", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpIStartsWith, op)
	})

	t.Run("iendswith maps to OpIEndsWith", func(t *testing.T) {
		op, err := ec.lookupToOperator("iendswith", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpIEndsWith, op)
	})

	t.Run("startswith maps to OpStartsWith", func(t *testing.T) {
		op, err := ec.lookupToOperator("startswith", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpStartsWith, op)
	})

	t.Run("endswith maps to OpEndsWith", func(t *testing.T) {
		op, err := ec.lookupToOperator("endswith", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpEndsWith, op)
	})

	t.Run("icontains maps to OpIContains", func(t *testing.T) {
		op, err := ec.lookupToOperator("icontains", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpIContains, op)
	})

	t.Run("iexact maps to OpIExact", func(t *testing.T) {
		op, err := ec.lookupToOperator("iexact", stringType)
		require.NoError(t, err)
		assert.Equal(t, orm.OpIExact, op)
	})
}
