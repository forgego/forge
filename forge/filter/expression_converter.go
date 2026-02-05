package filter

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/orm"
)

// ExpressionConverter converts filter AST nodes to ORM expressions
type ExpressionConverter[T any] struct {
	schema *orm.ModelSchema
}

// NewExpressionConverter creates a new expression converter
func NewExpressionConverter[T any](schema *orm.ModelSchema) *ExpressionConverter[T] {
	return &ExpressionConverter[T]{
		schema: schema,
	}
}

// ConvertNode converts a FilterNode to an ORM Expression
func (ec *ExpressionConverter[T]) ConvertNode(node *FilterNode) (orm.Expression, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot convert nil node")
	}

	switch node.Op {
	case OpField:
		return ec.convertFieldNode(node)

	case OpAnd:
		return ec.convertAndNode(node)

	case OpOr:
		return ec.convertOrNode(node)

	case OpNot:
		return ec.convertNotNode(node)

	default:
		return nil, fmt.Errorf("unknown filter operation: %s", node.Op)
	}
}

// convertFieldNode converts a field filter node to a comparison expression
func (ec *ExpressionConverter[T]) convertFieldNode(node *FilterNode) (orm.Expression, error) {
	if node.Field == "" {
		return nil, fmt.Errorf("field node must have a field path")
	}

	// Resolve field path to get field info
	fieldInfo, targetSchema, err := ec.schema.ResolvePath(node.Field)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve field path '%s': %w", node.Field, err)
	}
	
	// Use target schema for table name if we resolved through relations
	tableName := ec.schema.TableName
	if targetSchema != nil {
		tableName = targetSchema.TableName
	}

	// Create field expression based on field type with correct table
	fieldExpr := ec.createFieldExpressionWithTable(node.Field, fieldInfo.Type, tableName)

	// Convert lookup to operator
	op, err := ec.lookupToOperator(node.Lookup, fieldInfo.Type)
	if err != nil {
		return nil, err
	}

	// Create comparison expression
	comparison := ec.createComparisonExpression(fieldExpr, op, node.Value, node.Lookup)
	return comparison, nil
}

// createFieldExpression creates a FieldExpression based on field type
func (ec *ExpressionConverter[T]) createFieldExpression(fieldPath string, fieldType reflect.Type) orm.Expression {
	return ec.createFieldExpressionWithTable(fieldPath, fieldType, ec.schema.TableName)
}

// createFieldExpressionWithTable creates a FieldExpression with a specific table name
func (ec *ExpressionConverter[T]) createFieldExpressionWithTable(fieldPath string, fieldType reflect.Type, tableName string) orm.Expression {
	// Use reflection to create the appropriate FieldExpression type
	// This is a simplified version - in practice, you'd need to handle all types
	switch fieldType.Kind() {
	case reflect.String:
		return orm.NewField[string](fieldPath, tableName)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return orm.NewField[int64](fieldPath, tableName)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return orm.NewField[uint64](fieldPath, tableName)
	case reflect.Float32, reflect.Float64:
		return orm.NewField[float64](fieldPath, tableName)
	case reflect.Bool:
		return orm.NewField[bool](fieldPath, tableName)
	default:
		// For unknown types, use interface{} as fallback
		// This will work but loses type safety
		return orm.NewField[interface{}](fieldPath, tableName)
	}
}

// lookupToOperator converts a filter lookup to an ORM operator
func (ec *ExpressionConverter[T]) lookupToOperator(lookup string, fieldType reflect.Type) (orm.Operator, error) {
	switch lookup {
	case "exact":
		return orm.OpEquals, nil
	case "iexact":
		if fieldType.Kind() == reflect.String {
			return orm.OpIExact, nil
		}
		return orm.OpEquals, nil // Fallback for non-string
	case "contains":
		if fieldType.Kind() == reflect.String {
			return orm.OpContains, nil
		}
		return "", fmt.Errorf("contains lookup only valid for string fields")
	case "icontains":
		if fieldType.Kind() == reflect.String {
			return orm.OpIContains, nil
		}
		return "", fmt.Errorf("icontains lookup only valid for string fields")
	case "startswith":
		if fieldType.Kind() == reflect.String {
			return orm.OpStartsWith, nil
		}
		return "", fmt.Errorf("startswith lookup only valid for string fields")
	case "istartswith":
		if fieldType.Kind() == reflect.String {
			return orm.OpStartsWith, nil // Use StartsWith, dialect adapter will handle case
		}
		return "", fmt.Errorf("istartswith lookup only valid for string fields")
	case "endswith":
		if fieldType.Kind() == reflect.String {
			return orm.OpEndsWith, nil
		}
		return "", fmt.Errorf("endswith lookup only valid for string fields")
	case "iendswith":
		if fieldType.Kind() == reflect.String {
			return orm.OpEndsWith, nil // Use EndsWith, dialect adapter will handle case
		}
		return "", fmt.Errorf("iendswith lookup only valid for string fields")
	case "in":
		return orm.OpIn, nil
	case "gt":
		return orm.OpGreater, nil
	case "gte":
		return orm.OpGreaterOrEqual, nil
	case "lt":
		return orm.OpLess, nil
	case "lte":
		return orm.OpLessOrEqual, nil
	case "range":
		return orm.OpRange, nil
	case "isnull":
		return orm.OpIsNull, nil
	case "isnotnull":
		return orm.OpIsNotNull, nil
	case "year":
		return orm.OpYear, nil
	case "month":
		return orm.OpMonth, nil
	case "day":
		return orm.OpDay, nil
	default:
		return "", fmt.Errorf("unknown lookup: %s", lookup)
	}
}

