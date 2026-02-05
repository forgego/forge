package api

import (
	"encoding/json"
	"reflect"
	"time"
)

// Serializer is the base interface for all serializers
type Serializer interface {
	// Validate validates the serializer data
	Validate() error
	// ToJSON converts the serializer to JSON
	ToJSON() ([]byte, error)
	// IsValid returns whether the serializer is valid
	IsValid() bool
	// Errors returns validation errors
	Errors() map[string][]string
	// SetData sets the input data for the serializer
	SetData(data map[string]interface{})
	// New creates a new instance of the serializer
	New() Serializer
}

// BaseSerializer provides common serializer functionality
type BaseSerializer struct {
	valid   bool
	errors  map[string][]string
	data    map[string]interface{}
	initial map[string]interface{}
}

// NewBaseSerializer creates a new base serializer
func NewBaseSerializer(data map[string]interface{}) *BaseSerializer {
	return &BaseSerializer{
		valid:   false,
		errors:  make(map[string][]string),
		data:    data,
		initial: make(map[string]interface{}),
	}
}

// Validate validates the serializer (to be overridden)
func (s *BaseSerializer) Validate() error {
	s.valid = true
	return nil
}

// IsValid returns whether the serializer is valid
func (s *BaseSerializer) IsValid() bool {
	return s.valid
}

// Errors returns validation errors
func (s *BaseSerializer) Errors() map[string][]string {
	return s.errors
}

// AddError adds a validation error
func (s *BaseSerializer) AddError(field, message string) {
	if s.errors[field] == nil {
		s.errors[field] = []string{}
	}
	s.errors[field] = append(s.errors[field], message)
	s.valid = false
}

// Get gets a value from the data
func (s *BaseSerializer) Get(key string) interface{} {
	return s.data[key]
}

// GetString gets a string value
func (s *BaseSerializer) GetString(key string) string {
	if val, ok := s.data[key].(string); ok {
		return val
	}
	return ""
}

// GetInt gets an int value
func (s *BaseSerializer) GetInt(key string) int {
	if val, ok := s.data[key].(float64); ok {
		return int(val)
	}
	if val, ok := s.data[key].(int); ok {
		return val
	}
	return 0
}

// GetBool gets a bool value
func (s *BaseSerializer) GetBool(key string) bool {
	if val, ok := s.data[key].(bool); ok {
		return val
	}
	return false
}

// GetTime gets a time value
func (s *BaseSerializer) GetTime(key string) (time.Time, error) {
	if val, ok := s.data[key].(string); ok {
		return time.Parse(time.RFC3339, val)
	}
	return time.Time{}, nil
}

// Set sets a value in the data
func (s *BaseSerializer) Set(key string, value interface{}) {
	s.data[key] = value
}

// SetData sets the entire data map
func (s *BaseSerializer) SetData(data map[string]interface{}) {
	s.data = data
}

// ToJSON converts the serializer to JSON
func (s *BaseSerializer) ToJSON() ([]byte, error) {
	return json.Marshal(s.data)
}

// New creates a new base serializer
func (s *BaseSerializer) New() Serializer {
	return NewBaseSerializer(nil)
}

// ModelSerializer is a serializer for models
type ModelSerializer struct {
	*BaseSerializer
	model interface{}
}

// NewModelSerializer creates a new model serializer
func NewModelSerializer(model interface{}) *ModelSerializer {
	data := modelToMap(model)
	return &ModelSerializer{
		BaseSerializer: NewBaseSerializer(data),
		model:          model,
	}
}

// modelToMap converts a model struct to a map
func modelToMap(model interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Get JSON tag or use field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		key := jsonTag
		if key == "" {
			key = field.Name
		}

		// Handle different types
		switch value.Kind() {
		case reflect.String:
			result[key] = value.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[key] = value.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[key] = value.Uint()
		case reflect.Float32, reflect.Float64:
			result[key] = value.Float()
		case reflect.Bool:
			result[key] = value.Bool()
		case reflect.Struct:
			if value.Type() == reflect.TypeOf(time.Time{}) {
				if timeVal, ok := value.Interface().(time.Time); ok {
					result[key] = timeVal.Format(time.RFC3339)
				}
			} else {
				result[key] = modelToMap(value.Interface())
			}
		case reflect.Ptr:
			if !value.IsNil() {
				result[key] = modelToMap(value.Elem().Interface())
			}
		default:
			result[key] = value.Interface()
		}
	}

	return result
}

// SerializeModel serializes a model to a map
func SerializeModel(model interface{}) map[string]interface{} {
	return modelToMap(model)
}

// SerializeMany serializes multiple models
func SerializeMany(models interface{}) []map[string]interface{} {
	result := []map[string]interface{}{}
	v := reflect.ValueOf(models)

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			result = append(result, modelToMap(v.Index(i).Interface()))
		}
	}

	return result
}
