package validation

import (
	"fmt"
	"reflect"

	"github.com/forgego/forge/orm"
)

// TypedValidatorBuilder provides type-safe validator construction
type TypedValidatorBuilder[T any] struct {
	fields map[string]*FieldValidatorBuilder[T]
	validator *Validator
}

// For creates a new typed validator builder for a model type
func For[T any]() *TypedValidatorBuilder[T] {
	return &TypedValidatorBuilder[T]{
		fields:    make(map[string]*FieldValidatorBuilder[T]),
		validator: NewValidator(),
	}
}

// FieldFor starts building a validator for a specific field
// Usage: builder.FieldFor(User.Fields.Email)
func FieldFor[T any, V any](tvb *TypedValidatorBuilder[T], fieldExpr orm.FieldExpression[V]) *FieldValidatorBuilder[T] {
	fieldName := fieldExpr.Path()
	if tvb.fields[fieldName] == nil {
		tvb.fields[fieldName] = &FieldValidatorBuilder[T]{
			fieldName: fieldName,
			parent:    tvb,
			rules:     []string{},
		}
	}
	return tvb.fields[fieldName]
}

// Build builds the validator
func (tvb *TypedValidatorBuilder[T]) Build() *TypedValidator[T] {
	return &TypedValidator[T]{
		fields:    tvb.fields,
		validator: tvb.validator,
	}
}

// FieldValidatorBuilder provides fluent API for field validation rules
type FieldValidatorBuilder[T any] struct {
	fieldName string
	parent    *TypedValidatorBuilder[T]
	rules     []string
	min       *float64
	max       *float64
	minLength *int
	maxLength *int
	choices   []interface{}
}

// Required marks the field as required
func (fvb *FieldValidatorBuilder[T]) Required() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "required")
	return fvb
}

// Email validates email format
func (fvb *FieldValidatorBuilder[T]) Email() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "email")
	return fvb
}

// URL validates URL format
func (fvb *FieldValidatorBuilder[T]) URL() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "url")
	return fvb
}

// UUID validates UUID format
func (fvb *FieldValidatorBuilder[T]) UUID() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "uuid")
	return fvb
}

// Min sets minimum value (for numeric fields)
func (fvb *FieldValidatorBuilder[T]) Min(value float64) *FieldValidatorBuilder[T] {
	fvb.min = &value
	fvb.rules = append(fvb.rules, fmt.Sprintf("gte=%g", value))
	return fvb
}

// Max sets maximum value (for numeric fields)
func (fvb *FieldValidatorBuilder[T]) Max(value float64) *FieldValidatorBuilder[T] {
	fvb.max = &value
	fvb.rules = append(fvb.rules, fmt.Sprintf("lte=%g", value))
	return fvb
}

// MinLength sets minimum length (for string fields)
func (fvb *FieldValidatorBuilder[T]) MinLength(length int) *FieldValidatorBuilder[T] {
	fvb.minLength = &length
	fvb.rules = append(fvb.rules, fmt.Sprintf("min=%d", length))
	return fvb
}

// MaxLength sets maximum length (for string fields)
func (fvb *FieldValidatorBuilder[T]) MaxLength(length int) *FieldValidatorBuilder[T] {
	fvb.maxLength = &length
	fvb.rules = append(fvb.rules, fmt.Sprintf("max=%d", length))
	return fvb
}

// Range sets both min and max
func (fvb *FieldValidatorBuilder[T]) Range(min, max float64) *FieldValidatorBuilder[T] {
	return fvb.Min(min).Max(max)
}

// Length sets exact length
func (fvb *FieldValidatorBuilder[T]) Length(length int) *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, fmt.Sprintf("len=%d", length))
	return fvb
}

// Choices sets allowed choices
func (fvb *FieldValidatorBuilder[T]) Choices(choices ...interface{}) *FieldValidatorBuilder[T] {
	fvb.choices = choices
	// Build oneof tag
	choiceStrs := make([]string, len(choices))
	for i, choice := range choices {
		choiceStrs[i] = fmt.Sprintf("%v", choice)
	}
	fvb.rules = append(fvb.rules, fmt.Sprintf("oneof=%s", joinStrings(choiceStrs, " ")))
	return fvb
}

// Unique marks field as unique (requires database check)
func (fvb *FieldValidatorBuilder[T]) Unique() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "unique")
	return fvb
}

// Numeric validates numeric format
func (fvb *FieldValidatorBuilder[T]) Numeric() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "numeric")
	return fvb
}

// Alpha validates alphabetic characters only
func (fvb *FieldValidatorBuilder[T]) Alpha() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "alpha")
	return fvb
}

// Alphanum validates alphanumeric characters
func (fvb *FieldValidatorBuilder[T]) Alphanum() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "alphanum")
	return fvb
}

// Slug validates slug format
func (fvb *FieldValidatorBuilder[T]) Slug() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "slug")
	return fvb
}

// Phone validates phone number format
func (fvb *FieldValidatorBuilder[T]) Phone() *FieldValidatorBuilder[T] {
	fvb.rules = append(fvb.rules, "phone")
	return fvb
}

// ComplexityRules adds password complexity rules
func (fvb *FieldValidatorBuilder[T]) ComplexityRules() *FieldValidatorBuilder[T] {
	// Add multiple rules for password complexity
	fvb.MinLength(8)
	fvb.rules = append(fvb.rules, "alphanum") // At least alphanumeric
	return fvb
}

// Back returns to the parent validator builder
func (fvb *FieldValidatorBuilder[T]) Back() *TypedValidatorBuilder[T] {
	return fvb.parent
}

// TypedValidator provides type-safe validation
type TypedValidator[T any] struct {
	fields    map[string]*FieldValidatorBuilder[T]
	validator *Validator
}

// Validate validates a model instance
func (tv *TypedValidator[T]) Validate(instance *T) error {
	// Build validation tags dynamically
	instanceValue := reflect.ValueOf(instance)
	if instanceValue.Kind() == reflect.Ptr {
		instanceValue = instanceValue.Elem()
	}

	// Validate each field
	for fieldName, fieldBuilder := range tv.fields {
		fieldValue := instanceValue.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			// Try lowercase
			fieldValue = instanceValue.FieldByName(toPascalCase(fieldName))
		}

		if !fieldValue.IsValid() {
			continue
		}

		// Build validation tag
		tag := joinStrings(fieldBuilder.rules, ",")
		if tag != "" {
			if err := tv.validator.ValidateField(fieldValue.Interface(), tag); err != nil {
				return fmt.Errorf("field %s: %w", fieldName, err)
			}
		}
	}

	// Also run struct-level validation
	return tv.validator.ValidateStruct(instance)
}

// OverrideField allows overriding validation for a specific field
func OverrideField[T any, V any](tv *TypedValidator[T], fieldExpr orm.FieldExpression[V], builder *FieldValidatorBuilder[T]) *TypedValidator[T] {
	fieldName := fieldExpr.Path()
	tv.fields[fieldName] = builder
	return tv
}

// Helper functions
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

func toPascalCase(s string) string {
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
