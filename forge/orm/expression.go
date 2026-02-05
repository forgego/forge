package orm

import (
	"fmt"
	"reflect"
	"strings"
)

// Expression is the base interface for all query expressions
type Expression interface {
	// ToSQL converts the expression to SQL with parameters
	ToSQL(builder *SQLBuilder) (string, []interface{}, error)
	// Resolve validates and resolves the expression against a schema
	Resolve(schema *ModelSchema) error
}

// Field represents a type-safe field reference (like Django's F)
// T is the type of the field value
// This is the primary type-safe field API for queries
type Field[T any] struct {
	fieldPath string // e.g., "price", "author__name"
	table     string
	fieldType reflect.Type
}

// FieldExpression is an alias for Field for backward compatibility
// Deprecated: Use Field instead. FieldExpression will be removed in v2.0.
type FieldExpression[T any] = Field[T]

// NewField creates a type-safe field expression
func NewField[T any](path, table string) Field[T] {
	return Field[T]{
		fieldPath: path,
		table:     table,
		fieldType: reflect.TypeOf((*T)(nil)).Elem(),
	}
}

// Path returns the field path
func (f Field[T]) Path() string {
	return f.fieldPath
}

// Table returns the table name
func (f Field[T]) Table() string {
	return f.table
}

// ToSQL converts field expression to SQL
func (f Field[T]) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	// Escape identifier - handle table prefix if needed
	if f.table != "" && !strings.Contains(f.fieldPath, ".") {
		escaped := EscapeIdentifier(f.table) + "." + EscapeIdentifier(f.fieldPath)
		return escaped, nil, nil
	}
	escaped := EscapeIdentifier(f.fieldPath)
	return escaped, nil, nil
}

// Resolve validates the field exists in schema
func (f Field[T]) Resolve(schema *ModelSchema) error {
	// Validate field path exists
	parts := splitFieldPath(f.fieldPath)
	if len(parts) == 0 {
		return fmt.Errorf("empty field path")
	}

	// Check first part exists
	field := schema.GetField(parts[0])
	if field == nil {
		return fmt.Errorf("field %s not found in model", parts[0])
	}

	// Validate type matches
	if field.Type != f.fieldType {
		return fmt.Errorf("field %s has type %v, expected %v", parts[0], field.Type, f.fieldType)
	}

	// TODO: Validate relation paths for nested fields

	return nil
}

// Arithmetic operations

// Add creates an addition expression
func (f Field[T]) Add(other Field[T]) CombinedExpression[T] {
	return CombinedExpression[T]{
		Left:  f,
		Op:    OpAdd,
		Right: other,
	}
}

// Sub creates a subtraction expression
func (f Field[T]) Sub(other Field[T]) CombinedExpression[T] {
	return CombinedExpression[T]{
		Left:  f,
		Op:    OpSub,
		Right: other,
	}
}

// Mul creates a multiplication expression
func (f Field[T]) Mul(other Field[T]) CombinedExpression[T] {
	return CombinedExpression[T]{
		Left:  f,
		Op:    OpMul,
		Right: other,
	}
}

// Div creates a division expression
func (f Field[T]) Div(other Field[T]) CombinedExpression[T] {
	return CombinedExpression[T]{
		Left:  f,
		Op:    OpDiv,
		Right: other,
	}
}

// Comparison operations

// Eq creates an equality comparison
func (f Field[T]) Eq(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpEquals,
		Value: val,
	}
}

// Ne creates a not-equal comparison
func (f Field[T]) Ne(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpNotEquals,
		Value: val,
	}
}

// Gt creates a greater-than comparison
func (f Field[T]) Gt(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpGreater,
		Value: val,
	}
}

// Gte creates a greater-than-or-equal comparison
func (f Field[T]) Gte(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpGreaterOrEqual,
		Value: val,
	}
}

// Lt creates a less-than comparison
func (f Field[T]) Lt(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpLess,
		Value: val,
	}
}

