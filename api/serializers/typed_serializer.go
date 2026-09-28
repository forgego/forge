package serializers

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/orm"
)

// TypedSerializer provides type-safe serialization for a model type
type TypedSerializer[T any] struct {
	fields   TypedFields[T]
	data     map[string]interface{}
	errors   map[string][]string
	valid    bool
	instance *T
}

// TypedFields represents type-safe field definitions
type TypedFields[T any] map[string]TypedField[T]

// TypedField represents a single typed field in a serializer
type TypedField[T any] struct {
	fieldExpr    orm.FieldExpression[interface{}]
	required     bool
	readOnly     bool
	writeOnly    bool
	validators   []Validator
	defaultValue interface{}
}

// NewTypedSerializer creates a new type-safe serializer
func NewTypedSerializer[T any]() *TypedSerializer[T] {
	return &TypedSerializer[T]{
		fields: make(TypedFields[T]),
		data:   make(map[string]interface{}),
		errors: make(map[string][]string),
		valid:  false,
	}
}

// Field creates a typed field definition
func Field[T any, V any](fieldExpr orm.FieldExpression[V]) TypedField[T] {
	return TypedField[T]{
		fieldExpr:  convertFieldExpr(fieldExpr),
		required:   false,
		readOnly:   false,
		writeOnly:  false,
		validators: []Validator{},
	}
}

// convertFieldExpr converts a typed FieldExpression to interface{} for storage
func convertFieldExpr[V any](expr orm.FieldExpression[V]) orm.FieldExpression[interface{}] {
	// This is a type erasure - we store as interface{} but validate at runtime
	// The type safety is maintained through the generic constraint on Field()
	return orm.NewField[interface{}](expr.Path(), expr.Table())
}

// Required marks a field as required
func (tf TypedField[T]) Required() TypedField[T] {
	tf.required = true
	return tf
}

// ReadOnly marks a field as read-only
func (tf TypedField[T]) ReadOnly() TypedField[T] {
	tf.readOnly = true
	return tf
}

// WriteOnly marks a field as write-only
func (tf TypedField[T]) WriteOnly() TypedField[T] {
	tf.writeOnly = true
	return tf
}

// Default sets a default value
func (tf TypedField[T]) Default(value interface{}) TypedField[T] {
	tf.defaultValue = value
	return tf
}

// AddField adds a field to the serializer
func (ts *TypedSerializer[T]) AddField(name string, field TypedField[T]) *TypedSerializer[T] {
	ts.fields[name] = field
	return ts
}

// Deserialize deserializes data into a model instance (type-safe)
func (ts *TypedSerializer[T]) Deserialize(data map[string]interface{}, instance *T) error {
	ts.data = data
	ts.instance = instance

	// Validate
	if err := ts.Validate(); err != nil {
		return err
	}

	// Populate instance
	return ts.populateInstance()
}

// Validate validates the serializer data
func (ts *TypedSerializer[T]) Validate() error {
	ts.errors = make(map[string][]string)
	ts.valid = true

	for name, field := range ts.fields {
		if field.readOnly {
			continue
		}

		value, exists := ts.data[name]
		if !exists {
			if field.required {
				ts.addError(name, "This field is required.")
				continue
			}
			// Use default if available
			if field.defaultValue != nil {
				ts.data[name] = field.defaultValue
			}
			continue
		}

		// Run validators
		for _, validator := range field.validators {
			if err := validator.Validate(value); err != nil {
				ts.addError(name, err.Error())
			}
		}
	}

	if len(ts.errors) > 0 {
		ts.valid = false
		return fmt.Errorf("validation failed")
	}

	return nil
}

// Serialize serializes a model instance to a map (type-safe)
func (ts *TypedSerializer[T]) Serialize(instance *T) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Use reflection to get field values
	v := reflect.ValueOf(instance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for name, field := range ts.fields {
		if field.writeOnly {
			continue
		}

		// Get field value using field expression path
		fieldName := field.fieldExpr.Path()
		fieldValue := v.FieldByName(fieldName)

		if fieldValue.IsValid() && fieldValue.CanInterface() {
			result[name] = fieldValue.Interface()
		}
	}

	return result, nil
}

// populateInstance populates the model instance from validated data
func (ts *TypedSerializer[T]) populateInstance() error {
	if ts.instance == nil {
		return fmt.Errorf("instance is nil")
	}

	v := reflect.ValueOf(ts.instance)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for name, field := range ts.fields {
		if field.readOnly {
			continue
		}

		value, exists := ts.data[name]
		if !exists {
			continue
		}

		// Set field value
		fieldName := field.fieldExpr.Path()
		fieldValue := v.FieldByName(fieldName)

		if fieldValue.IsValid() && fieldValue.CanSet() {
			valueReflect := reflect.ValueOf(value)
			if valueReflect.Type().AssignableTo(fieldValue.Type()) {
				fieldValue.Set(valueReflect)
			} else if valueReflect.Type().ConvertibleTo(fieldValue.Type()) {
				fieldValue.Set(valueReflect.Convert(fieldValue.Type()))
			}
		}
	}

	return nil
}

// addError adds a validation error
func (ts *TypedSerializer[T]) addError(field, message string) {
	if ts.errors[field] == nil {
		ts.errors[field] = []string{}
	}
	ts.errors[field] = append(ts.errors[field], message)
	ts.valid = false
}

// IsValid returns whether the serializer is valid
func (ts *TypedSerializer[T]) IsValid() bool {
	return ts.valid
}

// Errors returns validation errors
func (ts *TypedSerializer[T]) Errors() map[string][]string {
	return ts.errors
}

// Validator is the interface for field validators
type Validator interface {
	Validate(value interface{}) error
}
