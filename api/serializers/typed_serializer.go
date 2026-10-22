package serializers

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	validation "github.com/forgego/forge/validate"
)

// TypedSerializer provides type-safe serialization for a model type
type TypedSerializer[T any] struct {
	fields    TypedFields[T]
	data      map[string]interface{}
	errors    map[string][]string
	valid     bool
	instance  *T
	configErr error
}

// TypedFields represents type-safe field definitions
type TypedFields[T any] map[string]TypedField[T]

// TypedField represents a single typed field in a serializer
type TypedField[T any] struct {
	fieldExpr    orm.FieldExpression[interface{}]
	fieldIndex   []int
	required     bool
	readOnly     bool
	writeOnly    bool
	validators   []Validator
	defaultValue interface{}
	hasDefault   bool
	schemaField  *schema.Field
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
	tf.hasDefault = true
	return tf
}

// AddField adds a field to the serializer
func (ts *TypedSerializer[T]) AddField(name string, field TypedField[T]) *TypedSerializer[T] {
	ts.fields[name] = field
	return ts
}

// Deserialize deserializes data into a model instance (type-safe)
func (ts *TypedSerializer[T]) Deserialize(data map[string]interface{}, instance *T) error {
	if instance == nil {
		ts.valid = false
		return fmt.Errorf("instance is nil")
	}
	if data == nil {
		data = make(map[string]interface{})
	}
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
	if ts.configErr != nil {
		ts.valid = false
		return ts.configErr
	}

	var schemaValidator *validation.FieldValidator
	for name, field := range ts.fields {
		if field.readOnly {
			continue
		}

		value, exists := ts.data[name]
		if !exists {
			if field.hasDefault {
				value = field.defaultValue
				ts.data[name] = value
			} else if field.required {
				ts.addError(name, "This field is required.")
				continue
			} else {
				continue
			}
		}
		if isNilValue(value) && field.required {
			ts.addError(name, "This field is required.")
			continue
		}
		destination, err := resolveTypedField(ts.instance, field)
		if err != nil {
			ts.addError(name, err.Error())
			continue
		}
		converted, err := valueForField(value, destination.Type())
		if err != nil {
			ts.addError(name, err.Error())
			continue
		}
		if !destination.CanSet() {
			ts.addError(name, "model field cannot be set")
			continue
		}

		// Run validators
		for _, validator := range field.validators {
			if err := validator.Validate(converted.Interface()); err != nil {
				ts.addError(name, err.Error())
			}
		}
		if field.schemaField != nil && !isNilValue(value) {
			if schemaValidator == nil {
				schemaValidator = validation.NewFieldValidator(validation.NewValidator())
			}
			if err := schemaValidator.ValidateField(*field.schemaField, converted.Interface()); err != nil {
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

func isNilValue(value interface{}) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// Serialize serializes a model instance to a map (type-safe)
func (ts *TypedSerializer[T]) Serialize(instance *T) (map[string]interface{}, error) {
	if ts.configErr != nil {
		return nil, ts.configErr
	}
	if instance == nil {
		return nil, fmt.Errorf("instance is nil")
	}
	result := make(map[string]interface{})

	for name, field := range ts.fields {
		if field.writeOnly {
			continue
		}

		fieldValue, err := resolveTypedField(instance, field)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", name, err)
		}
		result[name] = fieldValue.Interface()
	}

	return result, nil
}

// populateInstance populates the model instance from validated data
func (ts *TypedSerializer[T]) populateInstance() error {
	if ts.instance == nil {
		return fmt.Errorf("instance is nil")
	}

	for name, field := range ts.fields {
		if field.readOnly {
			continue
		}

		value, exists := ts.data[name]
		if !exists {
			continue
		}

		fieldValue, err := resolveTypedField(ts.instance, field)
		if err != nil {
			ts.valid = false
			return fmt.Errorf("field %s: %w", name, err)
		}
		converted, err := valueForField(value, fieldValue.Type())
		if err != nil {
			ts.valid = false
			return fmt.Errorf("field %s: %w", name, err)
		}
		fieldValue.Set(converted)
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
