package fields

import (
	"fmt"
	"regexp"
)

// UUIDField represents a UUID field
type UUIDField struct {
	*StringField
}

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// NewUUIDField creates a new UUID field
func NewUUIDField(fieldName string) *UUIDField {
	return &UUIDField{
		StringField: NewStringField(fieldName),
	}
}

// Validate validates the UUID value
func (f *UUIDField) Validate(value interface{}) error {
	if err := f.StringField.Validate(value); err != nil {
		return err
	}

	if value == nil {
		return nil
	}

	uuid, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s: Expected string", f.FieldName)
	}

	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("%s: Enter a valid UUID", f.FieldName)
	}

	return nil
}
