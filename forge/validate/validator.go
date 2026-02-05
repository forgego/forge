package validation

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator with framework-specific methods
type Validator struct {
	*validator.Validate
}

// NewValidator creates a new validator instance with all custom validators registered
func NewValidator() *Validator {
	v := validator.New()
	val := &Validator{Validate: v}

	// Register all custom validators
	val.registerCustomValidators()

	return val
}

// registerCustomValidators registers all custom validation functions
func (v *Validator) registerCustomValidators() {
	// Slug validator
	v.RegisterCustomValidator("slug", validateSlug)

	// Phone validator
	v.RegisterCustomValidator("phone", validatePhone)

	// Choices validator (validates against a set of allowed values)
	v.RegisterCustomValidator("choices", validateChoices)

	// Decimal validators
	v.RegisterCustomValidator("decimal_max_digits", validateDecimalMaxDigits)
	v.RegisterCustomValidator("decimal_places", validateDecimalPlaces)
}

// validateSlug validates that a string is a valid slug
func validateSlug(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true // Empty is handled by required validator
	}

	// Slug should only contain lowercase letters, numbers, hyphens, and underscores
	for _, r := range str {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// validatePhone validates that a string is a valid phone number
func validatePhone(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	if str == "" {
		return true // Empty is handled by required validator
	}

	// Basic phone validation - allows digits, spaces, hyphens, parentheses, and +
	hasDigit := false
	for _, r := range str {
		if r >= '0' && r <= '9' {
			hasDigit = true
		} else if r != ' ' && r != '-' && r != '(' && r != ')' && r != '+' {
			return false
		}
	}
	return hasDigit
}

// validateChoices validates that a value is one of the allowed choices
// This is a fallback validator. The oneof tag from go-playground/validator
// should be used instead when choices are known at validation tag generation time.
func validateChoices(fl validator.FieldLevel) bool {
	// This validator is a fallback - in practice, use the "oneof" tag with choice values
	// This will always pass - actual validation should use oneof
	return true
}

// validateDecimalMaxDigits validates that a decimal number has at most N digits
func validateDecimalMaxDigits(fl validator.FieldLevel) bool {
	param := fl.Param()
	maxDigits, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	// Get the field value as string to count digits
	fieldValue := fl.Field()
	var valueStr string

	switch fieldValue.Kind() {
	case reflect.Float32, reflect.Float64:
		valueStr = fmt.Sprintf("%g", fieldValue.Float())
	case reflect.String:
		valueStr = fieldValue.String()
	default:
		return false
	}

	// Remove decimal point and count digits
	digits := 0
	for _, r := range valueStr {
		if r >= '0' && r <= '9' {
			digits++
		}
	}

	return digits <= maxDigits
}

// validateDecimalPlaces validates that a decimal number has at most N decimal places
func validateDecimalPlaces(fl validator.FieldLevel) bool {
	param := fl.Param()
	maxPlaces, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	// Get the field value as string
	fieldValue := fl.Field()
	var valueStr string

	switch fieldValue.Kind() {
	case reflect.Float32, reflect.Float64:
		valueStr = fmt.Sprintf("%g", fieldValue.Float())
	case reflect.String:
		valueStr = fieldValue.String()
	default:
		return false
	}

	// Find decimal point and count places after it
	parts := strings.Split(valueStr, ".")
	if len(parts) == 2 {
		return len(parts[1]) <= maxPlaces
	}

	return true // No decimal point means 0 places
}

// ValidateStruct validates a struct
func (v *Validator) ValidateStruct(s interface{}) error {
	if err := v.Struct(s); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	if err := v.Var(field, tag); err != nil {
		return formatValidationErrors(err)
	}
	return nil
}

// formatValidationErrors formats validator errors into a readable format
func formatValidationErrors(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errMsg string
		for i, fe := range validationErrors {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += fmt.Sprintf("%s: %s", fe.Field(), getErrorMessage(fe))
		}
		return fmt.Errorf("validation failed: %s", errMsg)
	}
	return err
}

// getErrorMessage returns a user-friendly error message for a validation error
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "url":
		return "must be a valid URL"
	case "uuid":
		return "must be a valid UUID"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "lt":
		return fmt.Sprintf("must be less than %s", fe.Param())
	case "eq":
		return fmt.Sprintf("must equal %s", fe.Param())
	case "ne":
		return fmt.Sprintf("must not equal %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	case "numeric":
		return "must be numeric"
	case "alpha":
		return "must contain only letters"
	case "alphanum":
		return "must contain only letters and numbers"
	case "alphaunicode":
		return "must contain only unicode letters"
	case "alphanumunicode":
		return "must contain only unicode letters and numbers"
	case "number":
		return "must be a number"
	case "boolean":
		return "must be a boolean"
	case "datetime":
		return "must be a valid datetime"
	case "date":
		return "must be a valid date"
	case "timezone":
		return "must be a valid timezone"
	case "ip":
		return "must be a valid IP address"
	case "ipv4":
		return "must be a valid IPv4 address"
	case "ipv6":
		return "must be a valid IPv6 address"
	case "mac":
		return "must be a valid MAC address"
	case "base64":
		return "must be valid base64"
	case "base64url":
		return "must be valid base64url"
	case "json":
		return "must be valid JSON"
	case "jwt":
		return "must be a valid JWT"
	case "hostname":
		return "must be a valid hostname"
	case "fqdn":
		return "must be a valid FQDN"
	case "uri":
		return "must be a valid URI"
	case "url_encoded":
		return "must be URL encoded"
	case "slug":
		return "must be a valid slug (lowercase letters, numbers, hyphens, underscores)"
	case "phone":
		return "must be a valid phone number"
	case "choices":
		return "must be one of the allowed choices"
	case "decimal_max_digits":
		return fmt.Sprintf("must have at most %s digits", fe.Param())
	case "decimal_places":
		return fmt.Sprintf("must have at most %s decimal places", fe.Param())
	default:
		return fmt.Sprintf("failed validation for tag '%s'", fe.Tag())
	}
}

// RegisterCustomValidator registers a custom validation function
func (v *Validator) RegisterCustomValidator(tag string, fn validator.Func) error {
	return v.RegisterValidation(tag, fn)
}
