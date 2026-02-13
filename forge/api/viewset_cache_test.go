package api

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

// MockQuerySet for testing
type MockQuerySet struct {
	count int64
	items []interface{}
}

func (m *MockQuerySet) Count(ctx context.Context) (int64, error) {
	return m.count, nil
}

func (m *MockQuerySet) All(ctx context.Context) ([]interface{}, error) {
	return m.items, nil
}

func (m *MockQuerySet) Filter(expr interface{}) interface{} {
	return m
}

func (m *MockQuerySet) OrderBy(fields ...interface{}) interface{} {
	return m
}

func (m *MockQuerySet) Limit(limit int) interface{} {
	return m
}

func (m *MockQuerySet) Offset(offset int) interface{} {
	return m
}

// TestMethodCache_GetMethod tests the method caching functionality
func TestMethodCache_GetMethod(t *testing.T) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	// First call should cache the method
	method1, found1 := cache.GetMethod(qsType, "Count")
	if !found1 {
		t.Fatal("Expected to find Count method")
	}
	if method1.Name != "Count" {
		t.Fatal("Expected method name to be Count")
	}

	// Second call should return cached method
	method2, found2 := cache.GetMethod(qsType, "Count")
	if !found2 {
		t.Fatal("Expected to find cached Count method")
	}
	if method1.Index != method2.Index {
		t.Fatal("Expected same cached method")
	}

	// Non-existent method should return false
	_, found3 := cache.GetMethod(qsType, "NonExistentMethod")
	if found3 {
		t.Fatal("Expected not to find non-existent method")
	}
}

// TestMethodCache_GetField tests the field caching functionality
func TestMethodCache_GetField(t *testing.T) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	type TestStruct struct {
		Name string
		Age  int
	}

	tsType := reflect.TypeOf(TestStruct{})

	// First call should cache the field
	field1, found1 := cache.GetField(tsType, "Name")
	if !found1 {
		t.Fatal("Expected to find Name field")
	}
	if field1.Name != "Name" {
		t.Fatal("Expected field name to be Name")
	}

	// Second call should return cached field
	field2, found2 := cache.GetField(tsType, "Name")
	if !found2 {
		t.Fatal("Expected to find cached Name field")
	}
	if field1.Name != field2.Name {
		t.Fatal("Expected same cached field")
	}

	// Non-existent field should return false
	_, found3 := cache.GetField(tsType, "NonExistentField")
	if found3 {
		t.Fatal("Expected not to find non-existent field")
	}
}

// TestMethodCache_Concurrency tests thread safety of the cache
func TestMethodCache_Concurrency(t *testing.T) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			method, found := cache.GetMethod(qsType, "Count")
			if !found {
				t.Error("Expected to find Count method")
				return
			}
			if method.Name != "Count" {
				t.Error("Expected method name to be Count")
			}
		}()
	}

	wg.Wait()
}

// TestMethodCache_Clear tests the cache clearing functionality
func TestMethodCache_Clear(t *testing.T) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	// Cache a method
	_, _ = cache.GetMethod(qsType, "Count")

	// Verify it's cached
	if len(cache.methods) == 0 {
		t.Fatal("Expected methods to be cached")
	}

	// Clear the cache
	cache.Clear()

	// Verify it's empty
	if len(cache.methods) != 0 {
		t.Fatal("Expected methods to be cleared")
	}
	if len(cache.fields) != 0 {
		t.Fatal("Expected fields to be cleared")
	}
}

// TestMethodCache_CacheTypeMethods tests pre-caching methods
func TestMethodCache_CacheTypeMethods(t *testing.T) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	methodNames := []string{"Count", "All", "Filter", "OrderBy", "Limit", "Offset"}
	cache.CacheTypeMethods(qsType, methodNames)

	// Verify all methods are cached
	for _, name := range methodNames {
		_, found := cache.GetMethod(qsType, name)
		if !found {
			t.Fatalf("Expected to find cached method %s", name)
		}
	}
}

// TestPreCacheQuerySetType tests the convenience function
func TestPreCacheQuerySetType(t *testing.T) {
	// Clear global cache first
	globalCache.Clear()

	mockQS := &MockQuerySet{count: 10}
	PreCacheQuerySetType(mockQS)

	qsType := reflect.TypeOf(mockQS)

	// Verify methods that exist on MockQuerySet are cached
	expectedMethods := []string{"Count", "All", "Filter", "OrderBy", "Limit", "Offset"}
	for _, name := range expectedMethods {
		_, found := globalCache.GetMethod(qsType, name)
		if !found {
			t.Fatalf("Expected to find cached method %s", name)
		}
	}
}

// BenchmarkGetMethod_WithCache benchmarks cached method lookup
func BenchmarkGetMethod_WithCache(b *testing.B) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	// Pre-cache the method
	_, _ = cache.GetMethod(qsType, "Count")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.GetMethod(qsType, "Count")
	}
}

// BenchmarkGetMethod_WithoutCache benchmarks non-cached method lookup
func BenchmarkGetMethod_WithoutCache(b *testing.B) {
	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = qsType.MethodByName("Count")
	}
}

// BenchmarkMethodCall_WithCache benchmarks calling a cached method
func BenchmarkMethodCall_WithCache(b *testing.B) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsValue := reflect.ValueOf(mockQS)
	qsType := qsValue.Type()

	// Pre-cache the method
	method, _ := cache.GetMethod(qsType, "Count")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		method.Func.Call([]reflect.Value{qsValue, reflect.ValueOf(ctx)})
	}
}

// BenchmarkMethodCall_WithoutCache benchmarks calling a method without caching
func BenchmarkMethodCall_WithoutCache(b *testing.B) {
	mockQS := &MockQuerySet{count: 10}
	qsValue := reflect.ValueOf(mockQS)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		method := qsValue.MethodByName("Count")
		method.Call([]reflect.Value{reflect.ValueOf(ctx)})
	}
}

// BenchmarkConcurrentGetMethod benchmarks concurrent cached method lookups
func BenchmarkConcurrentGetMethod(b *testing.B) {
	cache := &MethodCache{
		methods: make(map[reflect.Type]map[string]reflect.Method),
		fields:  make(map[reflect.Type]map[string]reflect.StructField),
	}

	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	// Pre-cache the method
	_, _ = cache.GetMethod(qsType, "Count")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.GetMethod(qsType, "Count")
		}
	})
}

// BenchmarkConcurrentMethodByName benchmarks concurrent non-cached method lookups
func BenchmarkConcurrentMethodByName(b *testing.B) {
	mockQS := &MockQuerySet{count: 10}
	qsType := reflect.TypeOf(mockQS)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = qsType.MethodByName("Count")
		}
	})
}
