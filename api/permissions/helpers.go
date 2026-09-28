package permissions

import (
	"reflect"
	"strings"
)

// getMethod gets a method by name using reflection
func getMethod(obj interface{}, methodName string) reflect.Value {
	if obj == nil {
		return reflect.Value{}
	}

	v := reflect.ValueOf(obj)

	// Try the provided value first (supports pointer-receiver methods).
	if method := v.MethodByName(methodName); method.IsValid() {
		return method
	}

	// If it's a pointer, also try the dereferenced value (value-receiver methods).
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		if method := v.Elem().MethodByName(methodName); method.IsValid() {
			return method
		}
	}

	return reflect.Value{}
}

// fieldByTag finds a field whose db or json tag matches name (before first comma).
func fieldByTag(v reflect.Value, t reflect.Type, name string) reflect.Value {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		for _, tagKey := range []string{"db", "json"} {
			tag := sf.Tag.Get(tagKey)
			if tag == "" {
				continue
			}
			tagName, _, _ := strings.Cut(tag, ",")
			if tagName == name {
				return v.Field(i)
			}
		}
	}
	return reflect.Value{}
}

// getField gets a field by name using reflection
func getField(obj interface{}, fieldName string) reflect.Value {
	if obj == nil || fieldName == "" {
		return reflect.Value{}
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	// 1. exact Go field name
	if f := v.FieldByName(fieldName); f.IsValid() {
		return f
	}

	t := v.Type()

	// 2. field whose db or json tag equals name
	if f := fieldByTag(v, t, fieldName); f.IsValid() {
		return f
	}

	// 3. case-insensitive Go field name
	if sf, ok := t.FieldByNameFunc(func(n string) bool {
		return strings.EqualFold(n, fieldName)
	}); ok {
		return v.FieldByIndex(sf.Index)
	}

	return reflect.Value{}
}
