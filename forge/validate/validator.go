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

	// Unique validator (recognized for struct tags; full DB uniqueness is verified by ORM/storage layer)
	v.RegisterCustomValidator("unique", validateUnique)
}

// validateUnique recognizes the unique tag so struct validation succeeds without unknown tag errors.
func validateUnique(fl validator.FieldLevel) bool {
	return true
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
func validateChoices(fl validator.FieldLevel) bool {
	param := fl.Param()
	if param == "" {
		return true
	}
	val := fmt.Sprintf("%v", fl.Field().Interface())
	choices := strings.Fields(param)
	for _, c := range choices {
		if c == val {
			return true
		}
	}
	return false
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
		valueStr = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
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
		valueStr = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
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
	if message, ok := validationMessages[fe.Tag()]; ok {
		return message(fe.Field(), fe.Param())
	}
	return fmt.Sprintf("failed validation for tag '%s'", fe.Tag())
}

var validationMessages = map[string]func(string, string) string{
	"required":           constantValidationMessage("is required"),
	"email":              constantValidationMessage("must be a valid email address"),
	"url":                constantValidationMessage("must be a valid URL"),
	"uuid":               constantValidationMessage("must be a valid UUID"),
	"min":                parameterValidationMessage("must be at least %s characters"),
	"max":                parameterValidationMessage("must be at most %s characters"),
	"len":                parameterValidationMessage("must be exactly %s characters"),
	"gte":                parameterValidationMessage("must be greater than or equal to %s"),
	"lte":                parameterValidationMessage("must be less than or equal to %s"),
	"gt":                 parameterValidationMessage("must be greater than %s"),
	"lt":                 parameterValidationMessage("must be less than %s"),
	"eq":                 parameterValidationMessage("must equal %s"),
	"ne":                 parameterValidationMessage("must not equal %s"),
	"oneof":              parameterValidationMessage("must be one of: %s"),
	"numeric":            constantValidationMessage("must be numeric"),
	"alpha":              constantValidationMessage("must contain only letters"),
	"alphanum":           constantValidationMessage("must contain only letters and numbers"),
	"alphaunicode":       constantValidationMessage("must contain only unicode letters"),
	"alphanumunicode":    constantValidationMessage("must contain only unicode letters and numbers"),
	"number":             constantValidationMessage("must be a number"),
	"boolean":            constantValidationMessage("must be a boolean"),
	"datetime":           constantValidationMessage("must be a valid datetime"),
	"date":               constantValidationMessage("must be a valid date"),
	"timezone":           constantValidationMessage("must be a valid timezone"),
	"ip":                 constantValidationMessage("must be a valid IP address"),
	"ipv4":               constantValidationMessage("must be a valid IPv4 address"),
	"ipv6":               constantValidationMessage("must be a valid IPv6 address"),
	"mac":                constantValidationMessage("must be a valid MAC address"),
	"base64":             constantValidationMessage("must be valid base64"),
	"base64url":          constantValidationMessage("must be valid base64url"),
	"json":               constantValidationMessage("must be valid JSON"),
	"jwt":                constantValidationMessage("must be a valid JWT"),
	"hostname":           constantValidationMessage("must be a valid hostname"),
	"fqdn":               constantValidationMessage("must be a valid FQDN"),
	"uri":                constantValidationMessage("must be a valid URI"),
	"url_encoded":        constantValidationMessage("must be URL encoded"),
	"slug":               constantValidationMessage("must be a valid slug (lowercase letters, numbers, hyphens, underscores)"),
	"phone":              constantValidationMessage("must be a valid phone number"),
	"choices":            constantValidationMessage("must be one of the allowed choices"),
	"decimal_max_digits": parameterValidationMessage("must have at most %s digits"),
	"decimal_places":     parameterValidationMessage("must have at most %s decimal places"),
}

func constantValidationMessage(message string) func(string, string) string {
	return func(string, string) string { return message }
}

func parameterValidationMessage(format string) func(string, string) string {
	return func(_ string, param string) string { return fmt.Sprintf(format, param) }
}

// RegisterCustomValidator registers a custom validation function
func (v *Validator) RegisterCustomValidator(tag string, fn validator.Func) error {
	return v.RegisterValidation(tag, fn)
}
