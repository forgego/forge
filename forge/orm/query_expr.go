package orm

import (
	"fmt"
	"strings"
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
	OpContains       Operator = "LIKE"  // '%value%'
	OpStartsWith     Operator = "LIKE"  // 'value%'
	OpEndsWith       Operator = "LIKE"  // '%value'
	OpIContains      Operator = "ILIKE" // '%value%' (case-insensitive)
	OpIExact         Operator = "ILIKE" // 'value' (case-insensitive exact)
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
//   qs.Filter(Where("age", OpGreater, 18))
//   qs.Filter(Where("name", OpEquals, "John"))
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
//   // Old
//   expr := orm.NewFieldQueryExpr("age", orm.OpGreater, 18)
//   // New - Option 1: Where (explicit)
//   expr := orm.Where("age", orm.OpGreater, 18)
//   // New - Option 2: Type-safe (best)
//   expr := User.Age.Gt(18)
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
func (q QueryExpr) ToSQL(paramIndex int) (string, []interface{}, int) {
	if len(q.children) > 0 {
		return q.buildCombined(paramIndex)
	}

	if q.combiner != "" {
		return q.buildCombined(paramIndex)
	}

	return q.buildSingle(paramIndex)
}


// buildCombined builds SQL for combined conditions (AND/OR)
func (q QueryExpr) buildCombined(paramIndex int) (string, []interface{}, int) {
	var parts []string
	var allArgs []interface{}
	currentIndex := paramIndex

	for _, child := range q.children {
		sql, args, nextIndex := child.ToSQL(currentIndex)
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

// buildSingle builds SQL for a single condition
func (q QueryExpr) buildSingle(paramIndex int) (string, []interface{}, int) {
	var sql string
	var args []interface{}
	currentIndex := paramIndex

	// Use if-else to handle operators with same string values
	if q.op == OpIsNull {
		sql = fmt.Sprintf("%s IS NULL", q.field)
	} else if q.op == OpIsNotNull {
		sql = fmt.Sprintf("%s IS NOT NULL", q.field)
	} else if q.op == OpIn {
		// Handle IN clause - use PostgreSQL placeholders
		values, ok := q.value.([]interface{})
		if !ok {
			return "", nil, currentIndex
		}
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", currentIndex+i)
			args = append(args, values[i])
		}
		sql = fmt.Sprintf("%s IN (%s)", q.field, strings.Join(placeholders, ", "))
		currentIndex += len(values)
	} else if q.op == OpNotIn {
		values, ok := q.value.([]interface{})
		if !ok {
			return "", nil, currentIndex
		}
		placeholders := make([]string, len(values))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", currentIndex+i)
			args = append(args, values[i])
		}
		sql = fmt.Sprintf("%s NOT IN (%s)", q.field, strings.Join(placeholders, ", "))
		currentIndex += len(values)
	} else if q.op == OpContains {
		sql = fmt.Sprintf("%s LIKE $%d", q.field, currentIndex)
		if strVal, ok := q.value.(string); ok {
			args = []interface{}{"%" + strVal + "%"}
		} else {
			return "", nil, currentIndex
		}
		currentIndex++
	} else if q.op == OpStartsWith {
		sql = fmt.Sprintf("%s LIKE $%d", q.field, currentIndex)
		if strVal, ok := q.value.(string); ok {
			args = []interface{}{strVal + "%"}
		} else {
			return "", nil, currentIndex
		}
		currentIndex++
	} else if q.op == OpEndsWith {
		sql = fmt.Sprintf("%s LIKE $%d", q.field, currentIndex)
		if strVal, ok := q.value.(string); ok {
			args = []interface{}{"%" + strVal}
		} else {
			return "", nil, currentIndex
		}
		currentIndex++
	} else if q.op == OpIContains {
		sql = fmt.Sprintf("%s ILIKE $%d", q.field, currentIndex)
		if strVal, ok := q.value.(string); ok {
			args = []interface{}{"%" + strVal + "%"}
		} else {
			return "", nil, currentIndex
		}
		currentIndex++
	} else if q.op == OpIExact {
		sql = fmt.Sprintf("%s ILIKE $%d", q.field, currentIndex)
		if strVal, ok := q.value.(string); ok {
			args = []interface{}{strVal}
		} else {
			return "", nil, currentIndex
		}
		currentIndex++
	} else if q.op == OpRange {
		values, ok := q.value.([]interface{})
		if !ok || len(values) != 2 {
			return "", nil, currentIndex
		}
		if len(values) == 2 {
			sql = fmt.Sprintf("%s BETWEEN $%d AND $%d", q.field, currentIndex, currentIndex+1)
			args = []interface{}{values[0], values[1]}
			currentIndex += 2
		}
	} else if q.op == OpYear {
		sql = fmt.Sprintf("EXTRACT(YEAR FROM %s) = $%d", q.field, currentIndex)
		args = []interface{}{q.value}
		currentIndex++
	} else if q.op == OpMonth {
		sql = fmt.Sprintf("EXTRACT(MONTH FROM %s) = $%d", q.field, currentIndex)
		args = []interface{}{q.value}
		currentIndex++
	} else if q.op == OpDay {
		sql = fmt.Sprintf("EXTRACT(DAY FROM %s) = $%d", q.field, currentIndex)
		args = []interface{}{q.value}
		currentIndex++
	} else {
		// Standard operators (=, !=, >, >=, <, <=) - use PostgreSQL placeholders
		sql = fmt.Sprintf("%s %s $%d", q.field, q.op, currentIndex)
		args = []interface{}{q.value}
		currentIndex++
	}

	if q.negated {
		sql = fmt.Sprintf("NOT (%s)", sql)
	}

	return sql, args, currentIndex
}

// RegisterQueryExpr registers a custom query expression type
func RegisterQueryExpr(name string, builder func(...interface{}) QueryExpr) {
	// TODO: Implement custom query expression registry
}
