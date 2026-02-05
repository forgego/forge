package orm

// FieldExpr provides type-safe field access for queries
//
// Deprecated: Use Field[T] (from NewField) instead. FieldExpr will be removed in v2.0.
// Field[T] provides the same functionality with a cleaner API that matches the Expression interface.
//
// Migration:
//   // Old
//   f := orm.NewFieldExpr[string]("name", "users")
//   // New
//   f := orm.NewField[string]("name", "users")
type FieldExpr[T any] struct {
	name  string
	table string
}

// NewFieldExpr creates a new FieldExpr
//
// Deprecated: Use NewField instead. NewFieldExpr will be removed in v2.0.
func NewFieldExpr[T any](name, table string) FieldExpr[T] {
	return FieldExpr[T]{
		name:  name,
		table: table,
	}
}

// Name returns the field name
func (f FieldExpr[T]) Name() string {
	return f.name
}

// Table returns the table name
func (f FieldExpr[T]) Table() string {
	return f.table
}

// Equals creates a QueryExpr for equality
func (f FieldExpr[T]) Equals(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpEquals,
		value: val,
	}
}

// NotEquals creates a QueryExpr for inequality
func (f FieldExpr[T]) NotEquals(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpNotEquals,
		value: val,
	}
}

// In creates a QueryExpr for IN clause
func (f FieldExpr[T]) In(vals ...T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpIn,
		value: vals,
	}
}

// IsNull creates a QueryExpr for IS NULL
func (f FieldExpr[T]) IsNull() QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpIsNull,
	}
}

// IsNotNull creates a QueryExpr for IS NOT NULL
func (f FieldExpr[T]) IsNotNull() QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpIsNotNull,
	}
}

// Greater creates a QueryExpr for greater than (numeric types)
func (f FieldExpr[T]) Greater(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpGreater,
		value: val,
	}
}

// GreaterOrEqual creates a QueryExpr for greater than or equal
func (f FieldExpr[T]) GreaterOrEqual(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpGreaterOrEqual,
		value: val,
	}
}

// Less creates a QueryExpr for less than
func (f FieldExpr[T]) Less(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpLess,
		value: val,
	}
}

// LessOrEqual creates a QueryExpr for less than or equal
func (f FieldExpr[T]) LessOrEqual(val T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpLessOrEqual,
		value: val,
	}
}

// Range creates a QueryExpr for range (between)
// nolint:gocritic // builtinShadow: min/max are clear parameter names in this context
func (f FieldExpr[T]) Range(min, max T) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpRange,
		value: []T{min, max},
	}
}

// String-specific methods (only available for FieldExpr[string])

// Contains creates a QueryExpr for LIKE '%value%' (string only)
func (f FieldExpr[string]) Contains(val string) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpContains,
		value: val,
	}
}

// StartsWith creates a QueryExpr for LIKE 'value%' (string only)
func (f FieldExpr[string]) StartsWith(val string) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpStartsWith,
		value: val,
	}
}

// EndsWith creates a QueryExpr for LIKE '%value' (string only)
func (f FieldExpr[string]) EndsWith(val string) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpEndsWith,
		value: val,
	}
}

// IContains creates a QueryExpr for ILIKE '%value%' (case-insensitive, string only)
func (f FieldExpr[string]) IContains(val string) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpIContains,
		value: val,
	}
}

// IExact creates a QueryExpr for ILIKE 'value' (case-insensitive exact, string only)
func (f FieldExpr[string]) IExact(val string) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpIExact,
		value: val,
	}
}

// Date/time methods (for time.Time fields)

// Year creates a QueryExpr for year extraction
func (f FieldExpr[T]) Year(n int) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpYear,
		value: n,
	}
}

// Month creates a QueryExpr for month extraction
func (f FieldExpr[T]) Month(n int) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpMonth,
		value: n,
	}
}

// Day creates a QueryExpr for day extraction
func (f FieldExpr[T]) Day(n int) QueryExpr {
	return QueryExpr{
		field: f.name,
		op:    OpDay,
		value: n,
	}
}
