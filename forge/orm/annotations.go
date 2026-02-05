package orm

// AnnotationExpr represents a computed field annotation
// Annotations add computed fields to query results
type AnnotationExpr struct {
	Name string
	Expr QueryExpr
}

// NewAnnotation creates a new annotation
// nolint:gocritic // hugeParam: QueryExpr is small enough for value semantics
func NewAnnotation(name string, expr QueryExpr) AnnotationExpr {
	return AnnotationExpr{
		Name: name,
		Expr: expr,
	}
}

// RegisterAnnotation registers a custom annotation type
func RegisterAnnotation(name string, builder func(...interface{}) AnnotationExpr) {
	// TODO: Implement annotation registry
}
