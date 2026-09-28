package orm

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Operator represents a SQL operator
type Operator string

const (
	OpEquals         Operator = "="
	OpNotEquals      Operator = "!="
	OpGreater        Operator = ">"
	OpGreaterOrEqual Operator = ">="
	OpLess           Operator = "<"
	OpLessOrEqual    Operator = "<="
	OpIn             Operator = "IN"
	OpNotIn          Operator = "NOT IN"
	OpIsNull         Operator = "IS NULL"
	OpIsNotNull      Operator = "IS NOT NULL"
	OpContains       Operator = "LIKE"        // '%value%'
	OpStartsWith     Operator = "STARTS_WITH" // 'value%'
	OpEndsWith       Operator = "ENDS_WITH"   // '%value'
	OpIContains      Operator = "ILIKE"       // '%value%' (case-insensitive)
	OpIExact         Operator = "IEXACT"      // 'value' (case-insensitive exact)
	OpIStartsWith    Operator = "ISTARTSWITH"
	OpIEndsWith      Operator = "IENDSWITH"
	OpRange          Operator = "BETWEEN"
	OpYear           Operator = "EXTRACT(YEAR FROM"
	OpMonth          Operator = "EXTRACT(MONTH FROM"
	OpDay            Operator = "EXTRACT(DAY FROM"
)

// QueryExpr represents a query condition (renamed from Q to differentiate from Django)
// This is the primary type-safe query building API
type QueryExpr struct {
	field    string
	op       Operator
	value    interface{}
	children []QueryExpr
	combiner Combiner
	negated  bool
}

// Where creates a simple field condition (SQL-like, explicit)
// This is the recommended way to create conditions when you don't have type-safe fields
//
// Example:
//
//	qs.Filter(Where("age", OpGreater, 18))
//	qs.Filter(Where("name", OpEquals, "John"))
func Where(field string, op Operator, value interface{}) Expression {
	return &ComparisonExpression[interface{}]{
		Field: Field[interface{}]{
			fieldPath: field,
			table:     "",
		},
		Op:    op,
		Value: value,
	}
}

// NewFieldQueryExpr creates a simple QueryExpr for field = value
//
// Deprecated: Use Where() for explicit SQL-like conditions or field expression methods for type-safe queries.
// NewFieldQueryExpr will be removed in v2.0.
//
// Migration:
//
//	// Old
//	expr := orm.NewFieldQueryExpr("age", orm.OpGreater, 18)
//	// New - Option 1: Where (explicit)
//	expr := orm.Where("age", orm.OpGreater, 18)
//	// New - Option 2: Type-safe (best)
//	expr := User.Age.Gt(18)
func NewFieldQueryExpr(field string, op Operator, value interface{}) QueryExpr {
	return QueryExpr{
		field: field,
		op:    op,
		value: value,
	}
}

// Combiner represents how conditions are combined
type Combiner string

const (
	CombineAnd Combiner = "AND"
	CombineOr  Combiner = "OR"
)

// And combines this QueryExpr with another using AND
func (q QueryExpr) And(other QueryExpr) QueryExpr {
	return QueryExpr{
		children: []QueryExpr{q, other},
		combiner: CombineAnd,
	}
}

// Or combines this QueryExpr with another using OR
func (q QueryExpr) Or(other QueryExpr) QueryExpr {
	return QueryExpr{
		children: []QueryExpr{q, other},
		combiner: CombineOr,
	}
}

// Not negates this QueryExpr
func (q QueryExpr) Not() QueryExpr {
	q.negated = !q.negated
	return q
}

// ToSQL converts the QueryExpr to SQL with parameters
// paramIndex is the starting parameter index (1-based for PostgreSQL)
func (q QueryExpr) ToSQL(paramIndex int, placeholder ...func(int) string) (string, []interface{}, int) {
	ph := defaultPlaceholder
	if len(placeholder) > 0 && placeholder[0] != nil {
		ph = placeholder[0]
	}

	if len(q.children) > 0 || q.combiner != "" {
		return q.buildCombined(paramIndex, ph)
	}

	return q.buildSingle(paramIndex, ph)
}

// buildCombined builds SQL for combined conditions (AND/OR)
func (q QueryExpr) buildCombined(paramIndex int, ph func(int) string) (string, []interface{}, int) {
	var parts []string
	var allArgs []interface{}
	currentIndex := paramIndex

	for _, child := range q.children {
		sql, args, nextIndex := child.ToSQL(currentIndex, ph)
		parts = append(parts, fmt.Sprintf("(%s)", sql))
		allArgs = append(allArgs, args...)
		currentIndex = nextIndex
	}

	combinedSQL := strings.Join(parts, " "+string(q.combiner)+" ")
	if q.negated {
		combinedSQL = fmt.Sprintf("NOT (%s)", combinedSQL)
	}

	return combinedSQL, allArgs, currentIndex
}

// toInterfaceSlice converts a slice or array of any type to []interface{}.
// Returns (nil, false) for non-slice/array values or nil.
func toInterfaceSlice(v interface{}) ([]interface{}, bool) {
	if v == nil {
		return nil, false
	}
	val := reflect.ValueOf(v)
	kind := val.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return nil, false
	}
	length := val.Len()
	res := make([]interface{}, length)
	for i := 0; i < length; i++ {
		res[i] = val.Index(i).Interface()
	}
	return res, true
}