// Lte creates a less-than-or-equal comparison
func (f Field[T]) Lte(val T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpLessOrEqual,
		Value: val,
	}
}

// In creates an IN clause comparison
func (f Field[T]) In(vals ...T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpIn,
		Value: vals,
	}
}

// IsNull creates an IS NULL comparison
func (f Field[T]) IsNull() ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpIsNull,
		Value: nil,
	}
}

// IsNotNull creates an IS NOT NULL comparison
func (f Field[T]) IsNotNull() ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpIsNotNull,
		Value: nil,
	}
}

// Range creates a BETWEEN comparison
func (f Field[T]) Range(min, max T) ComparisonExpression[T] {
	return ComparisonExpression[T]{
		Field: f,
		Op:    OpRange,
		Value: []T{min, max},
	}
}

// String-specific operations (only for FieldExpression[string])

// Contains creates a LIKE '%value%' comparison (string only)
func (f Field[string]) Contains(val string) ComparisonExpression[string] {
	return ComparisonExpression[string]{
		Field: f,
		Op:    OpContains,
		Value: val,
	}
}

// StartsWith creates a LIKE 'value%' comparison (string only)
func (f Field[string]) StartsWith(val string) ComparisonExpression[string] {
	return ComparisonExpression[string]{
		Field: f,
		Op:    OpStartsWith,
		Value: val,
	}
}

// EndsWith creates a LIKE '%value' comparison (string only)
func (f Field[string]) EndsWith(val string) ComparisonExpression[string] {
	return ComparisonExpression[string]{
		Field: f,
		Op:    OpEndsWith,
		Value: val,
	}
}

// IContains creates an ILIKE '%value%' comparison (case-insensitive, string only)
func (f Field[string]) IContains(val string) ComparisonExpression[string] {
	return ComparisonExpression[string]{
		Field: f,
		Op:    OpIContains,
		Value: val,
	}
}

// IExact creates an ILIKE 'value' comparison (case-insensitive exact, string only)
func (f Field[string]) IExact(val string) ComparisonExpression[string] {
	return ComparisonExpression[string]{
		Field: f,
		Op:    OpIExact,
		Value: val,
	}
}

// Asc creates an ascending order field expression
func (f Field[T]) Asc() OrderFieldExpr[T] {
	return OrderFieldExpr[T]{
		field:     f,
		ascending: true,
	}
}

// Desc creates a descending order field expression
func (f Field[T]) Desc() OrderFieldExpr[T] {
	return OrderFieldExpr[T]{
		field:     f,
		ascending: false,
	}
}

// CombinedExpression represents arithmetic operations between fields
type CombinedExpression[T any] struct {
	Left  Field[T]
	Op    ArithmeticOperator
	Right Field[T]
}

// ArithmeticOperator represents arithmetic operations
type ArithmeticOperator string

const (
	OpAdd ArithmeticOperator = "+"
	OpSub ArithmeticOperator = "-"
	OpMul ArithmeticOperator = "*"
	OpDiv ArithmeticOperator = "/"
	OpMod ArithmeticOperator = "%"
)

// ToSQL converts combined expression to SQL
func (c CombinedExpression[T]) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	leftSQL, leftArgs, err := c.Left.ToSQL(builder)
	if err != nil {
		return "", nil, err
	}

	rightSQL, rightArgs, err := c.Right.ToSQL(builder)
	if err != nil {
		return "", nil, err
	}

	sql := fmt.Sprintf("(%s %s %s)", leftSQL, c.Op, rightSQL)
	args := append(leftArgs, rightArgs...)

	return sql, args, nil
}

// Resolve validates the combined expression
func (c CombinedExpression[T]) Resolve(schema *ModelSchema) error {
	if err := c.Left.Resolve(schema); err != nil {
		return err
	}
	return c.Right.Resolve(schema)
}

// ComparisonExpression represents comparison operations
type ComparisonExpression[T any] struct {
	Field Field[T]
	Op    Operator
	Value interface{}
}

