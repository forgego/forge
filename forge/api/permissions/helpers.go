package permissions

import "reflect"

// getMethod gets a method by name using reflection
func getMethod(obj interface{}, methodName string) reflect.Value {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Ptr {
		// Still a pointer, try to get method on pointer
		return reflect.ValueOf(obj).MethodByName(methodName)
	}
	return v.MethodByName(methodName)
}

// getField gets a field by name using reflection
func getField(obj interface{}, fieldName string) reflect.Value {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return v.FieldByName(fieldName)
}
