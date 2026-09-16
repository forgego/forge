package orm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	forgeerrors "github.com/forgego/forge/errors"
)

var aggregateChainError = forgeerrors.NewNotImplementedError("QuerySet.Aggregate chained into a row query; use AggregateValues")

func isAggregateChainError(err error) bool {
	return errors.Is(err, aggregateChainError)
}

func canReplaceDeferredError(err error) bool {
	return err == nil || isAggregateChainError(err)
}

// QuerySet is the type-safe QuerySet interface
type QuerySet[T any] interface {
	// SetDB sets the database connection
	SetDB(db interface{}) QuerySet[T]
	// Filtering - accepts new Expression interface
	Filter(expr Expression) QuerySet[T]
	Exclude(expr Expression) QuerySet[T]

	// Ordering
	OrderBy(fields ...any) QuerySet[T]
	Reverse() QuerySet[T]

	// Limiting
	Limit(n int) QuerySet[T]
	Offset(n int) QuerySet[T]
	Distinct(fields ...any) QuerySet[T]

	// Field Selection
	Select(fields ...any) QuerySet[T]
	Only(fields ...any) QuerySet[T]
	Defer(fields ...any) QuerySet[T]

	// Relations
	SelectRelated(relations ...any) QuerySet[T]
	PrefetchRelated(relations ...any) QuerySet[T]

	// Aggregation
	Aggregate(aggs ...Aggregate) QuerySet[T]
	Annotate(anns ...AnnotationExpr) QuerySet[T]

	// Values - type-safe return types
	Values(fields ...any) ValuesQuerySet[T]
	ValuesList(fields ...any) ValuesListQuerySet[T]

	// Project - type-safe projection to a different type
	// Note: Use Project[T, P](qs) function instead of method

	// Execution
	All(ctx context.Context) ([]*T, error)
	Get(ctx context.Context) (*T, error)
	First(ctx context.Context) (*T, error)
	Last(ctx context.Context) (*T, error)
	Count(ctx context.Context) (int64, error)
	Exists(ctx context.Context) (bool, error)

	// Updates - type-safe
	Update(ctx context.Context, updates UpdateMap) (int64, error)
	UpdateBuilder() (*UpdateBuilder[T], error)
	BulkUpdate(ctx context.Context, updates []UpdateMap) error

	// Deletes
	Delete(ctx context.Context) (int64, error)

	// Set Operations
	Union(other QuerySet[T]) QuerySet[T]
	Intersection(other QuerySet[T]) QuerySet[T]
	Difference(other QuerySet[T]) QuerySet[T]
}

// OrderField represents an ordering field with direction
type OrderField struct {
	Field     string
	Ascending bool
}

// GetFieldPath returns the field path for ordering
func (o OrderField) GetFieldPath() string {
	return o.Field
}

// IsAscending returns whether the ordering is ascending
func (o OrderField) IsAscending() bool {
	return o.Ascending
}

// Asc creates an ascending order field
func Asc(field string) OrderField {
	return OrderField{
		Field:     field,
		Ascending: true,
	}
}

// Desc creates a descending order field
func Desc(field string) OrderField {
	return OrderField{
		Field:     field,
		Ascending: false,
	}
}

// NewOrderField creates an order field with explicit direction.
//
// Deprecated: Use Asc(field) or Desc(field) instead for clarity. NewOrderField will be removed in v3.0.
// Migration:
//
//	// Old
//	order := orm.NewOrderField("created_at", true)
//	order := orm.NewOrderField("created_at", false)
//	// New
//	order := orm.Asc("created_at")
//	order := orm.Desc("created_at")
func NewOrderField(field string, ascending bool) OrderField {
	if ascending {
		return Asc(field)
	}
	return Desc(field)
}

// BaseQuerySet is the implementation
type BaseQuerySet[T any] struct {
	table           string
	schema          *ModelSchema
	conditions      []Expression
	excludes        []Expression
	orderBy         []OrderField
	limitVal        *int
	offsetVal       *int
	distinctFields  []string
	selectFields    []string
	onlyFields      []string
	deferFields     []string
	selectRelated   []string
	prefetchRelated []string
	preloaded       map[string]bool // Track which relations are preloaded (N+1 prevention)
	joins           []string
	joinMap         map[string]bool
	aggregates      []Aggregate
	annotations     []AnnotationExpr // Using existing AnnotationExpr type
	db              interface{}      // *db.DB
	err             error            // Deferred error from Filter/Exclude validation (checked at execution time)
}

