package orm

import (
	"testing"
)

// BenchmarkFieldExpression_ToSQL benchmarks field expression SQL generation
func BenchmarkFieldExpression_ToSQL(b *testing.B) {
	field := NewField[string]("name", "users")
	builder := NewSQLBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_, _, _ = field.ToSQL(builder)
	}
}

// BenchmarkComparisonExpression_ToSQL benchmarks comparison expression SQL generation
func BenchmarkComparisonExpression_ToSQL(b *testing.B) {
	priceField := NewField[float64]("price", "books")
	expr := priceField.Gt(10.0)
	builder := NewSQLBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_, _, _ = expr.ToSQL(builder)
	}
}

// BenchmarkCombinedExpression_ToSQL benchmarks combined expression SQL generation
func BenchmarkCombinedExpression_ToSQL(b *testing.B) {
	priceField := NewField[float64]("price", "books")
	taxField := NewField[float64]("tax", "books")
	expr := priceField.Add(taxField)
	builder := NewSQLBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_, _, _ = expr.ToSQL(builder)
	}
}

// BenchmarkQ_ToSQL benchmarks Q object SQL generation
func BenchmarkQ_ToSQL(b *testing.B) {
	priceField := NewField[float64]("price", "books")
	availableField := NewField[bool]("available", "books")

	q := NewQ(priceField.Gt(10.0)).
		And(NewQ(availableField.Eq(true))).
		Or(NewQ(priceField.Lt(5.0)))

	builder := NewSQLBuilder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_, _, _ = q.ToSQL(builder)
	}
}

// BenchmarkSQLBuilder_BuildSelect benchmarks SELECT clause building
func BenchmarkSQLBuilder_BuildSelect(b *testing.B) {
	builder := NewSQLBuilder()
	fields := []string{"id", "name", "email", "price", "available"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_ = builder.BuildSelect("users", fields, false)
	}
}

// BenchmarkSQLBuilder_BuildWhere benchmarks WHERE clause building
func BenchmarkSQLBuilder_BuildWhere(b *testing.B) {
	builder := NewSQLBuilder()
	conditions := []QueryExpr{
		NewFieldQueryExpr("price", OpGreater, 10.0),
		NewFieldQueryExpr("available", OpEquals, true),
		NewFieldQueryExpr("name", OpContains, "test"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		builder.Reset()
		_, _ = builder.BuildWhere(conditions, []QueryExpr{})
	}
}

// BenchmarkQuerySet_BuildSQL benchmarks full SQL query building
func BenchmarkQuerySet_BuildSQL(b *testing.B) {
	qs, _ := NewQuerySet[testModel]("test_table")
	priceField := NewField[float64]("price", "test_table")
	availableField := NewField[bool]("available", "test_table")

	qs = qs.
		Filter(priceField.Gt(10.0)).
		Filter(availableField.Eq(true)).
		OrderBy(Desc("price")).
		Limit(10).
		Offset(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = qs.(*BaseQuerySet[testModel]).buildSQL()
	}
}

// BenchmarkEscapeIdentifier benchmarks identifier escaping
func BenchmarkEscapeIdentifier(b *testing.B) {
	identifiers := []string{
		"users",
		"user_name",
		"userName",
		"user_name_table",
		"very_long_table_name_with_underscores",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, id := range identifiers {
			_ = EscapeIdentifier(id)
		}
	}
}

// BenchmarkQuerySet_Chaining benchmarks query set method chaining
func BenchmarkQuerySet_Chaining(b *testing.B) {
	qs, _ := NewQuerySet[testModel]("test_table")
	priceField := NewField[float64]("price", "test_table")
	availableField := NewField[bool]("available", "test_table")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qs.
			Filter(priceField.Gt(10.0)).
			Filter(availableField.Eq(true)).
			OrderBy(Desc("price")).
			Limit(10).
			Offset(0)
	}
}

// BenchmarkUpdateBuilder_Chaining benchmarks update builder method chaining
func BenchmarkUpdateBuilder_Chaining(b *testing.B) {
	qs, _ := NewQuerySet[testModel]("test_table")
	ub, err := NewUpdateBuilder[testModel](qs)
	if err != nil {
		b.Skip("Schema not registered")
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ub.
			Set("name", "Test").
			Set("price", 10.0).
			Increment("id", int64(1))
	}
}

// BenchmarkStringOperations benchmarks string operation expressions
func BenchmarkStringOperations(b *testing.B) {
	nameField := NewField[string]("name", "books")
	builder := NewSQLBuilder()

	operations := []func() (Expression, error){
		func() (Expression, error) {
			return nameField.Contains("test"), nil
		},
		func() (Expression, error) {
			return nameField.StartsWith("test"), nil
		},
		func() (Expression, error) {
			return nameField.EndsWith("test"), nil
		},
		func() (Expression, error) {
			return nameField.IContains("test"), nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, op := range operations {
			expr, _ := op()
			builder.Reset()
			_, _, _ = expr.ToSQL(builder)
		}
	}
}
