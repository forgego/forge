package orm

import (
	"fmt"
	"strings"
)

// hasAnyPathJoins checks if any path joins exist across annotations, where conditions, or order by.
func (qs *BaseQuerySet[T]) hasAnyPathJoins() bool {
	if len(qs.annotations) == 0 && len(qs.conditions) == 0 && len(qs.excludes) == 0 && len(qs.orderBy) == 0 {
		return false
	}
	probeBuilder := qs.newSQLBuilder()
	var joins []string
	seen := make(map[string]bool)
	var multi bool
	probeBuilder.SetJoinResolver(qs.createJoinResolver(&joins, seen, &multi))

	for _, ann := range qs.annotations {
		expr := ann.Expression
		if expr == nil {
			expr = newQueryExprAdapter(ann.Expr)
		}
		if expr != nil {
			_, _, _ = expr.ToSQL(probeBuilder)
		}
	}
	_, _, _ = qs.buildWhereClause(probeBuilder)
	_, _ = qs.buildOrderByClause(probeBuilder)

	return len(joins) > 0 || multi
}

// buildSQL builds the SQL query
func (qs *BaseQuerySet[T]) buildSQL() (string, []interface{}, error) {
	builder := qs.newSQLBuilder()

	// Build select_related JOINs first (populates qs.joins and qs.joinMap)
	qs.buildJoinClause(builder)

	hasPathJoins := qs.hasAnyPathJoins()

	// Build SELECT clause first so annotation arguments are added in text order
	var annotationJoins []string
	annotationSeen := make(map[string]bool)
	var annotationMulti bool
	builder.SetJoinResolver(qs.createJoinResolver(&annotationJoins, annotationSeen, &annotationMulti))

	selectClause := qs.buildSelectClause(builder, hasPathJoins)
	if qs.err != nil {
		return "", nil, qs.err
	}

	// Build WHERE clause with resolver W
	var whereJoins []string
	whereSeen := make(map[string]bool)
	var whereMulti bool
	builder.SetJoinResolver(qs.createJoinResolver(&whereJoins, whereSeen, &whereMulti))

	whereClause, _, whereErr := qs.buildWhereClause(builder)
	if whereErr != nil {
		return "", nil, whereErr
	}

	// Build ORDER BY clause with resolver O
	var orderJoins []string
	orderSeen := make(map[string]bool)
	var orderMulti bool
	builder.SetJoinResolver(qs.createJoinResolver(&orderJoins, orderSeen, &orderMulti))

	orderByClause, orderErr := qs.buildOrderByClause(builder)
	if orderErr != nil {
		return "", nil, orderErr
	}

	fromClause := fmt.Sprintf("FROM %s", EscapeIdentifier(qs.table))
	limitClause := qs.buildLimitClause()
	offsetClause := qs.buildOffsetClause()

	var parts []string
	if whereMulti {
		parts = []string{selectClause, fromClause}
		if len(qs.joins) > 0 {
			parts = append(parts, strings.Join(qs.joins, " "))
		}
		outerJoins := mergeJoins(annotationJoins, orderJoins)
		if len(outerJoins) > 0 {
			parts = append(parts, strings.Join(outerJoins, " "))
		}
		parts = append(parts, qs.pkSubquery(whereJoins, whereClause))
	} else {
		pathJoins := mergeJoins(annotationJoins, whereJoins, orderJoins)
		parts = []string{selectClause, fromClause}
		if len(qs.joins) > 0 {
			parts = append(parts, strings.Join(qs.joins, " "))
		}
		if len(pathJoins) > 0 {
			parts = append(parts, strings.Join(pathJoins, " "))
		}
		if whereClause != "" {
			parts = append(parts, whereClause)
		}
	}

	if orderByClause != "" {
		parts = append(parts, orderByClause)
	}
	if limitClause != "" {
		parts = append(parts, limitClause)
	}
	if offsetClause != "" {
		parts = append(parts, offsetClause)
	}

	sql := strings.Join(parts, " ")
	args := builder.Args()

	return sql, args, nil
}

// mergeJoins merges join lists without duplicating identical joins.
func mergeJoins(joinLists ...[]string) []string {
	var merged []string
	seen := make(map[string]bool)
	for _, list := range joinLists {
		for _, j := range list {
			if !seen[j] {
				seen[j] = true
				merged = append(merged, j)
			}
		}
	}
	return merged
}

