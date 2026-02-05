package orm

import (
	"testing"

	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuerySet(t *testing.T) {
	t.Run("create queryset", func(t *testing.T) {
		qs, err := NewQuerySet[testModel]("test_table")
		require.NoError(t, err)
		assert.NotNil(t, qs)
	})

	t.Run("create queryset with empty table name", func(t *testing.T) {
		qs, err := NewQuerySet[testModel]("")
		// Should still work, might use model name
		if err == nil {
			assert.NotNil(t, qs)
		}
	})
}

func TestQuerySet_Filter(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	priceField := NewField[float64]("price", "test_table")
	expr := priceField.Gt(10.0)

	filtered := qs.Filter(expr)
	assert.NotNil(t, filtered)
	assert.NotEqual(t, qs, filtered) // Should return new instance
}

func TestQuerySet_Exclude(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	availableField := NewField[bool]("available", "test_table")
	expr := availableField.Eq(false)

	excluded := qs.Exclude(expr)
	assert.NotNil(t, excluded)
	assert.NotEqual(t, qs, excluded) // Should return new instance
}

func TestQuerySet_OrderBy(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	ordered := qs.OrderBy(Asc("price"))
	assert.NotNil(t, ordered)
	assert.NotEqual(t, qs, ordered)

	// Multiple fields
	multiOrdered := qs.OrderBy(
		Desc("price"),
		Asc("name"),
	)
	assert.NotNil(t, multiOrdered)
}

func TestQuerySet_Reverse(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	qs = qs.OrderBy(Asc("price"))
	reversed := qs.Reverse()
	assert.NotNil(t, reversed)
	assert.NotEqual(t, qs, reversed)
}

func TestQuerySet_Limit(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	limited := qs.Limit(10)
	assert.NotNil(t, limited)
	assert.NotEqual(t, qs, limited)
}

func TestQuerySet_Offset(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	offset := qs.Offset(20)
	assert.NotNil(t, offset)
	assert.NotEqual(t, qs, offset)
}

func TestQuerySet_Distinct(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	distinct := qs.Distinct("email")
	assert.NotNil(t, distinct)
	assert.NotEqual(t, qs, distinct)
}

func TestQuerySet_Select(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	selected := qs.Select("id", "name", "email")
	assert.NotNil(t, selected)
	assert.NotEqual(t, qs, selected)
}

func TestQuerySet_Only(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	only := qs.Only("id", "name")
	assert.NotNil(t, only)
	assert.NotEqual(t, qs, only)
}

func TestQuerySet_Defer(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	deferred := qs.Defer("description", "metadata")
	assert.NotNil(t, deferred)
	assert.NotEqual(t, qs, deferred)
}

func TestQuerySet_SelectRelated(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	related := qs.SelectRelated("author", "category")
	assert.NotNil(t, related)
	assert.NotEqual(t, qs, related)
}

func TestQuerySet_PrefetchRelated(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	prefetch := qs.PrefetchRelated("tags", "comments")
	assert.NotNil(t, prefetch)
	assert.NotEqual(t, qs, prefetch)
}

func TestQuerySet_Aggregate(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// Create aggregate using the correct function
	agg := Aggregate{
		Name:  "total_price",
		Field: "price",
		Func:  string(AggSum),
	}
	aggregated := qs.Aggregate(agg)
	assert.NotNil(t, aggregated)
	assert.NotEqual(t, qs, aggregated)
}

func TestQuerySet_Annotate(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	// Create a simple annotation
	expr := NewFieldQueryExpr("price", OpGreater, 10.0)
	annotation := NewAnnotation("high_price", expr)

	annotated := qs.Annotate(annotation)
	assert.NotNil(t, annotated)
	assert.NotEqual(t, qs, annotated)
}

func TestQuerySet_Values(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	values := qs.Values("id", "name")
	assert.NotNil(t, values)
}

func TestQuerySet_ValuesList(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	valuesList := qs.ValuesList("id", "name")
	assert.NotNil(t, valuesList)
}

func TestQuerySet_Chaining(t *testing.T) {
	qs, err := NewQuerySet[testModel]("test_table")
	require.NoError(t, err)

	priceField := NewField[float64]("price", "test_table")
	availableField := NewField[bool]("available", "test_table")

	// Chain multiple operations
	chained := qs.
		Filter(priceField.Gt(10.0)).
		Filter(availableField.Eq(true)).
		OrderBy(Desc("price")).
		Limit(10).
		Offset(0)

	assert.NotNil(t, chained)
	assert.NotEqual(t, qs, chained)
}

// Note: buildSQL and other build methods are not exported
// These would need to be tested through integration tests or by exporting them
// Skipping these tests for now as they test internal implementation

// testModel is a simple test model for testing
// It implements schema.Schema for UpdateBuilder tests
type testModel struct {
	schema.BaseSchema
	ID        int64   `db:"id"`
	Name      string  `db:"name"`
	Email     string  `db:"email"`
	Price     float64 `db:"price"`
	Available bool    `db:"available"`
}

// Fields returns field definitions
func (testModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email"),
		schema.Float64Field("price", schema.Default(0.0)),
		schema.BoolField("available", schema.Default(true)),
	}
}

// Meta returns model metadata
func (testModel) Meta() schema.Meta {
	return schema.Meta{
		TableName: "test_table",
	}
}

// Relations returns relations
func (testModel) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Hooks returns hooks
func (testModel) Hooks() *schema.ModelHooks {
	return nil
}
