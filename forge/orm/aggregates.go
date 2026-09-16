package orm

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	forgeerrors "github.com/forgego/forge/errors"
)

// AggregateValuer is an optional capability that QuerySet implementations may
// provide to support AggregateValues.
type AggregateValuer interface {
	AggregateValues(context.Context, ...Aggregate) (map[string]any, error)
}

type resolvedAggregate struct {
	name         string
	function     string
	field        string
	column       string
	fieldType    reflect.Type
	countStar    bool
	relationPath string
}

// AggregateValues executes ungrouped aggregates for any QuerySet implementation
// that supports them. Each aggregate is evaluated in its own relation scope.
// Predicates on a relation constrain aggregates over that relation.
// Aggregates across many-to-many relations are not supported yet.
func AggregateValues[T any](ctx context.Context, qs QuerySet[T], aggs ...Aggregate) (map[string]any, error) {
	valuer, ok := qs.(AggregateValuer)
	if !ok {
		return nil, forgeerrors.NewNotImplementedError("AggregateValues for this QuerySet implementation")
	}
	return valuer.AggregateValues(ctx, aggs...)
}

// AggregateValues executes ungrouped aggregate queries. Ordering, limits,
// and offsets are intentionally ignored. Each aggregate is evaluated in its own
// relation scope (base model or relation path).
func (qs *BaseQuerySet[T]) AggregateValues(ctx context.Context, aggs ...Aggregate) (map[string]any, error) {
	return qs.aggregateValues(ctx, aggs...)
}

type aggregateGroup struct {
	scope string
	aggs  []resolvedAggregate
}

func groupAggregatesByScope(resolved []resolvedAggregate) []aggregateGroup {
	var groups []aggregateGroup
	scopeIndex := make(map[string]int)
	for _, agg := range resolved {
		idx, exists := scopeIndex[agg.relationPath]
		if !exists {
			idx = len(groups)
			scopeIndex[agg.relationPath] = idx
			groups = append(groups, aggregateGroup{scope: agg.relationPath})
		}
		groups[idx].aggs = append(groups[idx].aggs, agg)
	}
	return groups
}

func (qs *BaseQuerySet[T]) aggregateValues(ctx context.Context, aggs ...Aggregate) (map[string]any, error) {
	if qs.err != nil && !isAggregateChainError(qs.err) {
		return nil, qs.err
	}
	resolved, err := qs.resolveAggregates(aggs)
	if err != nil {
		return nil, err
	}
	database, err := qs.getDB(ctx)
	if err != nil {
		return nil, err
	}

	groups := groupAggregatesByScope(resolved)
	result := make(map[string]any, len(resolved))

	for _, group := range groups {
		query, args, err := qs.buildAggregateSQL(group.aggs)
		if err != nil {
			return nil, err
		}

		values := make([]any, len(group.aggs))
		destinations := make([]any, len(values))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := database.QueryRowContext(ctx, query, args...).Scan(destinations...); err != nil {
			return nil, fmt.Errorf("aggregate query failed: %w", err)
		}

		for i, aggregate := range group.aggs {
			value, err := convertAggregateValue(aggregate, values[i])
			if err != nil {
				return nil, err
			}
			result[aggregate.name] = value
			if aggregate.field != "" {
				fallback := strings.ToLower(aggregate.function) + "_" + aggregate.field
				if _, exists := result[fallback]; !exists {
					result[fallback] = value
				}
			}
		}
	}
	return result, nil
}

func (qs *BaseQuerySet[T]) checkManyToManyHop(fieldPath string) error {
	if !strings.Contains(fieldPath, "__") || qs.schema == nil {
		return nil
	}
	parts := splitFieldPath(fieldPath)
	currentSchema := qs.schema
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if currentSchema == nil {
			break
		}
		rel := currentSchema.GetRelation(part)
		if rel == nil {
			break
		}
		if rel.Type == RelationManyToMany {
			return forgeerrors.NewNotImplementedError(
				fmt.Sprintf("aggregate field %s: aggregates across many-to-many relations are not supported yet", fieldPath),
			)
		}
		targetSchema, err := GetModelSchemaByName(rel.TargetModel)
		if err != nil {
			break
		}
		currentSchema = targetSchema
	}
	return nil
}