// pkSubquery formats a WHERE <table>.<pk> IN (SELECT <table>.<pk> FROM <table> <whereJoins> <where>) clause.
func (qs *BaseQuerySet[T]) pkSubquery(joins []string, where string) string {
	pkCol := "id"
	if qs.schema != nil && qs.schema.PrimaryKey != "" {
		pkCol = qs.schema.PrimaryKey
	}
	table := EscapeIdentifier(qs.table)
	pk := EscapeIdentifier(pkCol)
	innerParts := []string{fmt.Sprintf("SELECT %s.%s FROM %s", table, pk, table)}
	if len(joins) > 0 {
		innerParts = append(innerParts, strings.Join(joins, " "))
	}
	if where != "" {
		innerParts = append(innerParts, where)
	}
	return fmt.Sprintf("WHERE %s.%s IN (%s)", table, pk, strings.Join(innerParts, " "))
}

// buildSelectClause builds the SELECT clause
func (qs *BaseQuerySet[T]) buildSelectClause(builder *SQLBuilder, hasPathJoins bool) string {
	var fields []string

	if len(qs.selectFields) > 0 {
		for _, field := range qs.selectFields {
			if hasPathJoins && !strings.Contains(field, ".") {
				fields = append(fields, EscapeIdentifier(qs.table)+"."+EscapeIdentifier(field))
			} else {
				fields = append(fields, EscapeIdentifier(field))
			}
		}
	} else if len(qs.onlyFields) > 0 {
		for _, field := range qs.onlyFields {
			if hasPathJoins && !strings.Contains(field, ".") {
				fields = append(fields, EscapeIdentifier(qs.table)+"."+EscapeIdentifier(field))
			} else {
				fields = append(fields, EscapeIdentifier(field))
			}
		}
	} else {
		if hasPathJoins {
			fields = []string{EscapeIdentifier(qs.table) + ".*"}
		} else {
			fields = []string{"*"}
		}
	}

	// Add annotations to SELECT
	if len(qs.annotations) > 0 {
		for _, ann := range qs.annotations {
			expr := ann.Expression
			if expr == nil {
				expr = newQueryExprAdapter(ann.Expr)
			}
			if err := expr.Resolve(qs.schema); err != nil {
				if qs.err == nil {
					qs.err = err
				}
				continue
			}
			annSQL, _, err := expr.ToSQL(builder)
			if err != nil && qs.err == nil {
				qs.err = err
			}
			alias := EscapeIdentifier(ann.Name)
			fields = append(fields, fmt.Sprintf("%s AS %s", annSQL, alias))
		}
	}

	selectClause := "SELECT "
	if len(qs.distinctFields) > 0 {
		selectClause += "DISTINCT "
	}
	selectClause += strings.Join(fields, ", ")

	// Add related fields
	for _, path := range qs.selectRelated {
		if !qs.joinMap[path] {
			continue
		}
		rel := qs.schema.GetRelation(path)
		if rel == nil {
			continue
		}
		targetSchema, err := GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			continue
		}

		// Select all fields from target schema
		for _, f := range targetSchema.Fields {
			alias := EscapeIdentifier(rel.Name)
			col := EscapeIdentifier(f.DBColumn)
			colAlias := EscapeIdentifier(rel.Name + "__" + f.DBColumn)

			selectClause += fmt.Sprintf(", %s.%s AS %s", alias, col, colAlias)
		}
	}

	return selectClause
}

// buildWhereClause builds the WHERE clause.
// Returns the SQL string, arguments, and any error encountered during SQL generation.
func (qs *BaseQuerySet[T]) buildWhereClause(builder *SQLBuilder) (string, []interface{}, error) {
	var parts []string
	var allArgs []interface{}

	// Build conditions
	for _, cond := range qs.conditions {
		sql, args, err := cond.ToSQL(builder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build condition SQL: %w", err)
		}
		parts = append(parts, sql)
		allArgs = append(allArgs, args...)
	}

	// Build excludes (with NOT)
	for _, exclude := range qs.excludes {
		sql, args, err := exclude.ToSQL(builder)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build exclude SQL: %w", err)
		}
		parts = append(parts, fmt.Sprintf("NOT (%s)", sql))
		allArgs = append(allArgs, args...)
	}

	if len(parts) == 0 {
		return "", nil, nil
	}

	return "WHERE " + strings.Join(parts, " AND "), allArgs, nil
}

