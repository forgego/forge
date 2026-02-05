package utils

import (
	"fmt"
	"reflect"
)

// GetIDFromInstance extracts ID from an instance using reflection
func GetIDFromInstance(instance interface{}) int64 {
	if instance == nil {
		return 0
	}
	
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	idField := val.FieldByName("ID")
	if !idField.IsValid() {
		return 0
	}
	
	switch idField.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return idField.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(idField.Uint())
	}
	
	return 0
}

// GetFieldValue gets a field value from an instance using reflection
func GetFieldValue(instance interface{}, fieldName string) interface{} {
	if instance == nil {
		return nil
	}
	
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	
	if field.CanInterface() {
		return field.Interface()
	}
	
	return nil
}

// SetFieldValue sets a field value on an instance using reflection
func SetFieldValue(instance interface{}, fieldName string, value interface{}) error {
	if instance == nil {
		return fmt.Errorf("instance is nil")
	}
	
	val := reflect.ValueOf(instance)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("instance must be a pointer")
	}
	
	val = val.Elem()
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}
	
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}
	
	valueVal := reflect.ValueOf(value)
	if valueVal.Type().AssignableTo(field.Type()) {
		field.Set(valueVal)
		return nil
	}
	
	if valueVal.Type().ConvertibleTo(field.Type()) {
		field.Set(valueVal.Convert(field.Type()))
		return nil
	}
	
	return fmt.Errorf("cannot assign value of type %v to field %s of type %v",
		valueVal.Type(), fieldName, field.Type())
}

// HasField checks if an instance has a field
func HasField(instance interface{}, fieldName string) bool {
	if instance == nil {
		return false
	}
	
	val := reflect.ValueOf(instance)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	field := val.FieldByName(fieldName)
	return field.IsValid()
}

// GetModelName gets the model name from a type
func GetModelName(instance interface{}) string {
	if instance == nil {
		return ""
	}
	
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	
	return typ.Name()
}
