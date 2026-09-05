package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// UpdateMap is a type-safe map for updates
type UpdateMap map[string]interface{}

// UpdateBuilder provides type-safe update operations
type UpdateBuilder[T any] struct {
	qs      updatableQuerySet[T]
	schema  *ModelSchema
	updates map[string]interface{}
	err     error // Deferred error from Set/SetExpr/etc. (checked at Execute time)
}

// updatableQuerySet is an interface for QuerySets that support updates
type updatableQuerySet[T any] interface {
	Update(ctx context.Context, updates UpdateMap) (int64, error)
}

// NewUpdateBuilder creates an update builder from an updatable QuerySet
func NewUpdateBuilder[T any](qs updatableQuerySet[T]) (*UpdateBuilder[T], error) {
	schema, err := GetModelSchema[T]()
	if err != nil {
		return nil, err
	}

	return &UpdateBuilder[T]{
		qs:      qs,
		schema:  schema,
		updates: make(map[string]interface{}),
	}, nil
}

// Set sets a field value with type checking (dynamic API).
// Validation errors are deferred and returned when Execute() is called.
func (ub *UpdateBuilder[T]) Set(fieldName string, value interface{}) *UpdateBuilder[T] {
	// Validate field exists and type matches
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	expectedType := fieldInfo.Type
	actualType := reflect.TypeOf(value)

	// Check if types are assignable
	if !actualType.AssignableTo(expectedType) {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s expects %v, got %v", fieldName, expectedType, actualType)
		}
		return ub
	}

	ub.updates[fieldName] = value
	return ub
}

// SetFieldValue sets a field value using a type-safe FieldExpression (type-safe API).
// This provides compile-time type checking.
// Usage: SetFieldValue(ub, User.Fields.Email, "new@example.com")
func SetFieldValue[T any, V any](ub *UpdateBuilder[T], field FieldExpression[V], value V) *UpdateBuilder[T] {
	// Validate field exists in schema
	fieldInfo := ub.schema.GetField(field.Path())
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", field.Path())
		}
		return ub
	}

	// Type is already validated by FieldExpression[V] generic constraint
	ub.updates[field.Path()] = value
	return ub
}

// SetExpr sets a field to an expression value.
// Validation errors are deferred and returned when Execute() is called.
func (ub *UpdateBuilder[T]) SetExpr(fieldName string, expr Expression) *UpdateBuilder[T] {
	// Validate field exists
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	// Validate expression against schema
	if err := expr.Resolve(ub.schema); err != nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("invalid expression for field %s: %w", fieldName, err)
		}
		return ub
	}

	ub.updates[fieldName] = expr
	return ub
}

// SetField sets a field to another field's value.
// Validation errors are deferred and returned when Execute() is called.
func (ub *UpdateBuilder[T]) SetField(fieldName string, sourceField Expression) *UpdateBuilder[T] {
	// Validate field exists
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	// Validate source field
	if err := sourceField.Resolve(ub.schema); err != nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("invalid source field: %w", err)
		}
		return ub
	}

	ub.updates[fieldName] = sourceField
	return ub
}

// Execute executes the update.
// Returns any deferred validation errors from Set/SetExpr/SetField calls.
func (ub *UpdateBuilder[T]) Execute(ctx context.Context) (int64, error) {
	if ub.err != nil {
		return 0, ub.err
	}

	// Convert updates map to the format expected by QuerySet.Update
	// Handle expressions specially
	updateMap := make(map[string]interface{})

	for fieldName, value := range ub.updates {
		// If it's an expression, we need to handle it in SQL generation
		if expr, ok := value.(Expression); ok {
			// Store expression for SQL builder to handle
			updateMap[fieldName] = expr
		} else {
			updateMap[fieldName] = value
		}
	}

	// Use the QuerySet's Update method
	// Note: This requires QuerySet.Update to handle Expression values
	return ub.qs.Update(ctx, updateMap)
}

// Increment increments a numeric field.
// Validation errors are deferred and returned when Execute() is called.
func (ub *UpdateBuilder[T]) Increment(fieldName string, amount interface{}) *UpdateBuilder[T] {
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	// Create a raw SQL expression for field + value
	// This bypasses type checking issues with CombinedExpression
	// Format: "field" + $1
	fieldSQL := EscapeIdentifier(fieldName)
	placeholder := fmt.Sprintf("$%d", 1) // Will be replaced by SQL builder

	// Store as a special expression that represents field + value
	// We'll handle this in the SQL generation
	rawExpr := &RawExpression{
		SQL:  fmt.Sprintf("%s + %s", fieldSQL, placeholder),
		Args: []interface{}{amount},
	}

	ub.updates[fieldName] = rawExpr
	return ub
}

// RawExpression represents a raw SQL expression
type RawExpression struct {
	SQL  string
	Args []interface{}
}

// ToSQL converts raw expression to SQL
func (r *RawExpression) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	// Replace placeholders with actual parameter placeholders
	sql := r.SQL
	args := []interface{}{}

	for i, arg := range r.Args {
		placeholder := builder.AddArg(arg)
		sql = strings.Replace(sql, fmt.Sprintf("$%d", i+1), placeholder, 1)
		args = append(args, arg)
	}

	return sql, args, nil
}

// Resolve validates the raw expression (always succeeds for now)
func (r *RawExpression) Resolve(schema *ModelSchema) error {
	return nil
}

// Decrement decrements a numeric field.
// Validation errors are deferred and returned when Execute() is called.
func (ub *UpdateBuilder[T]) Decrement(fieldName string, amount interface{}) *UpdateBuilder[T] {
	fieldInfo := ub.schema.GetField(fieldName)
	if fieldInfo == nil {
		if ub.err == nil {
			ub.err = fmt.Errorf("field %s not found on model", fieldName)
		}
		return ub
	}

	// Create a raw SQL expression for field - value
	fieldSQL := EscapeIdentifier(fieldName)
	placeholder := fmt.Sprintf("$%d", 1)

	rawExpr := &RawExpression{
		SQL:  fmt.Sprintf("%s - %s", fieldSQL, placeholder),
		Args: []interface{}{amount},
	}

	ub.updates[fieldName] = rawExpr
	return ub
}

// Numeric constraint for numeric types
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}



