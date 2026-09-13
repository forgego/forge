package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/forgego/forge/schema"
)

// FieldValidator validates a field value against a schema.Field definition
type FieldValidator struct {
	validator *Validator
}

// NewFieldValidator creates a new field validator
func NewFieldValidator(v *Validator) *FieldValidator {
	return &FieldValidator{validator: v}
}

// ValidateField validates a field value against its schema definition
func (fv *FieldValidator) ValidateField(field schema.Field, value interface{}) error {
	// Check required
	if field.Required {
		if value == nil || isEmpty(value) {
			return fmt.Errorf("%s: is required", field.Name)
		}
	}

	// If value is empty and field allows blank, skip other validations
	if isEmpty(value) && field.Blank {
		return nil
	}

	// Validate choices if defined
	if len(field.Choices) > 0 {
		if err := fv.validateChoices(field, value); err != nil {
			return err
		}
	}

	// Validate length constraints (for strings, bytes, arrays)
	if field.MinLength != nil || field.MaxLength != nil {
		if err := fv.validateLength(field, value); err != nil {
			return err
		}
	}

	// Validate numeric value constraints
	if field.MinValue != nil || field.MaxValue != nil {
		if err := fv.validateNumericRange(field, value); err != nil {
			return err
		}
	}

	// Validate decimal constraints
	if field.Type == schema.TypeDecimal {
		if err := fv.validateDecimal(field, value); err != nil {
			return err
		}
	}

	// Run custom validators
	for _, validator := range field.Validators {
		if err := validator.Validate(value); err != nil {
			return fmt.Errorf("%s: %v", field.Name, err)
		}
	}

	// Validate using struct tag if available
	if field.ValidationTag != "" {
		if err := fv.validator.ValidateField(value, field.ValidationTag); err != nil {
			return fmt.Errorf("%s: %v", field.Name, err)
		}
	}

	return nil
}

// validateChoices validates that a value is one of the allowed choices
func (fv *FieldValidator) validateChoices(field schema.Field, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)

	for _, choice := range field.Choices {
		if choice.Value == valueStr {
			return nil
		}
	}

	// Build list of valid choices for error message
	validChoices := make([]string, len(field.Choices))
	for i, choice := range field.Choices {
		validChoices[i] = choice.Value
	}

	return fmt.Errorf("%s: must be one of: %v", field.Name, validChoices)
}

// validateLength validates string/array length constraints
func (fv *FieldValidator) validateLength(field schema.Field, value interface{}) error {
	var length int

	switch v := value.(type) {
	case string:
		length = len(v)
	case []byte:
		length = len(v)
	default:
		// Try to get length via reflection
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array || rv.Kind() == reflect.String {
			length = rv.Len()
		} else {
			return nil // Not a length-validatable type
		}
	}

	if field.MinLength != nil && length < *field.MinLength {
		return fmt.Errorf("%s: must be at least %d characters", field.Name, *field.MinLength)
	}

	if field.MaxLength != nil && length > *field.MaxLength {
		return fmt.Errorf("%s: must be at most %d characters", field.Name, *field.MaxLength)
	}

	return nil
}

// validateNumericRange validates numeric value constraints
func (fv *FieldValidator) validateNumericRange(field schema.Field, value interface{}) error {
	var numValue float64

	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int32:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float32:
		numValue = float64(v)
	case float64:
		numValue = v
	default:
		// Try to convert via reflection
		rv := reflect.ValueOf(value)
		if rv.Kind() >= reflect.Int && rv.Kind() <= reflect.Int64 {
			numValue = float64(rv.Int())
		} else if rv.Kind() >= reflect.Uint && rv.Kind() <= reflect.Uint64 {
			numValue = float64(rv.Uint())
		} else if rv.Kind() == reflect.Float32 || rv.Kind() == reflect.Float64 {
			numValue = rv.Float()
		} else {
			return nil // Not a numeric type
		}
	}

	if field.MinValue != nil && numValue < *field.MinValue {
		return fmt.Errorf("%s: must be at least %g", field.Name, *field.MinValue)
	}

	if field.MaxValue != nil && numValue > *field.MaxValue {
		return fmt.Errorf("%s: must be at most %g", field.Name, *field.MaxValue)
	}

	return nil
}

// validateDecimal validates decimal-specific constraints
func (fv *FieldValidator) validateDecimal(field schema.Field, value interface{}) error {
	// Convert value to string to analyze digits
	valueStr := fmt.Sprintf("%v", value)

	// Count total digits (excluding decimal point)
	totalDigits := 0
	decimalPlaces := 0
	foundDecimal := false

	for _, r := range valueStr {
		if r >= '0' && r <= '9' {
			totalDigits++
			if foundDecimal {
				decimalPlaces++
			}
		} else if r == '.' {
			foundDecimal = true
		}
	}

	if field.MaxDigits != nil && totalDigits > *field.MaxDigits {
		return fmt.Errorf("%s: must have at most %d digits", field.Name, *field.MaxDigits)
	}

	if field.DecimalPlaces != nil && decimalPlaces > *field.DecimalPlaces {
		return fmt.Errorf("%s: must have at most %d decimal places", field.Name, *field.DecimalPlaces)
	}

	return nil
}

// isEmpty checks if a value is empty
func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return rv.Len() == 0
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// ValidateModel validates an entire model instance against its schema
func (fv *FieldValidator) ValidateModel(model interface{}, fields []schema.Field) error {
	rv := reflect.ValueOf(model)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct")
	}

	rt := rv.Type()

	// Create a map of field names to schema fields
	fieldMap := make(map[string]schema.Field)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	// Validate each struct field
	for i := 0; i < rv.NumField(); i++ {
		structField := rt.Field(i)
		fieldValue := rv.Field(i)

		// Get field name from struct tag or use field name
		fieldName := structField.Name
		if jsonTag := structField.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// Extract field name from json tag (handle "name,omitempty")
			parts := strings.Split(jsonTag, ",")
			fieldName = parts[0]
		}

		// Find corresponding schema field
		schemaField, exists := fieldMap[fieldName]
		if !exists {
			continue // Skip fields not in schema
		}

		// Get field value
		var value interface{}
		if fieldValue.CanInterface() {
			value = fieldValue.Interface()
		} else {
			continue
		}

		// Validate the field
		if err := fv.ValidateField(schemaField, value); err != nil {
			return err
		}
	}

	return nil
}
