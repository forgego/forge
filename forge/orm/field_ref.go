package orm

// FieldPath is an interface for types that can provide a field path string
// This allows both string and FieldExpression[T] to be used in the same methods
type FieldPath interface {
	Path() string
}

// ExtractPathFromAny extracts the field path from any type (string or FieldExpression)
func ExtractPathFromAny(field any) string {
	switch v := field.(type) {
	case string:
		return v
	case FieldPath:
		return v.Path()
	default:
		return ""
	}
}

// OrderFieldSpec is an interface for types that can be used for ordering
// Both OrderField (string-based) and OrderFieldExpr[T] (type-safe) implement this
type OrderFieldSpec interface {
	GetFieldPath() string
	IsAscending() bool
}

// OrderFieldExpr represents a type-safe ordering field with direction
type OrderFieldExpr[T any] struct {
	field     FieldExpression[T]
	ascending bool
}

// GetFieldPath returns the field path for ordering
func (o OrderFieldExpr[T]) GetFieldPath() string {
	return o.field.Path()
}

// IsAscending returns whether the ordering is ascending
func (o OrderFieldExpr[T]) IsAscending() bool {
	return o.ascending
}

// Field returns the underlying field expression
func (o OrderFieldExpr[T]) Field() FieldExpression[T] {
	return o.field
}

// Path returns the string representation of the ordering (e.g. "-created_at")
func (o OrderFieldExpr[T]) Path() string {
	path := o.field.Path()
	if !o.ascending {
		return "-" + path
	}
	return path
}

// RelationPath is an interface for types that can provide a relation path string
type RelationPath interface {
	RelationPath() string
}

// extractRelationPathFromAny extracts the relation path from any type (string or RelationExpression)
func extractRelationPathFromAny(relation any) string {
	switch v := relation.(type) {
	case string:
		return v
	case RelationPath:
		return v.RelationPath()
	default:
		return ""
	}
}

// RelationExpression represents a type-safe relation reference
type RelationExpression struct {
	relationName string
}

// NewRelationExpression creates a new relation expression
func NewRelationExpression(name string) RelationExpression {
	return RelationExpression{
		relationName: name,
	}
}

// RelationPath returns the relation path
func (r RelationExpression) RelationPath() string {
	return r.relationName
}