// buildOrderByClause builds the ORDER BY clause
func (qs *BaseQuerySet[T]) buildOrderByClause(builder *SQLBuilder) (string, error) {
	if len(qs.orderBy) == 0 {
		return "", nil
	}

	var parts []string
	for _, field := range qs.orderBy {
		escaped := EscapeIdentifier(field.Field)
		if builder != nil && strings.Contains(field.Field, "__") {
			resolved, err := builder.resolveColumn(field.Field)
			if err != nil {
				return "", fmt.Errorf("failed to resolve order by field %s: %w", field.Field, err)
			}
			escaped = resolved
		}
		if field.Ascending {
			parts = append(parts, escaped+" ASC")
		} else {
			parts = append(parts, escaped+" DESC")
		}
	}

	return "ORDER BY " + strings.Join(parts, ", "), nil
}

// buildLimitClause builds the LIMIT clause
func (qs *BaseQuerySet[T]) buildLimitClause() string {
	if qs.limitVal == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d", *qs.limitVal)
}

// buildOffsetClause builds the OFFSET clause
func (qs *BaseQuerySet[T]) buildOffsetClause() string {
	if qs.offsetVal == nil {
		return ""
	}
	return fmt.Sprintf("OFFSET %d", *qs.offsetVal)
}

// buildCountOrExistsSQL builds SQL for Count or Exists queries reusing the WHERE and JOIN builder.
func (qs *BaseQuerySet[T]) buildCountOrExistsSQL(isExists bool) (string, []interface{}, error) {
	builder := qs.newSQLBuilder()
	qs.buildJoinClause(builder)

	var whereJoins []string
	whereSeen := make(map[string]bool)
	var whereMulti bool
	builder.SetJoinResolver(qs.createJoinResolver(&whereJoins, whereSeen, &whereMulti))

	whereClause, _, err := qs.buildWhereClause(builder)
	if err != nil {
		return "", nil, err
	}

	table := EscapeIdentifier(qs.table)
	if isExists {
		parts := []string{fmt.Sprintf("SELECT 1 FROM %s", table)}
		if whereMulti {
			parts = append(parts, qs.pkSubquery(whereJoins, whereClause))
		} else {
			parts = qs.appendWhereAndJoinParts(parts, whereJoins, whereClause)
		}
		parts = append(parts, "LIMIT 1")
		return strings.Join(parts, " "), builder.Args(), nil
	}

	var parts []string
	if whereMulti {
		parts = []string{fmt.Sprintf("SELECT COUNT(*) FROM %s", table), qs.pkSubquery(whereJoins, whereClause)}
	} else {
		selectClause := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
		if len(whereJoins) > 0 {
			pkCol := "id"
			if qs.schema != nil && qs.schema.PrimaryKey != "" {
				pkCol = qs.schema.PrimaryKey
			}
			selectClause = fmt.Sprintf("SELECT COUNT(DISTINCT %s.%s) FROM %s", table, EscapeIdentifier(pkCol), table)
		}
		parts = qs.appendWhereAndJoinParts([]string{selectClause}, whereJoins, whereClause)
	}
	return strings.Join(parts, " "), builder.Args(), nil
}

func (qs *BaseQuerySet[T]) appendWhereAndJoinParts(parts []string, whereJoins []string, whereClause string) []string {
	if len(qs.joins) > 0 {
		parts = append(parts, strings.Join(qs.joins, " "))
	}
	if len(whereJoins) > 0 {
		parts = append(parts, strings.Join(whereJoins, " "))
	}
	if whereClause != "" {
		parts = append(parts, whereClause)
	}
	return parts
}

// buildCountSQL returns the SQL query and arguments for Count
func (qs *BaseQuerySet[T]) buildCountSQL() (string, []interface{}, error) {
	return qs.buildCountOrExistsSQL(false)
}

