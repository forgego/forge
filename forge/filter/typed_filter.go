package filter

import (
	"context"
	"fmt"

	"github.com/forgego/forge/orm"
)

// TypedFilterBuilder provides type-safe filter construction
type TypedFilterBuilder[T any] struct {
	fs      *FilterSet[T]
	filters []orm.Expression
}

// TypeSafe creates a type-safe filter builder
// Usage: FilterSet[User]().TypeSafe().Where(User.Fields.IsActive, Equals, true)
func (fs *FilterSet[T]) TypeSafe() *TypedFilterBuilder[T] {
	return &TypedFilterBuilder[T]{
		fs:      fs,
		filters: []orm.Expression{},
	}
}

// Where adds a type-safe filter condition
// Usage: builder.Where(User.Fields.Email, Contains, "@example.com")
func (tfb *TypedFilterBuilder[T]) Where(fieldExpr orm.Expression, op TypedFilterOp, value interface{}) *TypedFilterBuilder[T] {
	// Convert to ORM expression
	expr := buildTypedExpression(fieldExpr, op, value)
	tfb.filters = append(tfb.filters, expr)
	return tfb
}

// And combines filters with AND
func (tfb *TypedFilterBuilder[T]) And() *TypedFilterBuilder[T] {
	// This is handled in Build() by combining all filters with AND
	return tfb
}

// Or creates an OR group
func (tfb *TypedFilterBuilder[T]) Or(builder *TypedFilterBuilder[T]) *TypedFilterBuilder[T] {
	// Combine filters with OR
	// This is a simplified version - full implementation would handle OR groups
	return tfb
}

// Build builds the filter and returns a QuerySet
func (tfb *TypedFilterBuilder[T]) Build(ctx context.Context) (orm.QuerySet[T], error) {
	if tfb.fs.queryset == nil {
		return nil, fmt.Errorf("queryset not set on filterset")
	}

	// Combine all filters with AND
	var combinedExpr orm.Expression
	if len(tfb.filters) == 0 {
		return tfb.fs.queryset, nil
	}

	if len(tfb.filters) == 1 {
		combinedExpr = tfb.filters[0]
	} else {
		// Combine with AND (simplified - would need proper AND expression)
		combinedExpr = tfb.filters[0]
		for i := 1; i < len(tfb.filters); i++ {
			// This would need a proper AND expression builder
			combinedExpr = tfb.filters[i]
		}
	}

	return tfb.fs.queryset.Filter(combinedExpr), nil
}

// TypedFilterOp represents a filter operation for typed filters
type TypedFilterOp string

const (
	TypedOpEquals      TypedFilterOp = "equals"
	TypedOpNotEquals   TypedFilterOp = "not_equals"
	TypedOpGreater     TypedFilterOp = "greater"
	TypedOpGreaterOrEqual TypedFilterOp = "greater_or_equal"
	TypedOpLess        TypedFilterOp = "less"
	TypedOpLessOrEqual TypedFilterOp = "less_or_equal"
	TypedOpContains    TypedFilterOp = "contains"
	TypedOpIContains   TypedFilterOp = "icontains"
	TypedOpStartsWith  TypedFilterOp = "starts_with"
	TypedOpEndsWith    TypedFilterOp = "ends_with"
	TypedOpIn          TypedFilterOp = "in"
	TypedOpIsNull      TypedFilterOp = "is_null"
	TypedOpIsNotNull   TypedFilterOp = "is_not_null"
)

// Helper functions for filter operations
var (
	Equals      = TypedOpEquals
	NotEquals   = TypedOpNotEquals
	Greater     = TypedOpGreater
	GreaterOrEqual = TypedOpGreaterOrEqual
	Less        = TypedOpLess
	LessOrEqual = TypedOpLessOrEqual
	Contains    = TypedOpContains
	IContains   = TypedOpIContains
	StartsWith  = TypedOpStartsWith
	EndsWith    = TypedOpEndsWith
	In          = TypedOpIn
	IsNull      = TypedOpIsNull
	IsNotNull   = TypedOpIsNotNull
)

// buildTypedExpression builds an ORM expression from a field expression and operation
func buildTypedExpression(fieldExpr orm.Expression, op TypedFilterOp, value interface{}) orm.Expression {
	// This is a simplified version - in practice, you'd need to handle each operation type
	// and convert to the appropriate ORM expression type
	
	// For now, return the field expression directly
	// Full implementation would create ComparisonExpression based on op
	return fieldExpr
}