// ToSQL converts comparison expression to SQL
func (c ComparisonExpression[T]) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	fieldSQL, _, err := c.Field.ToSQL(builder)
	if err != nil {
		return "", nil, err
	}

	// Track args added by this expression
	startArgCount := len(builder.Args())

	// Build SQL based on operator
	var sql string
	
	// Handle operators that share the same string value using if-else
	if c.Op == OpContains {
		// LIKE '%value%'
		if strVal, ok := c.Value.(string); ok {
			pattern := "%" + strVal + "%"
			placeholder := builder.AddArg(pattern)
			sql = fmt.Sprintf("%s LIKE %s", fieldSQL, placeholder)
		} else {
			return "", nil, fmt.Errorf("Contains operator requires string value")
		}
	} else if c.Op == OpStartsWith {
		// LIKE 'value%'
		if strVal, ok := c.Value.(string); ok {
			pattern := strVal + "%"
			placeholder := builder.AddArg(pattern)
			sql = fmt.Sprintf("%s LIKE %s", fieldSQL, placeholder)
		} else {
			return "", nil, fmt.Errorf("StartsWith operator requires string value")
		}
	} else if c.Op == OpEndsWith {
		// LIKE '%value'
		if strVal, ok := c.Value.(string); ok {
			pattern := "%" + strVal
			placeholder := builder.AddArg(pattern)
			sql = fmt.Sprintf("%s LIKE %s", fieldSQL, placeholder)
		} else {
			return "", nil, fmt.Errorf("EndsWith operator requires string value")
		}
	} else if c.Op == OpIContains {
		// ILIKE '%value%'
		if strVal, ok := c.Value.(string); ok {
			pattern := "%" + strVal + "%"
			placeholder := builder.AddArg(pattern)
			sql = fmt.Sprintf("%s ILIKE %s", fieldSQL, placeholder)
		} else {
			return "", nil, fmt.Errorf("IContains operator requires string value")
		}
	} else if c.Op == OpIExact {
		// ILIKE 'value'
		if strVal, ok := c.Value.(string); ok {
			placeholder := builder.AddArg(strVal)
			sql = fmt.Sprintf("%s ILIKE %s", fieldSQL, placeholder)
		} else {
			return "", nil, fmt.Errorf("IExact operator requires string value")
		}
	} else {
		// Use switch for operators with unique values
		switch c.Op {
		case OpIsNull:
			sql = fmt.Sprintf("%s IS NULL", fieldSQL)
		case OpIsNotNull:
			sql = fmt.Sprintf("%s IS NOT NULL", fieldSQL)
		case OpIn:
			values, ok := c.Value.([]T)
			if !ok {
				return "", nil, fmt.Errorf("IN operator requires slice value")
			}
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = builder.AddArg(val)
			}
			sql = fmt.Sprintf("%s IN (%s)", fieldSQL, strings.Join(placeholders, ", "))
		case OpNotIn:
			values, ok := c.Value.([]T)
			if !ok {
				return "", nil, fmt.Errorf("NOT IN operator requires slice value")
			}
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = builder.AddArg(val)
			}
			sql = fmt.Sprintf("%s NOT IN (%s)", fieldSQL, strings.Join(placeholders, ", "))
		case OpRange:
			values, ok := c.Value.([]T)
			if !ok || len(values) != 2 {
				return "", nil, fmt.Errorf("Range operator requires slice of 2 values")
			}
			placeholder1 := builder.AddArg(values[0])
			placeholder2 := builder.AddArg(values[1])
			sql = fmt.Sprintf("%s BETWEEN %s AND %s", fieldSQL, placeholder1, placeholder2)
		default:
			// Standard operators (=, !=, >, >=, <, <=)
			placeholder := builder.AddArg(c.Value)
			sql = fmt.Sprintf("%s %s %s", fieldSQL, c.Op, placeholder)
		}
	}

	// Return only the args added by this expression
	allArgs := builder.Args()
	newArgs := allArgs[startArgCount:]
	return sql, newArgs, nil
}

