package validation

import (
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type messageFieldError struct{ tag, param string }

func (e messageFieldError) Tag() string                    { return e.tag }
func (e messageFieldError) ActualTag() string              { return e.tag }
func (e messageFieldError) Namespace() string              { return "" }
func (e messageFieldError) StructNamespace() string        { return "" }
func (e messageFieldError) Field() string                  { return "" }
func (e messageFieldError) StructField() string            { return "" }
func (e messageFieldError) Value() interface{}             { return nil }
func (e messageFieldError) Param() string                  { return e.param }
func (e messageFieldError) Kind() reflect.Kind             { return reflect.Invalid }
func (e messageFieldError) Type() reflect.Type             { return nil }
func (e messageFieldError) Translate(ut.Translator) string { return "" }
func (e messageFieldError) Error() string                  { return "" }

var _ validator.FieldError = messageFieldError{}

func TestGetErrorMessage_Characterization(t *testing.T) {
	constantCases := map[string]string{
		"required": "is required", "email": "must be a valid email address", "url": "must be a valid URL", "uuid": "must be a valid UUID", "numeric": "must be numeric", "alpha": "must contain only letters", "alphanum": "must contain only letters and numbers", "alphaunicode": "must contain only unicode letters", "alphanumunicode": "must contain only unicode letters and numbers", "number": "must be a number", "boolean": "must be a boolean", "datetime": "must be a valid datetime", "date": "must be a valid date", "timezone": "must be a valid timezone", "ip": "must be a valid IP address", "ipv4": "must be a valid IPv4 address", "ipv6": "must be a valid IPv6 address", "mac": "must be a valid MAC address", "base64": "must be valid base64", "base64url": "must be valid base64url", "json": "must be valid JSON", "jwt": "must be a valid JWT", "hostname": "must be a valid hostname", "fqdn": "must be a valid FQDN", "uri": "must be a valid URI", "url_encoded": "must be URL encoded", "slug": "must be a valid slug (lowercase letters, numbers, hyphens, underscores)", "phone": "must be a valid phone number", "choices": "must be one of the allowed choices",
	}
	parameterCases := map[string]string{
		"min": "must be at least 12 characters", "max": "must be at most 12 characters", "len": "must be exactly 12 characters", "gte": "must be greater than or equal to 12", "lte": "must be less than or equal to 12", "gt": "must be greater than 12", "lt": "must be less than 12", "eq": "must equal 12", "ne": "must not equal 12", "oneof": "must be one of: 12", "decimal_max_digits": "must have at most 12 digits", "decimal_places": "must have at most 12 decimal places",
	}
	for tag, want := range constantCases {
		t.Run(tag, func(t *testing.T) { assert.Equal(t, want, getErrorMessage(messageFieldError{tag: tag})) })
	}
	for tag, want := range parameterCases {
		t.Run(tag, func(t *testing.T) { assert.Equal(t, want, getErrorMessage(messageFieldError{tag: tag, param: "12"})) })
	}
	assert.Equal(t, "failed validation for tag 'unknown'", getErrorMessage(messageFieldError{tag: "unknown"}))
}
