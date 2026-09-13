package orm

import (
	"strings"
	"sync"
)

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
