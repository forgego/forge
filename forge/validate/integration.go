package validation

import (
	"fmt"
	"strings"

	"github.com/forgego/forge/schema"
)

// GenerateValidationTag generates a validation tag for a field
func GenerateValidationTag(field schema.Field) string {
	var tags []string

	// Required
	if field.Required {
		tags = append(tags, "required")
	}

	// Type-specific validations
	switch field.Type {
	case schema.TypeEmail:
		tags = append(tags, "email")
	case schema.TypeURL:
		tags = append(tags, "url")
	case schema.TypeUUID:
		tags = append(tags, "uuid")
	}

	// Length validations (for strings, bytes, arrays)
	if field.MaxLength != nil {
		tags = append(tags, fmt.Sprintf("max=%d", *field.MaxLength))
	}
	if field.MinLength != nil {
		tags = append(tags, fmt.Sprintf("min=%d", *field.MinLength))
	}

	// Numeric value validations
	if field.MaxValue != nil {
		tags = append(tags, fmt.Sprintf("lte=%g", *field.MaxValue))
	}
	if field.MinValue != nil {
		tags = append(tags, fmt.Sprintf("gte=%g", *field.MinValue))
	}

	// Decimal validations
	if field.Type == schema.TypeDecimal {
		if field.MaxDigits != nil && field.DecimalPlaces != nil {
			// For decimal, we validate the total digits and decimal places
			// This is a custom validation that would need to be registered
			tags = append(tags, fmt.Sprintf("decimal_max_digits=%d", *field.MaxDigits))
			tags = append(tags, fmt.Sprintf("decimal_places=%d", *field.DecimalPlaces))
		}
	}

	// Choices validation (if choices are defined)
	if len(field.Choices) > 0 {
		// Create a oneof validation tag with all choice values
		choiceValues := make([]string, len(field.Choices))
		for i, choice := range field.Choices {
			// Escape choice values if they contain spaces or special characters
			choiceValues[i] = choice.Value
		}
		// Use oneof tag with all choice values
		oneofTag := fmt.Sprintf("oneof=%s", strings.Join(choiceValues, " "))
		tags = append(tags, oneofTag)
	}

	// Use custom validation tag if provided
	if field.ValidationTag != "" {
		return field.ValidationTag
	}

	// Join tags
	if len(tags) == 0 {
		return ""
	}

	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += "," + tags[i]
	}

	return result
}

// ValidateModel validates a model instance using the validator
func ValidateModel(v *Validator, instance interface{}) error {
	return v.ValidateStruct(instance)
}

// ValidateModelWithSchema validates a model instance against its schema fields
func ValidateModelWithSchema(v *Validator, instance interface{}, fields []schema.Field) error {
	fv := NewFieldValidator(v)
	return fv.ValidateModel(instance, fields)
}
