package orm

import (
	"fmt"
	"reflect"
)

// PathToken represents a pre-validated, pre-compiled field path
// This eliminates reflection overhead in hot paths
type PathToken struct {
	segments    []string
	fieldTypes  []reflect.Type
	traversalFn func(interface{}) (interface{}, error)
	hash        uint64
	table       string
}

// CompilePath compiles a field path into a PathToken for fast repeated access
// This is a one-time compilation that can be cached and reused
func (ms *ModelSchema) CompilePath(path string) (*PathToken, error) {
	parts := splitFieldPath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty field path")
	}

	// Validate path exists
	fieldInfo, targetSchema, err := ms.ResolvePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	// Build traversal function
	traversalFn := buildTraversalFunction(parts, ms, targetSchema)

	// Calculate hash for caching
	hash := hashPath(path)

	table := ms.TableName
	if targetSchema != nil {
		table = targetSchema.TableName
	}

	return &PathToken{
		segments:    parts,
		fieldTypes:  []reflect.Type{fieldInfo.Type},
		traversalFn: traversalFn,
		hash:        hash,
		table:       table,
	}, nil
}

// Traverse executes the compiled path traversal
// This is fast because it uses a pre-compiled function, not reflection
func (pt *PathToken) Traverse(instance interface{}) (interface{}, error) {
	if pt.traversalFn == nil {
		return nil, fmt.Errorf("path token not properly compiled")
	}
	return pt.traversalFn(instance)
}

// Path returns the original path string
func (pt *PathToken) Path() string {
	result := pt.segments[0]
	for _, seg := range pt.segments[1:] {
		result += "__" + seg
	}
	return result
}

// Table returns the target table name
func (pt *PathToken) Table() string {
	return pt.table
}

// Hash returns the path hash for caching
func (pt *PathToken) Hash() uint64 {
	return pt.hash
}

// buildTraversalFunction creates a fast traversal function for a path
// This compiles the path into a function that can be called without reflection
func buildTraversalFunction(parts []string, sourceSchema *ModelSchema, targetSchema *ModelSchema) func(interface{}) (interface{}, error) {
	return func(instance interface{}) (interface{}, error) {
		current := reflect.ValueOf(instance)
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}

		// Traverse each segment
		for i, part := range parts {
			if current.Kind() != reflect.Struct {
				return nil, fmt.Errorf("cannot traverse %s: not a struct", part)
			}

			field := current.FieldByName(toPascalCasePath(part))
			if !field.IsValid() {
				// Try lowercase
				field = current.FieldByName(part)
				if !field.IsValid() {
					return nil, fmt.Errorf("field %s not found", part)
				}
			}

			// If this is the last segment, return the value
			if i == len(parts)-1 {
				if field.CanInterface() {
					return field.Interface(), nil
				}
				return nil, fmt.Errorf("cannot access field %s", part)
			}

			// Otherwise, continue traversal
			if field.Kind() == reflect.Ptr {
				if field.IsNil() {
					return nil, fmt.Errorf("nil pointer at %s", part)
				}
				current = field.Elem()
			} else {
				current = field
			}
		}

		return nil, fmt.Errorf("unexpected end of traversal")
	}
}

// hashPath creates a simple hash for path caching
func hashPath(path string) uint64 {
	var hash uint64 = 5381
	for _, c := range path {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// toPascalCasePath converts snake_case to PascalCase (path_token version)
func toPascalCasePath(s string) string {
	if len(s) == 0 {
		return s
	}
	result := ""
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			if r >= 'a' && r <= 'z' {
				result += string(r - 32)
			} else {
				result += string(r)
			}
			nextUpper = false
		} else {
			result += string(r)
		}
	}
	return result
}
