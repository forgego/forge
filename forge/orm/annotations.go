package orm

import "sync"

// AnnotationExpr represents a computed field annotation
// Annotations add computed fields to query results
type AnnotationExpr struct {
	Name string
	Expr QueryExpr
}

var (
	annotationRegistryMu sync.RWMutex
	annotationRegistry   = map[string]func(...interface{}) AnnotationExpr{}
)

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
	return annotation, true
}
