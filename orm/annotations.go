package orm

import "sync"

// AnnotationExpr represents a computed field annotation.
// Annotations add computed fields to query results.
type AnnotationExpr struct {
	Name       string
	Expr       QueryExpr
	Expression Expression
}

var (
	annotationRegistryMu sync.RWMutex
	annotationRegistry   = map[string]func(...interface{}) AnnotationExpr{}
)

// NewAnnotation creates a new annotation from a QueryExpr.
// nolint:gocritic // hugeParam: QueryExpr is small enough for value semantics
func NewAnnotation(name string, expr QueryExpr) AnnotationExpr {
	return AnnotationExpr{
		Name:       name,
		Expr:       expr,
		Expression: newQueryExprAdapter(expr),
	}
}

// NewExpressionAnnotation creates a new annotation from an Expression.
func NewExpressionAnnotation(name string, expr Expression) AnnotationExpr {
	return AnnotationExpr{
		Name:       name,
		Expression: expr,
	}
}

type queryExprAdapter struct {
	expr QueryExpr
}

func newQueryExprAdapter(expr QueryExpr) Expression {
	return queryExprAdapter{expr: expr}
}

func (a queryExprAdapter) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	if builder == nil {
		sql, args, _ := a.expr.ToSQL(1, defaultPlaceholder)
		return sql, args, nil
	}
	sql, args, nextIndex := a.expr.ToSQL(builder.paramIndex, builder.Placeholder)
	builder.paramIndex = nextIndex
	builder.args = append(builder.args, args...)
	return sql, args, nil
}

func (a queryExprAdapter) Resolve(schema *ModelSchema) error {
	return nil
}

// RegisterAnnotation registers a custom annotation type.
func RegisterAnnotation(name string, builder func(...interface{}) AnnotationExpr) {
	if builder == nil || normalizeRegistryName(name) == "" {
		return
	}

	annotationRegistryMu.Lock()
	defer annotationRegistryMu.Unlock()
	annotationRegistry[normalizeRegistryName(name)] = builder
}

// BuildAnnotation builds an annotation from a registered custom annotation name.
func BuildAnnotation(name string, args ...interface{}) (AnnotationExpr, bool) {
	annotationRegistryMu.RLock()
	builder, ok := annotationRegistry[normalizeRegistryName(name)]
	annotationRegistryMu.RUnlock()
	if !ok {
		return AnnotationExpr{}, false
	}

	annotation := builder(args...)
	if annotation.Name == "" {
		annotation.Name = normalizeRegistryName(name)
	}
	if annotation.Expression == nil && (annotation.Expr.field != "" || len(annotation.Expr.children) > 0 || annotation.Expr.combiner != "") {
		annotation.Expression = newQueryExprAdapter(annotation.Expr)
	}
	return annotation, true
}