// NewQuerySet creates a new QuerySet
func NewQuerySet[T any](tableName string) (QuerySet[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	if tableName == "" {
		tableName = schema.TableName
	}

	return &BaseQuerySet[T]{
		table:      tableName,
		schema:     schema,
		conditions: []Expression{},
		excludes:   []Expression{},
		orderBy:    []OrderField{},
		preloaded:  make(map[string]bool),
		joins:      []string{},
		joinMap:    make(map[string]bool),
	}, nil
}

// SetDB sets the database connection
func (qs *BaseQuerySet[T]) SetDB(db interface{}) QuerySet[T] {
	qs.db = db
	return qs
}

// getDB retrieves the database connection
func (qs *BaseQuerySet[T]) getDB(ctx context.Context) (DBTX, error) {
	if qs.db != nil {
		return GetDBTX(qs.db)
	}
	return nil, fmt.Errorf("database connection not set on QuerySet")
}

// getDialect retrieves the SQL dialect from the database connection
func (qs *BaseQuerySet[T]) getDialect() (interface {
	BuildPlaceholders(n int) string
}, error) {
	if qs.db != nil {
		return GetDialect(qs.db)
	}
	return nil, fmt.Errorf("database connection not set on QuerySet")
}

// newSQLBuilder creates an SQLBuilder configured with the queryset's dialect if available.
func (qs *BaseQuerySet[T]) newSQLBuilder() *SQLBuilder {
	if qs != nil && qs.db != nil {
		if d, err := GetDialect(qs.db); err == nil && d != nil {
			return NewSQLBuilderWithDialect(d)
		}
	}
	return NewSQLBuilder()
}

// clone creates a deep copy
func (qs *BaseQuerySet[T]) clone() *BaseQuerySet[T] {
	clone := &BaseQuerySet[T]{
		table:           qs.table,
		schema:          qs.schema,
		conditions:      append([]Expression{}, qs.conditions...),
		excludes:        append([]Expression{}, qs.excludes...),
		orderBy:         append([]OrderField{}, qs.orderBy...),
		limitVal:        qs.limitVal,
		offsetVal:       qs.offsetVal,
		distinctFields:  append([]string{}, qs.distinctFields...),
		selectFields:    append([]string{}, qs.selectFields...),
		onlyFields:      append([]string{}, qs.onlyFields...),
		deferFields:     append([]string{}, qs.deferFields...),
		selectRelated:   append([]string{}, qs.selectRelated...),
		prefetchRelated: append([]string{}, qs.prefetchRelated...),
		preloaded:       make(map[string]bool),
		joins:           append([]string{}, qs.joins...),
		joinMap:         make(map[string]bool),
		aggregates:      append([]Aggregate{}, qs.aggregates...),
		annotations:     append([]AnnotationExpr{}, qs.annotations...),
		db:              qs.db,
		err:             qs.err,
	}
	// Copy preloaded map
	for k, v := range qs.preloaded {
		clone.preloaded[k] = v
	}
	// Copy joinMap
	for k, v := range qs.joinMap {
		clone.joinMap[k] = v
	}
	return clone
}

// Filter adds a filter condition.
// Validation errors are deferred and returned when the QuerySet is executed
// (e.g. via All, Get, Count). This avoids panics while preserving the
// chainable API.
func (qs *BaseQuerySet[T]) Filter(expr Expression) QuerySet[T] {
	// Check for nil queryset or schema
	if qs == nil || qs.schema == nil {
		clone := qs.clone()
		clone.err = fmt.Errorf("cannot filter: queryset or schema is nil")
		return clone
	}

	// Validate expression -- store error instead of panicking
	if err := expr.Resolve(qs.schema); err != nil {
		clone := qs.clone()
		if canReplaceDeferredError(clone.err) {
			clone.err = fmt.Errorf("invalid filter expression: %w", err)
		}
		return clone
	}

	clone := qs.clone()
	clone.conditions = append(clone.conditions, expr)
	return clone
}

// Exclude adds an exclude condition.
// Validation errors are deferred and returned when the QuerySet is executed.
func (qs *BaseQuerySet[T]) Exclude(expr Expression) QuerySet[T] {
	// Validate expression -- store error instead of panicking
	if err := expr.Resolve(qs.schema); err != nil {
		clone := qs.clone()
		if canReplaceDeferredError(clone.err) {
			clone.err = fmt.Errorf("invalid exclude expression: %w", err)
		}
		return clone
	}

	clone := qs.clone()
	clone.excludes = append(clone.excludes, expr)
	return clone
}

// OrderBy sets ordering - accepts both OrderField (string) and OrderFieldExpr[T] (type-safe)
func (qs *BaseQuerySet[T]) OrderBy(fields ...any) QuerySet[T] {
	clone := qs.clone()
	for _, field := range fields {
		orderField := OrderField{
			Field:     extractOrderFieldPath(field),
			Ascending: extractOrderFieldAscending(field),
		}
		clone.orderBy = append(clone.orderBy, orderField)
	}
	return clone
}

// extractOrderFieldPath extracts the field path from an OrderFieldSpec
func extractOrderFieldPath(field any) string {
	// Use type assertion to OrderFieldSpec interface
	if spec, ok := field.(OrderFieldSpec); ok {
		return spec.GetFieldPath()
	}
	// Allow FieldPath (e.g. orm.Field[T]) directly
	if fp, ok := field.(FieldPath); ok {
		return fp.Path()
	}
	// Allow plain string field names, including "-field" prefix
	if s, ok := field.(string); ok {
		return strings.TrimPrefix(s, "-")
	}
	return ""
}

// extractOrderFieldAscending extracts the ascending flag from an OrderFieldSpec
func extractOrderFieldAscending(field any) bool {
	// Use type assertion to OrderFieldSpec interface
	if spec, ok := field.(OrderFieldSpec); ok {
		return spec.IsAscending()
	}
	// FieldPath defaults to ascending
	if _, ok := field.(FieldPath); ok {
		return true
	}
	// String supports "-" prefix for descending
	if s, ok := field.(string); ok {
		return !strings.HasPrefix(s, "-")
	}
	return true
}

// Reverse reverses the current ordering
func (qs *BaseQuerySet[T]) Reverse() QuerySet[T] {
	clone := qs.clone()
	for i := range clone.orderBy {
		clone.orderBy[i].Ascending = !clone.orderBy[i].Ascending
	}
	return clone
}

// Limit sets the limit
func (qs *BaseQuerySet[T]) Limit(n int) QuerySet[T] {
	clone := qs.clone()
	clone.limitVal = &n
	return clone
}

// Offset sets the offset
func (qs *BaseQuerySet[T]) Offset(n int) QuerySet[T] {
	clone := qs.clone()
	clone.offsetVal = &n
	return clone
}

// Distinct sets distinct fields - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Distinct(fields ...any) QuerySet[T] {
	clone := qs.clone()
	if len(fields) > 0 {
		paths := make([]string, len(fields))
		for i, field := range fields {
			paths[i] = ExtractPathFromAny(field)
		}
		clone.distinctFields = paths
	} else {
		clone.distinctFields = []string{"*"}
	}
	return clone
}