func buildInOrRange(q QueryExpr, currentIndex int, ph func(int) string) (string, []interface{}, int, bool) {
	if q.op == OpIn || q.op == OpNotIn {
		values, ok := toInterfaceSlice(q.value)
		if !ok || len(values) == 0 {
			if q.op == OpIn {
				return "1=0", nil, currentIndex, true
			}
			return "1=1", nil, currentIndex, true
		}
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = ph(currentIndex + i)
		}
		opStr := "IN"
		if q.op == OpNotIn {
			opStr = "NOT IN"
		}
		sql := fmt.Sprintf("%s %s (%s)", q.field, opStr, strings.Join(placeholders, ", "))
		return sql, values, currentIndex + len(values), true
	}
	if q.op == OpRange {
		values, ok := toInterfaceSlice(q.value)
		if !ok || len(values) != 2 {
			return "1=0", nil, currentIndex, true
		}
		sql := fmt.Sprintf("%s BETWEEN %s AND %s", q.field, ph(currentIndex), ph(currentIndex+1))
		return sql, []interface{}{values[0], values[1]}, currentIndex + 2, true
	}
	return "", nil, currentIndex, false
}

func buildLikeExpr(q QueryExpr, currentIndex int, ph func(int) string) (string, []interface{}, int, bool) {
	strVal, ok := q.value.(string)
	if !ok {
		return "", nil, currentIndex, false
	}
	switch q.op {
	case OpContains:
		return fmt.Sprintf("%s LIKE %s", q.field, ph(currentIndex)), []interface{}{"%" + strVal + "%"}, currentIndex + 1, true
	case OpStartsWith:
		return fmt.Sprintf("%s LIKE %s", q.field, ph(currentIndex)), []interface{}{strVal + "%"}, currentIndex + 1, true
	case OpEndsWith:
		return fmt.Sprintf("%s LIKE %s", q.field, ph(currentIndex)), []interface{}{"%" + strVal}, currentIndex + 1, true
	case OpIContains:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", q.field, ph(currentIndex)), []interface{}{"%" + strVal + "%"}, currentIndex + 1, true
	case OpIExact:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", q.field, ph(currentIndex)), []interface{}{strVal}, currentIndex + 1, true
	case OpIStartsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", q.field, ph(currentIndex)), []interface{}{strVal + "%"}, currentIndex + 1, true
	case OpIEndsWith:
		return fmt.Sprintf("LOWER(%s) LIKE LOWER(%s)", q.field, ph(currentIndex)), []interface{}{"%" + strVal}, currentIndex + 1, true
	default:
		return "", nil, currentIndex, false
	}
}

func buildDateExtract(q QueryExpr, currentIndex int, ph func(int) string) (string, []interface{}, int, bool) {
	var part string
	switch q.op {
	case OpYear:
		part = "YEAR"
	case OpMonth:
		part = "MONTH"
	case OpDay:
		part = "DAY"
	default:
		return "", nil, currentIndex, false
	}
	sql := fmt.Sprintf("EXTRACT(%s FROM %s) = %s", part, q.field, ph(currentIndex))
	return sql, []interface{}{q.value}, currentIndex + 1, true
}

func (q QueryExpr) wrapNegated(sql string) string {
	if q.negated {
		return fmt.Sprintf("NOT (%s)", sql)
	}
	return sql
}

// buildSingle builds SQL for a single condition
func (q QueryExpr) buildSingle(paramIndex int, ph func(int) string) (string, []interface{}, int) {
	if q.op == OpIsNull {
		if b, ok := q.value.(bool); ok && !b {
			return q.wrapNegated(fmt.Sprintf("%s IS NOT NULL", q.field)), nil, paramIndex
		}
		return q.wrapNegated(fmt.Sprintf("%s IS NULL", q.field)), nil, paramIndex
	}
	if q.op == OpIsNotNull {
		return q.wrapNegated(fmt.Sprintf("%s IS NOT NULL", q.field)), nil, paramIndex
	}
	if sql, args, next, ok := buildInOrRange(q, paramIndex, ph); ok {
		return q.wrapNegated(sql), args, next
	}
	if sql, args, next, ok := buildLikeExpr(q, paramIndex, ph); ok {
		return q.wrapNegated(sql), args, next
	}
	if sql, args, next, ok := buildDateExtract(q, paramIndex, ph); ok {
		return q.wrapNegated(sql), args, next
	}
	sql := fmt.Sprintf("%s %s %s", q.field, q.op, ph(paramIndex))
	return q.wrapNegated(sql), []interface{}{q.value}, paramIndex + 1
}

// RegisterQueryExpr registers a custom query expression type
func RegisterQueryExpr(name string, builder func(...interface{}) QueryExpr) {
	if builder == nil || normalizeRegistryName(name) == "" {
		return
	}

	queryExprRegistryMu.Lock()
	defer queryExprRegistryMu.Unlock()
	queryExprRegistry[normalizeRegistryName(name)] = builder
}

// BuildQueryExpr builds a query expression from a registered custom query expression name.
func BuildQueryExpr(name string, args ...interface{}) (QueryExpr, bool) {
	queryExprRegistryMu.RLock()
	builder, ok := queryExprRegistry[normalizeRegistryName(name)]
	queryExprRegistryMu.RUnlock()
	if !ok {
		return QueryExpr{}, false
	}
	return builder(args...), true
}

var (
	queryExprRegistryMu sync.RWMutex
	queryExprRegistry   = map[string]func(...interface{}) QueryExpr{}
)

func normalizeRegistryName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