// buildAggregateSQL builds an ungrouped aggregate query using the same join and
// WHERE construction as Count. aggregates must have been validated first.
func (qs *BaseQuerySet[T]) buildAggregateSQL(aggregates []resolvedAggregate) (string, []interface{}, error) {
	builder := qs.newSQLBuilder()
	qs.buildJoinClause(builder)

	scope := ""
	if len(aggregates) > 0 {
		scope = aggregates[0].relationPath
	}

	var aggJoins []string
	aggSeen := make(map[string]bool)
	builder.SetJoinResolver(qs.createJoinResolver(&aggJoins, aggSeen, nil))

	selects := make([]string, 0, len(aggregates))
	for _, aggregate := range aggregates {
		column := aggregate.column
		if aggregate.countStar {
			column = "*"
		} else if strings.Contains(aggregate.field, "__") {
			var err error
			column, err = builder.resolveColumn(aggregate.field)
			if err != nil {
				return "", nil, err
			}
		} else {
			column = EscapeIdentifier(qs.table) + "." + column
		}
		selects = append(selects, fmt.Sprintf("%s(%s)", aggregate.function, column))
	}

	var whereJoins []string
	whereSeen := make(map[string]bool)
	var whereMulti bool
	builder.SetJoinResolver(qs.createJoinResolver(&whereJoins, whereSeen, &whereMulti))

	whereClause, _, err := qs.buildWhereClause(builder)
	if err != nil {
		return "", nil, err
	}
	parts := []string{fmt.Sprintf("SELECT %s FROM %s", strings.Join(selects, ", "), EscapeIdentifier(qs.table))}
	if whereMulti {
		if scope == "" {
			if len(qs.joins) > 0 {
				parts = append(parts, strings.Join(qs.joins, " "))
			}
			parts = append(parts, qs.aggregatePKSubquery(whereJoins, whereClause))
		} else {
			if len(qs.joins) > 0 {
				parts = append(parts, strings.Join(qs.joins, " "))
			}
			if len(aggJoins) > 0 {
				parts = append(parts, strings.Join(aggJoins, " "))
			}
			pkSub := qs.aggregatePKSubquery(whereJoins, whereClause)
			builder.SetJoinResolver(qs.createJoinResolver(&aggJoins, aggSeen, nil))
			scopePreds, err := qs.buildScopePredicates(builder, scope)
			if err != nil {
				return "", nil, err
			}
			if scopePreds != "" {
				parts = append(parts, pkSub+" AND "+scopePreds)
			} else {
				parts = append(parts, pkSub)
			}
		}
	} else {
		outerJoins := mergeJoins(aggJoins, whereJoins)
		parts = qs.appendWhereAndJoinParts(parts, outerJoins, whereClause)
	}
	return strings.Join(parts, " "), builder.Args(), nil
}

func (qs *BaseQuerySet[T]) conditionTouchesScope(expr Expression, scope string) bool {
	if expr == nil || scope == "" {
		return false
	}
	probeBuilder := qs.newSQLBuilder()
	touchesScope := false
	touchesOther := false
	probeBuilder.SetJoinResolver(func(parts []string) (string, string, error) {
		if len(parts) > 1 {
			relPath := strings.Join(parts[:len(parts)-1], "__")
			if relPath == scope || strings.HasPrefix(relPath, scope+"__") || strings.HasPrefix(scope, relPath+"__") {
				touchesScope = true
			} else {
				touchesOther = true
			}
		}
		return "alias", "col", nil
	})
	_, _, _ = expr.ToSQL(probeBuilder)
	return touchesScope && !touchesOther
}

func (qs *BaseQuerySet[T]) buildScopePredicates(builder *SQLBuilder, scope string) (string, error) {
	var parts []string
	for _, cond := range qs.conditions {
		if qs.conditionTouchesScope(cond, scope) {
			sql, _, err := cond.ToSQL(builder)
			if err != nil {
				return "", err
			}
			parts = append(parts, sql)
		}
	}
	for _, exclude := range qs.excludes {
		if qs.conditionTouchesScope(exclude, scope) {
			sql, _, err := exclude.ToSQL(builder)
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("NOT (%s)", sql))
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " AND "), nil
}

// aggregatePKSubquery restricts an aggregate to distinct base rows when a
// reverse or many-to-many filter would otherwise multiply joined rows.
func (qs *BaseQuerySet[T]) aggregatePKSubquery(joins []string, where string) string {
	pkCol := "id"
	if qs.schema != nil && qs.schema.PrimaryKey != "" {
		pkCol = qs.schema.PrimaryKey
	}
	table := EscapeIdentifier(qs.table)
	pk := EscapeIdentifier(pkCol)
	innerParts := []string{fmt.Sprintf("SELECT DISTINCT %s.%s FROM %s", table, pk, table)}
	if len(joins) > 0 {
		innerParts = append(innerParts, strings.Join(joins, " "))
	}
	if where != "" {
		innerParts = append(innerParts, where)
	}
	return fmt.Sprintf("WHERE %s.%s IN (%s)", table, pk, strings.Join(innerParts, " "))
}

// BuildExistsSQL builds the SQL query and arguments for Exists
func (qs *BaseQuerySet[T]) BuildExistsSQL() (string, []interface{}, error) {
	return qs.buildCountOrExistsSQL(true)
}

// buildExistsSQL builds the SQL query and arguments for Exists
func (qs *BaseQuerySet[T]) buildExistsSQL() (string, []interface{}, error) {
	return qs.BuildExistsSQL()
}