// Resolve validates the comparison expression
func (c ComparisonExpression[T]) Resolve(schema *ModelSchema) error {
	return c.Field.Resolve(schema)
}

// ValueExpression represents a literal value
type ValueExpression[T any] struct {
	value T
}

// NewValue creates a value expression
func NewValue[T any](val T) ValueExpression[T] {
	return ValueExpression[T]{value: val}
}

// ToSQL converts value expression to SQL
func (v ValueExpression[T]) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	placeholder := builder.AddArg(v.value)
	return placeholder, nil, nil
}

// Resolve validates the value expression (always valid)
func (v ValueExpression[T]) Resolve(schema *ModelSchema) error {
	return nil
}

// BoolExpression represents boolean combinations of expressions (AND/OR)
type BoolExpression struct {
	operator Connector
	children []Expression
}

// ToSQL converts boolean expression to SQL
func (b *BoolExpression) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	if len(b.children) == 0 {
		return "1=1", nil, nil
	}

	var parts []string
	var allArgs []interface{}

	for _, expr := range b.children {
		sql, args, err := expr.ToSQL(builder)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, fmt.Sprintf("(%s)", sql))
		if args != nil {
			allArgs = append(allArgs, args...)
		}
	}

	combinedSQL := strings.Join(parts, " "+string(b.operator)+" ")
	return combinedSQL, allArgs, nil
}

// Resolve validates all expressions in boolean expression
func (b *BoolExpression) Resolve(schema *ModelSchema) error {
	for _, expr := range b.children {
		if err := expr.Resolve(schema); err != nil {
			return err
		}
	}
	return nil
}

// NotExpression represents a negated expression
type NotExpression struct {
	inner Expression
}

// ToSQL converts not expression to SQL
func (n *NotExpression) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	sql, args, err := n.inner.ToSQL(builder)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("NOT (%s)", sql), args, nil
}

// Resolve validates the inner expression
func (n *NotExpression) Resolve(schema *ModelSchema) error {
	return n.inner.Resolve(schema)
}

// EmptyExpression represents an empty query (matches all)
type EmptyExpression struct{}

// ToSQL converts empty expression to SQL
func (e *EmptyExpression) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	return "1=1", nil, nil
}

// Resolve validates empty expression (always valid)
func (e *EmptyExpression) Resolve(schema *ModelSchema) error {
	return nil
}

// And combines multiple expressions with AND logic
func And(expressions ...Expression) Expression {
	return &BoolExpression{
		operator: ConnectorAnd,
		children: expressions,
	}
}

// Or combines multiple expressions with OR logic
func Or(expressions ...Expression) Expression {
	return &BoolExpression{
		operator: ConnectorOr,
		children: expressions,
	}
}

// Not negates an expression
func Not(expr Expression) Expression {
	return &NotExpression{inner: expr}
}

// Q is for complex query composition (like Django's Q)
type Q struct {
	expressions []Expression
	connector   Connector // AND, OR
	negated     bool
}

// NewQ creates a new Q object from an expression
func NewQ(expr Expression) *Q {
	return &Q{
		expressions: []Expression{expr},
		connector:   ConnectorAnd,
	}
}

// And combines with AND
func (q *Q) And(other *Q) *Q {
	return &Q{
		expressions: []Expression{q, other},
		connector:   ConnectorAnd,
	}
}

// Or combines with OR
func (q *Q) Or(other *Q) *Q {
	return &Q{
		expressions: []Expression{q, other},
		connector:   ConnectorOr,
	}
}

// Not negates the expression
func (q *Q) Not() *Q {
	return &Q{
		expressions: q.expressions,
		connector:   q.connector,
		negated:     !q.negated,
	}
}

