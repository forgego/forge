package filter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/forgego/forge/orm"
)

// FilterSet is the main entry point for filtering
type FilterSet[T any] struct {
	schema    *orm.ModelSchema
	filters   map[string]Filter[T]
	ast       *FilterNode
	security  *SecurityConfig
	optimizer *QueryOptimizer
	queryset  orm.QuerySet[T]
}

// GetSecurityConfig returns the security configuration
func (fs *FilterSet[T]) GetSecurityConfig() *SecurityConfig {
	return fs.security
}

// NewFilterSet creates a new FilterSet for a model type
func NewFilterSet[T any]() (*FilterSet[T], error) {
	schema, err := orm.GetModelSchema[T]()
	if err != nil {
		return nil, fmt.Errorf("failed to get model schema: %w", err)
	}

	return &FilterSet[T]{
		schema:    schema,
		filters:   make(map[string]Filter[T]),
		security:  NewSecurityConfig(),
		optimizer: NewQueryOptimizer(),
	}, nil
}

// WithQueryset sets the base queryset for filtering
func (fs *FilterSet[T]) WithQueryset(qs orm.QuerySet[T]) *FilterSet[T] {
	fs.queryset = qs
	return fs
}

// WithSecurity sets the security configuration
func (fs *FilterSet[T]) WithSecurity(security *SecurityConfig) *FilterSet[T] {
	fs.security = security
	return fs
}

// AddFilter adds a filter to the filterset
func (fs *FilterSet[T]) AddFilter(name string, filter Filter[T]) *FilterSet[T] {
	fs.filters[name] = filter
	return fs
}

// GetFilter gets a filter by name
func (fs *FilterSet[T]) GetFilter(name string) (Filter[T], bool) {
	filter, ok := fs.filters[name]
	return filter, ok
}

// Where starts a filter builder for a field path
func (fs *FilterSet[T]) Where(fieldPath string) *FilterBuilder[T] {
	// Validate path using schema
	if err := fs.schema.ValidatePath(fieldPath); err != nil {
		return &FilterBuilder[T]{
			fs:  fs,
			err: err,
		}
	}

	return &FilterBuilder[T]{
		fs:        fs,
		fieldPath: fieldPath,
	}
}

// ApplyAST applies a filter AST to the queryset
func (fs *FilterSet[T]) ApplyAST(ctx context.Context, ast *FilterNode) (orm.QuerySet[T], error) {
	if ast == nil {
		return fs.queryset, nil
	}

	if err := ast.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter AST: %w", err)
	}

	// Convert AST to ORM expression
	expr, err := fs.astToExpression(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AST to expression: %w", err)
	}

	// Apply to queryset
	qs := fs.queryset.Filter(expr)
	return qs, nil
}

// ApplyQueryParams applies query parameters from an HTTP request
func (fs *FilterSet[T]) ApplyQueryParams(ctx context.Context, r *http.Request) (orm.QuerySet[T], error) {
	if r == nil {
		return fs.queryset, nil
	}

	// Parse query parameters into AST
	parser := NewParser(WithSecurity(fs.security))
	ast, err := parser.ParseFilterNode(r, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query parameters: %w", err)
	}

	if ast == nil {
		return fs.queryset, nil
	}

	// Apply AST to queryset
	return fs.ApplyAST(ctx, ast)
}

// astToExpression converts a FilterNode AST to an ORM Expression
func (fs *FilterSet[T]) astToExpression(node *FilterNode) (orm.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot convert nil node to expression")
	}

	converter := NewExpressionConverter[T](fs.schema)
	return converter.ConvertNode(node)
}

// GetAST returns the current filter AST
func (fs *FilterSet[T]) GetAST() *FilterNode {
	return fs.ast
}

// SetAST sets the filter AST
func (fs *FilterSet[T]) SetAST(ast *FilterNode) *FilterSet[T] {
	fs.ast = ast
	return fs
}

// GetFilters returns all filters from the filter set
func (fs *FilterSet[T]) GetFilters() map[string]Filter[T] {
	return fs.filters
}

// Copy creates a copy of the filterset
func (fs *FilterSet[T]) Copy() *FilterSet[T] {
	copy := &FilterSet[T]{
		schema:    fs.schema,
		filters:   make(map[string]Filter[T]),
		security:  fs.security,
		optimizer: fs.optimizer,
		queryset:  fs.queryset,
	}

	// Copy filters
	for k, v := range fs.filters {
		copy.filters[k] = v
	}

	// Copy AST
	if fs.ast != nil {
		copy.ast = fs.ast.Clone()
	}

	return copy
}

// FilterBuilder provides a fluent API for building filters
type FilterBuilder[T any] struct {
	fs        *FilterSet[T]
	fieldPath string
	err       error
}

// Contains creates a contains filter
func (fb *FilterBuilder[T]) Contains(value string) *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "contains", value)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// Equals creates an exact match filter
func (fb *FilterBuilder[T]) Equals(value interface{}) *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "exact", value)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// In creates an IN filter
func (fb *FilterBuilder[T]) In(values ...interface{}) *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "in", values)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// Greater creates a greater than filter
func (fb *FilterBuilder[T]) Greater(value interface{}) *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "gt", value)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// Less creates a less than filter
func (fb *FilterBuilder[T]) Less(value interface{}) *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "lt", value)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// IsNull creates an IS NULL filter
func (fb *FilterBuilder[T]) IsNull() *FilterSet[T] {
	if fb.err != nil {
		return fb.fs
	}

	node := NewFieldNode(fb.fieldPath, "isnull", true)
	fb.fs.ast = combineWithAnd(fb.fs.ast, node)
	return fb.fs
}

// AndGroup creates an AND group
func (fs *FilterSet[T]) AndGroup(fn func(*QueryBuilder[T])) *FilterSet[T] {
	builder := NewQueryBuilder[T](fs)
	fn(builder)
	
	if len(builder.nodes) > 0 {
		andNode := NewAndNode(builder.nodes...)
		fs.ast = combineWithAnd(fs.ast, andNode)
	}
	
	return fs
}

// OrGroup creates an OR group
func (fs *FilterSet[T]) OrGroup(fn func(*QueryBuilder[T])) *FilterSet[T] {
	builder := NewQueryBuilder[T](fs)
	fn(builder)
	
	if len(builder.nodes) > 0 {
		orNode := NewOrNode(builder.nodes...)
		fs.ast = combineWithAnd(fs.ast, orNode)
	}
	
	return fs
}

// combineWithAnd combines a node with an existing AST using AND
func combineWithAnd(existing *FilterNode, new *FilterNode) *FilterNode {
	if existing == nil {
		return new
	}
	return NewAndNode(existing, new)
}

// QueryBuilder provides a builder for boolean tree composition
type QueryBuilder[T any] struct {
	fs    *FilterSet[T]
	nodes []*FilterNode
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder[T any](fs *FilterSet[T]) *QueryBuilder[T] {
	return &QueryBuilder[T]{
		fs:    fs,
		nodes: make([]*FilterNode, 0),
	}
}

// Where starts a filter for a field path
func (qb *QueryBuilder[T]) Where(fieldPath string) *FilterBuilder[T] {
	return qb.fs.Where(fieldPath)
}

// Filter is an alias for Where
func (qb *QueryBuilder[T]) Filter(fieldPath string) *FilterBuilder[T] {
	return qb.Where(fieldPath)
}

// OrFilter creates an OR filter
func (qb *QueryBuilder[T]) OrFilter(fieldPath string) *FilterBuilder[T] {
	// This would create a filter that gets added to an OR group
	// For now, just use Where
	return qb.Where(fieldPath)
}