// Select sets fields to select - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Select(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return clone
}

// Only sets fields to only load - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Only(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.onlyFields = paths
	return clone
}

// Defer sets fields to defer loading - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Defer(fields ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.deferFields = paths
	return clone
}

// SelectRelated adds relations to select - accepts both string and RelationExpression
func (qs *BaseQuerySet[T]) SelectRelated(relations ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(relations))
	for i, relation := range relations {
		paths[i] = extractRelationPathFromAny(relation)
	}
	clone.selectRelated = append(clone.selectRelated, paths...)
	return clone
}

// PrefetchRelated adds relations to prefetch - accepts both string and RelationExpression
// Marks relations as preloaded to prevent N+1 queries
func (qs *BaseQuerySet[T]) PrefetchRelated(relations ...any) QuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(relations))
	for i, relation := range relations {
		path := extractRelationPathFromAny(relation)
		paths[i] = path
		// Mark as preloaded
		clone.preloaded[path] = true
	}
	clone.prefetchRelated = append(clone.prefetchRelated, paths...)
	return clone
}

// Aggregate adds aggregate expressions
func (qs *BaseQuerySet[T]) Aggregate(aggs ...Aggregate) QuerySet[T] {
	clone := qs.clone()
	clone.aggregates = append(clone.aggregates, aggs...)
	if canReplaceDeferredError(clone.err) {
		clone.err = aggregateChainError
	}
	return clone
}

// Annotate adds annotation expressions
func (qs *BaseQuerySet[T]) Annotate(anns ...AnnotationExpr) QuerySet[T] {
	clone := qs.clone()
	clone.annotations = append(clone.annotations, anns...)
	return clone
}

// Values returns a ValuesQuerySet - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) Values(fields ...any) ValuesQuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return &BaseValuesQuerySet[T]{base: clone}
}

// ValuesList returns a ValuesListQuerySet - accepts both string and FieldExpression[T]
func (qs *BaseQuerySet[T]) ValuesList(fields ...any) ValuesListQuerySet[T] {
	clone := qs.clone()
	paths := make([]string, len(fields))
	for i, field := range fields {
		paths[i] = ExtractPathFromAny(field)
	}
	clone.selectFields = paths
	return &BaseValuesListQuerySet[T]{base: clone}
}

// Project is implemented via the Project function in projection.go
// Usage: Project[User, UserProjection](qs)

// UpdateBuilder returns an UpdateBuilder
func (qs *BaseQuerySet[T]) UpdateBuilder() (*UpdateBuilder[T], error) {
	// Create a wrapper that implements the interface needed by UpdateBuilder
	return NewUpdateBuilderFromQuerySet(qs)
}

// Union performs a UNION operation
func (qs *BaseQuerySet[T]) Union(other QuerySet[T]) QuerySet[T] {
	clone := qs.clone()
	if canReplaceDeferredError(clone.err) {
		clone.err = forgeerrors.NewNotImplementedError("QuerySet.Union")
	}
	return clone
}

// Intersection performs an INTERSECT operation
func (qs *BaseQuerySet[T]) Intersection(other QuerySet[T]) QuerySet[T] {
	clone := qs.clone()
	if canReplaceDeferredError(clone.err) {
		clone.err = forgeerrors.NewNotImplementedError("QuerySet.Intersection")
	}
	return clone
}

// Difference performs an EXCEPT operation
func (qs *BaseQuerySet[T]) Difference(other QuerySet[T]) QuerySet[T] {
	clone := qs.clone()
	if canReplaceDeferredError(clone.err) {
		clone.err = forgeerrors.NewNotImplementedError("QuerySet.Difference")
	}
	return clone
}
