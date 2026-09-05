package permissions

import "reflect"

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

// getField gets a field by name using reflection
func getField(obj interface{}, fieldName string) reflect.Value {
	if obj == nil {
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
	return v.FieldByName(fieldName)
}