func isNumericType(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func (qs *BaseQuerySet[T]) resolveAggregates(aggs []Aggregate) ([]resolvedAggregate, error) {
	if len(aggs) == 0 {
		return nil, forgeerrors.NewInvalidInputError("aggregates", "at least one aggregate is required")
	}
	resolved := make([]resolvedAggregate, 0, len(aggs))
	names := make(map[string]bool, len(aggs))
	for _, aggregate := range aggs {
		function := strings.ToUpper(strings.TrimSpace(aggregate.Func))
		switch function {
		case string(AggCount), string(AggSum), string(AggAvg), string(AggMin), string(AggMax):
		default:
			return nil, forgeerrors.NewNotImplementedError("aggregate " + aggregate.Func)
		}
		name := aggregate.Name
		if name == "" {
			name = strings.ToLower(function) + "_" + aggregate.Field
		}
		if names[name] && aggregate.Name == strings.ToLower(function) && aggregate.Field != "" {
			name = strings.ToLower(function) + "_" + aggregate.Field
		}
		if names[name] {
			return nil, forgeerrors.NewInvalidInputError("aggregates", "duplicate aggregate name "+name)
		}
		names[name] = true

		item := resolvedAggregate{name: name, function: function, field: aggregate.Field}
		if strings.Contains(aggregate.Field, "__") {
			parts := splitFieldPath(aggregate.Field)
			if len(parts) > 1 {
				item.relationPath = strings.Join(parts[:len(parts)-1], "__")
			}
		}
		if function == string(AggCount) && (aggregate.Field == "" || aggregate.Field == "*") {
			item.countStar = true
		} else {
			if err := qs.checkManyToManyHop(aggregate.Field); err != nil {
				return nil, err
			}
			field, _, err := qs.schema.ResolvePath(aggregate.Field)
			if err != nil {
				return nil, forgeerrors.NewInvalidInputError(aggregate.Field, "field does not resolve to a model column")
			}
			item.column = EscapeIdentifier(field.DBColumn)
			item.fieldType = field.Type
			if function == string(AggSum) || function == string(AggAvg) {
				if !isNumericType(item.fieldType) {
					return nil, forgeerrors.NewInvalidInputError(aggregate.Field, fmt.Sprintf("%s aggregate requires a numeric field, got %v", function, item.fieldType))
				}
			}
		}
		resolved = append(resolved, item)
	}
	return resolved, nil
}

func convertAggregateValue(aggregate resolvedAggregate, value any) (any, error) {
	if aggregate.function == string(AggMin) || aggregate.function == string(AggMax) {
		if bytes, ok := value.([]byte); ok {
			if aggregate.fieldType != nil && aggregate.fieldType.Kind() == reflect.String {
				return string(bytes), nil
			}
			return bytes, nil
		}
		return value, nil
	}
	if value == nil {
		if aggregate.function == string(AggCount) {
			return int64(0), nil
		}
		return nil, nil
	}
	var text string
	switch number := value.(type) {
	case int64:
		if aggregate.function == string(AggCount) {
			return number, nil
		}
		return float64(number), nil
	case float64:
		if aggregate.function == string(AggCount) {
			return int64(number), nil
		}
		return number, nil
	case []byte:
		text = string(number)
	case string:
		text = number
	default:
		text = fmt.Sprint(value)
	}
	if aggregate.function == string(AggCount) {
		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("aggregate %s: convert COUNT result: %w", aggregate.name, err)
		}
		return parsed, nil
	}
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return nil, fmt.Errorf("aggregate %s: convert numeric result: %w", aggregate.name, err)
	}
	return parsed, nil
}

// Aggregate represents an aggregate function
type Aggregate struct {
	Name  string
	Field string
	Func  string // COUNT, SUM, AVG, MAX, MIN, etc.
}

// AggregateFunc represents an aggregate function
type AggregateFunc string

const (
	AggCount    AggregateFunc = "COUNT"
	AggSum      AggregateFunc = "SUM"
	AggAvg      AggregateFunc = "AVG"
	AggMax      AggregateFunc = "MAX"
	AggMin      AggregateFunc = "MIN"
	AggStdDev   AggregateFunc = "STDDEV"
	AggVariance AggregateFunc = "VARIANCE"
)

var (
	aggregateRegistryMu sync.RWMutex
	aggregateRegistry   = map[string]aggregateRegistration{}
)

type aggregateRegistration struct {
	funcName string
	builder  func(string) Aggregate
}

// Count creates a COUNT aggregate
func Count(field string) Aggregate {
	return Aggregate{
		Name:  "count",
		Field: field,
		Func:  string(AggCount),
	}
}

// Sum creates a SUM aggregate
func Sum(field string) Aggregate {
	return Aggregate{
		Name:  "sum",
		Field: field,
		Func:  string(AggSum),
	}
}

// Avg creates an AVG aggregate
func Avg(field string) Aggregate {
	return Aggregate{
		Name:  "avg",
		Field: field,
		Func:  string(AggAvg),
	}
}

// Max creates a MAX aggregate
func Max(field string) Aggregate {
	return Aggregate{
		Name:  "max",
		Field: field,
		Func:  string(AggMax),
	}
}

// Min creates a MIN aggregate
func Min(field string) Aggregate {
	return Aggregate{
		Name:  "min",
		Field: field,
		Func:  string(AggMin),
	}
}

// StdDev creates a STDDEV aggregate
func StdDev(field string) Aggregate {
	return Aggregate{
		Name:  "stddev",
		Field: field,
		Func:  string(AggStdDev),
	}
}

// Variance creates a VARIANCE aggregate
func Variance(field string) Aggregate {
	return Aggregate{
		Name:  "variance",
		Field: field,
		Func:  string(AggVariance),
	}
}

// RegisterAggregate registers a custom aggregate function
func RegisterAggregate(name, funcName string, builder func(string) Aggregate) {
	if builder == nil || strings.TrimSpace(name) == "" {
		return
	}

	aggregateRegistryMu.Lock()
	defer aggregateRegistryMu.Unlock()
	aggregateRegistry[normalizeRegistryName(name)] = aggregateRegistration{
		funcName: strings.TrimSpace(funcName),
		builder:  builder,
	}
}

// BuildAggregate builds an aggregate from a registered custom aggregate name.
func BuildAggregate(name, field string) (Aggregate, bool) {
	aggregateRegistryMu.RLock()
	registration, ok := aggregateRegistry[normalizeRegistryName(name)]
	aggregateRegistryMu.RUnlock()
	if !ok {
		return Aggregate{}, false
	}

	aggregate := registration.builder(field)
	if aggregate.Name == "" {
		aggregate.Name = normalizeRegistryName(name)
	}
	if aggregate.Field == "" {
		aggregate.Field = field
	}
	if aggregate.Func == "" {
		aggregate.Func = registration.funcName
	}
	return aggregate, true
}