// createComparisonExpression creates a ComparisonExpression
func (ec *ExpressionConverter[T]) createComparisonExpression(
	fieldExpr orm.Expression,
	op orm.Operator,
	value interface{},
	lookup string,
) orm.Expression {
	// This is tricky because ComparisonExpression is generic
	// We need to use type assertion or reflection to create the right type
	// For now, we'll use a helper that works with the actual field type

	// Try to get the field expression's type
	switch fe := fieldExpr.(type) {
	case orm.FieldExpression[string]:
		return orm.ComparisonExpression[string]{
			Field: fe,
			Op:    op,
			Value: value,
		}
	case orm.FieldExpression[int64]:
		return orm.ComparisonExpression[int64]{
			Field: fe,
			Op:    op,
			Value: value,
		}
	case orm.FieldExpression[float64]:
		return orm.ComparisonExpression[float64]{
			Field: fe,
			Op:    op,
			Value: value,
		}
	case orm.FieldExpression[bool]:
		return orm.ComparisonExpression[bool]{
			Field: fe,
			Op:    op,
			Value: value,
		}
	case orm.FieldExpression[interface{}]:
		return orm.ComparisonExpression[interface{}]{
			Field: fe,
			Op:    op,
			Value: value,
		}
	default:
		// Fallback: try to create with interface{} type
		// This requires the fieldExpr to be FieldExpression[interface{}]
		if fe, ok := fieldExpr.(orm.FieldExpression[interface{}]); ok {
			return orm.ComparisonExpression[interface{}]{
				Field: fe,
				Op:    op,
				Value: value,
			}
		}
		// Last resort: create a new field expression with interface{} type
		// This loses type safety but allows the expression to work
		// We need to extract the field path - for now use empty and let Resolve handle it
		feNew := orm.NewField[interface{}]("", ec.schema.TableName)
		return orm.ComparisonExpression[interface{}]{
			Field: feNew,
			Op:    op,
			Value: value,
		}
	}
}

// convertAndNode converts an AND node to ORM Q expression
func (ec *ExpressionConverter[T]) convertAndNode(node *FilterNode) (orm.Expression, error) {
	if len(node.Children) < 2 {
		return nil, fmt.Errorf("AND node must have at least 2 children")
	}

	// Convert all children
	var q *orm.Q
	for i, child := range node.Children {
		expr, err := ec.ConvertNode(child)
		if err != nil {
			return nil, fmt.Errorf("child %d: %w", i, err)
		}

		if i == 0 {
			q = orm.NewQ(expr)
		} else {
			q = q.And(orm.NewQ(expr))
		}
	}

	return q, nil
}

// convertOrNode converts an OR node to ORM Q expression
func (ec *ExpressionConverter[T]) convertOrNode(node *FilterNode) (orm.Expression, error) {
	if len(node.Children) < 2 {
		return nil, fmt.Errorf("OR node must have at least 2 children")
	}

	// Convert all children
	var q *orm.Q
	for i, child := range node.Children {
		expr, err := ec.ConvertNode(child)
		if err != nil {
			return nil, fmt.Errorf("child %d: %w", i, err)
		}

		if i == 0 {
			q = orm.NewQ(expr)
		} else {
			q = q.Or(orm.NewQ(expr))
		}
	}

	return q, nil
}

// convertNotNode converts a NOT node to ORM Q expression
func (ec *ExpressionConverter[T]) convertNotNode(node *FilterNode) (orm.Expression, error) {
	if len(node.Children) != 1 {
		return nil, fmt.Errorf("NOT node must have exactly 1 child")
	}

	expr, err := ec.ConvertNode(node.Children[0])
	if err != nil {
		return nil, fmt.Errorf("NOT child: %w", err)
	}

	return orm.NewQ(expr).Not(), nil
}
