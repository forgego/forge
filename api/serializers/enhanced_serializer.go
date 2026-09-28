package serializers

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/api/serializers/fields"
)

// EnhancedSerializer provides enhanced serialization with field types
type EnhancedSerializer struct {
	Fields map[string]fields.Field
	Data   map[string]interface{}
	Errors map[string][]string
	Valid  bool
}

// NewEnhancedSerializer creates a new enhanced serializer
func NewEnhancedSerializer() *EnhancedSerializer {
	return &EnhancedSerializer{
		Fields: make(map[string]fields.Field),
		Data:   make(map[string]interface{}),
		Errors: make(map[string][]string),
		Valid:  false,
	}
}

// AddField adds a field to the serializer
func (s *EnhancedSerializer) AddField(name string, field fields.Field) {
	s.Fields[name] = field
}

// Validate validates all fields
func (s *EnhancedSerializer) Validate() error {
	s.Valid = true
	s.Errors = make(map[string][]string)

	for name, field := range s.Fields {
		if field.IsReadOnly() {
			continue
		}

		value := s.Data[name]
		if value == nil && !field.IsRequired() {
			continue
		}

		// Convert to internal value
		internalValue, err := field.ToInternalValue(value)
		if err != nil {
			s.addError(name, err.Error())
			continue
		}

		// Validate
		if err := field.Validate(internalValue); err != nil {
			s.addError(name, err.Error())
			continue
		}

		// Store validated value
		s.Data[name] = internalValue
	}

	if len(s.Errors) > 0 {
		s.Valid = false
		return fmt.Errorf("validation failed")
	}

	return nil
}

// addError adds a validation error
func (s *EnhancedSerializer) addError(field, message string) {
	if s.Errors[field] == nil {
		s.Errors[field] = []string{}
	}
	s.Errors[field] = append(s.Errors[field], message)
	s.Valid = false
}

// IsValid returns whether the serializer is valid
func (s *EnhancedSerializer) IsValid() bool {
	return s.Valid
}

// GetErrors returns validation errors
func (s *EnhancedSerializer) GetErrors() map[string][]string {
	return s.Errors
}

// ToRepresentation converts data to representation
func (s *EnhancedSerializer) ToRepresentation(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Use reflection to get field values
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result, nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldName := jsonTag
		if serializerField, ok := s.Fields[fieldName]; ok {
			if serializerField.IsWriteOnly() {
				continue
			}

			repr, err := serializerField.ToRepresentation(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			result[fieldName] = repr
		} else {
			// No field definition, use value as-is
			result[fieldName] = fieldValue.Interface()
		}
	}

	return result, nil
}
