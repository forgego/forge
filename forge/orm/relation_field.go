package orm

import (
	"fmt"
	"reflect"
)

// RelationField represents a type-safe relation field that can be traversed
// T is the target model type, D is the max depth (for cycle prevention)
type RelationField[T any, D Depth] struct {
	relationName string
	relationType RelationType
	targetSchema *ModelSchema
	sourceSchema *ModelSchema
	table        string
	depth        int
}

// Depth is a type-level depth counter for compile-time cycle prevention
type Depth interface {
	~int
}

// MaxDepth is the maximum allowed relation depth
const MaxDepth = 5

// NewRelationField creates a new relation field
func NewRelationField[T any, D Depth](
	relationName string,
	relationType RelationType,
	sourceSchema, targetSchema *ModelSchema,
	table string,
) RelationField[T, D] {
	return RelationField[T, D]{
		relationName: relationName,
		relationType: relationType,
		targetSchema: targetSchema,
		sourceSchema: sourceSchema,
		table:        table,
	}
}

// Fields returns the field expressions for the related model
// This allows chaining: Post.Fields.Author.Fields.Email
func (rf RelationField[T, D]) Fields() *FieldAccessor[T] {
	accessor, err := NewFieldAccessor[T]()
	if err != nil {
		panic(fmt.Sprintf("failed to create field accessor for relation %s: %v", rf.relationName, err))
	}
	return accessor
}

// Path returns the relation path for SQL generation
func (rf RelationField[T, D]) Path() string {
	return rf.relationName
}

// Table returns the target table name
func (rf RelationField[T, D]) Table() string {
	return rf.table
}

// TraverseField creates a field expression on the related model
// This is used for queries like: Post.Fields.Author.Fields.Email.Eq("test@example.com")
func TraverseField[T any, D Depth, V any](rf RelationField[T, D], fieldName string) FieldExpression[V] {
	// Validate field exists on target model
	fieldInfo := rf.targetSchema.GetField(fieldName)
	if fieldInfo == nil {
		panic(fmt.Sprintf("field %s not found on related model %s", fieldName, rf.targetSchema.TableName))
	}

	// Build relation path: "author__email"
	path := fmt.Sprintf("%s__%s", rf.relationName, fieldName)

	return NewField[V](path, rf.table)
}

// Resolve validates the relation exists in the source schema
func (rf RelationField[T, D]) Resolve(schema *ModelSchema) error {
	relation := schema.GetRelation(rf.relationName)
	if relation == nil {
		return fmt.Errorf("relation %s not found in model", rf.relationName)
	}

	if relation.Type != rf.relationType {
		return fmt.Errorf("relation %s has type %v, expected %v", rf.relationName, relation.Type, rf.relationType)
	}

	return nil
}

// GetDepth returns the current depth
func (rf RelationField[T, D]) GetDepth() int {
	return rf.depth
}

// CheckDepth validates depth hasn't exceeded maximum
func (rf RelationField[T, D]) CheckDepth() error {
	if rf.depth >= MaxDepth {
		return fmt.Errorf("relation depth %d exceeds maximum %d", rf.depth, MaxDepth)
	}
	return nil
}

// TraversalContext tracks visited nodes to prevent cycles at runtime
type TraversalContext struct {
	visited map[visitKey]bool
	depth   int
	maxDepth int
}

type visitKey struct {
	modelType reflect.Type
	id        int64
}

// NewTraversalContext creates a new traversal context
func NewTraversalContext(maxDepth int) *TraversalContext {
	return &TraversalContext{
		visited:  make(map[visitKey]bool),
		depth:    0,
		maxDepth: maxDepth,
	}
}

// Visit marks a model as visited and checks for cycles
func (ctx *TraversalContext) Visit(model interface{}) error {
	key := visitKey{
		modelType: reflect.TypeOf(model),
		id:        getIDFromModel(model),
	}

	if ctx.visited[key] {
		return fmt.Errorf("cycle detected: model %v already visited", key)
	}

	if ctx.depth >= ctx.maxDepth {
		return fmt.Errorf("max depth %d exceeded", ctx.maxDepth)
	}

	ctx.visited[key] = true
	ctx.depth++
	return nil
}

// Leave marks a model as no longer visited (for backtracking)
func (ctx *TraversalContext) Leave(model interface{}) {
	key := visitKey{
		modelType: reflect.TypeOf(model),
		id:        getIDFromModel(model),
	}
	delete(ctx.visited, key)
	ctx.depth--
}

// getIDFromModel extracts the ID from a model instance
func getIDFromModel(model interface{}) int64 {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	idField := v.FieldByName("ID")
	if !idField.IsValid() {
		return 0
	}

	if idField.Kind() == reflect.Int64 {
		return idField.Int()
	}

	return 0
}

// ErrCycleDetected is returned when a cycle is detected in relation traversal
var ErrCycleDetected = fmt.Errorf("cycle detected in relation traversal")

// ErrMaxDepthExceeded is returned when max depth is exceeded
var ErrMaxDepthExceeded = fmt.Errorf("max relation depth exceeded")
