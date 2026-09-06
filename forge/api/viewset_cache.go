package api

import (
	"reflect"
	"sync"
)

// MethodCache caches reflection lookups for QuerySet methods to improve performance
// on high-traffic endpoints by avoiding repeated reflection calls.
type MethodCache struct {
	mu      sync.RWMutex
	methods map[reflect.Type]map[string]reflect.Method
	fields  map[reflect.Type]map[string]reflect.StructField
}

// globalCache is the shared cache instance for all viewset operations
var globalCache = &MethodCache{
	methods: make(map[reflect.Type]map[string]reflect.Method),
	fields:  make(map[reflect.Type]map[string]reflect.StructField),
}

// GetMethod returns a cached method or looks it up and caches it.
// This is thread-safe and uses double-checked locking for performance.
func (c *MethodCache) GetMethod(t reflect.Type, name string) (reflect.Method, bool) {
	// First check with read lock (fast path)
	c.mu.RLock()
	if typeMethods, ok := c.methods[t]; ok {
		if method, found := typeMethods[name]; found {
			c.mu.RUnlock()
			return method, true
		}
	}
	c.mu.RUnlock()

	// Method not cached, lookup and cache (slow path)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if typeMethods, ok := c.methods[t]; ok {
		if method, found := typeMethods[name]; found {
			return method, true
		}
	}

	// Lookup method
	method, found := t.MethodByName(name)
	if !found {
		return method, false
	}

	// Cache the method
	if c.methods[t] == nil {
		c.methods[t] = make(map[string]reflect.Method)
	}
	c.methods[t][name] = method
	return method, true
}

// GetField returns a cached struct field or looks it up and caches it.
// This is thread-safe and uses double-checked locking for performance.
func (c *MethodCache) GetField(t reflect.Type, name string) (reflect.StructField, bool) {
	// First check with read lock (fast path)
	c.mu.RLock()
	if typeFields, ok := c.fields[t]; ok {
		if field, found := typeFields[name]; found {
			c.mu.RUnlock()
			return field, true
		}
	}
	c.mu.RUnlock()

	// Field not cached, lookup and cache (slow path)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if typeFields, ok := c.fields[t]; ok {
		if field, found := typeFields[name]; found {
			return field, true
		}
	}

	// Lookup field
	field, found := t.FieldByName(name)
	if !found {
		return field, false
	}

	// Cache the field
	if c.fields[t] == nil {
		c.fields[t] = make(map[string]reflect.StructField)
	}
	c.fields[t][name] = field
	return field, true
}

// CacheTypeMethods pre-caches all commonly used methods for a given type.
// This can be called during initialization to avoid runtime cache misses.
func (c *MethodCache) CacheTypeMethods(t reflect.Type, methodNames []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.methods[t] == nil {
		c.methods[t] = make(map[string]reflect.Method)
	}

	for _, name := range methodNames {
		if _, found := c.methods[t][name]; !found {
			if method, ok := t.MethodByName(name); ok {
				c.methods[t][name] = method
			}
		}
	}
}

// Clear clears all cached methods and fields.
// This is primarily useful for testing.
func (c *MethodCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.methods = make(map[reflect.Type]map[string]reflect.Method)
	c.fields = make(map[reflect.Type]map[string]reflect.StructField)
}

// GetGlobalCache returns the global method cache instance.
func GetGlobalCache() *MethodCache {
	return globalCache
}

// CachedMethodCall represents a cached method with its receiver type.
// It can be used to call methods without repeated lookups.
type CachedMethodCall struct {
	method reflect.Method
	receiverType reflect.Type
}

// Call invokes the cached method with the given arguments.
func (c *CachedMethodCall) Call(receiver reflect.Value, args ...reflect.Value) []reflect.Value {
	allArgs := make([]reflect.Value, 0, len(args)+1)
	allArgs = append(allArgs, receiver)
	allArgs = append(allArgs, args...)
	return c.method.Func.Call(allArgs)
}

// GetCachedMethod returns a CachedMethodCall for efficient repeated calls.
func (c *MethodCache) GetCachedMethod(t reflect.Type, name string) (*CachedMethodCall, bool) {
	method, found := c.GetMethod(t, name)
	if !found {
		return nil, false
	}
	return &CachedMethodCall{
		method:       method,
		receiverType: t,
	}, true
}

// CommonQuerySetMethods lists the methods commonly used on QuerySet objects.
var CommonQuerySetMethods = []string{
	"Count",
	"All",
	"Filter",
	"OrderBy",
	"Limit",
	"Offset",
	"Get",
	"Create",
	"Update",
	"Delete",
	"First",
}

// CommonManagerMethods lists the methods commonly used on Manager objects.
var CommonManagerMethods = []string{
	"Get",
	"Create",
	"Update",
	"Delete",
	"All",
	"Filter",
	"Count",
}

// PreCacheQuerySetType pre-caches all common QuerySet methods for a type.
// Call this during initialization for optimal performance.
func PreCacheQuerySetType(qs interface{}) {
	qsType := reflect.TypeOf(qs)
	globalCache.CacheTypeMethods(qsType, CommonQuerySetMethods)
}

// PreCacheManagerType pre-caches all common Manager methods for a type.
// Call this during initialization for optimal performance.
func PreCacheManagerType(manager interface{}) {
	managerType := reflect.TypeOf(manager)
	globalCache.CacheTypeMethods(managerType, CommonManagerMethods)
}