// ToSQL converts Q to SQL
func (q *Q) ToSQL(builder *SQLBuilder) (string, []interface{}, error) {
	if len(q.expressions) == 0 {
		return "1=1", nil, nil
	}

	var parts []string
	var allArgs []interface{}

	for _, expr := range q.expressions {
		sql, args, err := expr.ToSQL(builder)
		if err != nil {
			return "", nil, err
		}
		parts = append(parts, fmt.Sprintf("(%s)", sql))
		if args != nil {
			allArgs = append(allArgs, args...)
		}
	}

	combinedSQL := ""
	if len(parts) > 0 {
		combinedSQL = parts[0]
		for i := 1; i < len(parts); i++ {
			combinedSQL = fmt.Sprintf("%s %s %s", combinedSQL, q.connector, parts[i])
		}
	}

	if q.negated {
		combinedSQL = fmt.Sprintf("NOT (%s)", combinedSQL)
	}

	return combinedSQL, allArgs, nil
}

// Resolve validates all expressions in Q
func (q *Q) Resolve(schema *ModelSchema) error {
	for _, expr := range q.expressions {
		if err := expr.Resolve(schema); err != nil {
			return err
		}
	}
	return nil
}

// Connector represents how expressions are combined
type Connector string

const (
	ConnectorAnd Connector = "AND"
	ConnectorOr  Connector = "OR"
)

// splitFieldPath splits a field path by double underscores
func splitFieldPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "__")
}

// FieldRef represents a runtime field reference
type FieldRef struct {
	path string
}

// F creates a field reference
func F(fieldPath string) FieldRef {
	return FieldRef{path: fieldPath}
}

// NewFieldRef creates a field reference
func NewFieldRef(fieldPath string) FieldRef {
	return FieldRef{path: fieldPath}
}

// FieldRef methods (Eq, Ne, etc.)
// ... (Keeping these as they delegate to ComparisonExpression which we fixed)

// Eq creates an equality comparison
func (f FieldRef) Eq(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpEquals,
		Value: value,
	}
}

// Ne creates a not-equal comparison
func (f FieldRef) Ne(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpNotEquals,
		Value: value,
	}
}

// Gt creates a greater-than comparison
func (f FieldRef) Gt(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpGreater,
		Value: value,
	}
}

// Gte creates a greater-than-or-equal comparison
func (f FieldRef) Gte(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpGreaterOrEqual,
		Value: value,
	}
}

// Lt creates a less-than comparison
func (f FieldRef) Lt(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpLess,
		Value: value,
	}
}

// Lte creates a less-than-or-equal comparison
func (f FieldRef) Lte(value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpLessOrEqual,
		Value: value,
	}
}

// In creates an IN clause comparison
func (f FieldRef) In(values ...interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpIn,
		Value: values,
	}
}

// NotIn creates a NOT IN clause comparison
func (f FieldRef) NotIn(values ...interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpNotIn,
		Value: values,
	}
}

// IsNull creates an IS NULL comparison
func (f FieldRef) IsNull() Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpIsNull,
		Value: nil,
	}
}

// IsNotNull creates an IS NOT NULL comparison
func (f FieldRef) IsNotNull() Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpIsNotNull,
		Value: nil,
	}
}

// Contains creates a LIKE '%value%' comparison (string only)
func (f FieldRef) Contains(value string) Expression {
	return &ComparisonExpression[string]{
		Field: Field[string]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpContains,
		Value: value,
	}
}

// StartsWith creates a LIKE 'value%' comparison (string only)
func (f FieldRef) StartsWith(value string) Expression {
	return &ComparisonExpression[string]{
		Field: Field[string]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpStartsWith,
		Value: value,
	}
}

// EndsWith creates a LIKE '%value' comparison (string only)
func (f FieldRef) EndsWith(value string) Expression {
	return &ComparisonExpression[string]{
		Field: Field[string]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpEndsWith,
		Value: value,
	}
}

// IContains creates an ILIKE '%value%' comparison (case-insensitive, string only)
func (f FieldRef) IContains(value string) Expression {
	return &ComparisonExpression[string]{
		Field: Field[string]{
			fieldPath: f.path,
			table:     "",
		},
		Op:    OpIContains,
		Value: value,
	}
}
