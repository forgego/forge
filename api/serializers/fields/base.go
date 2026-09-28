package fields

import "fmt"

// Field is the base interface for all serializer fields
type Field interface {
	// ToRepresentation converts internal value to representation
	ToRepresentation(value interface{}) (interface{}, error)
	// ToInternalValue converts representation to internal value
	ToInternalValue(data interface{}) (interface{}, error)
	// Validate validates a value
	Validate(value interface{}) error
	// GetFieldName returns the field name
	GetFieldName() string
	// IsRequired returns whether the field is required
	IsRequired() bool
	// IsReadOnly returns whether the field is read-only
	IsReadOnly() bool
	// IsWriteOnly returns whether the field is write-only
	IsWriteOnly() bool
}

// BaseField provides common field functionality
type BaseField struct {
	FieldName  string
	Required   bool
	ReadOnly   bool
	WriteOnly  bool
	AllowNull  bool
	AllowBlank bool
	Default    interface{}
}

// NewBaseField creates a new base field
func NewBaseField(fieldName string) *BaseField {
	return &BaseField{
		FieldName:  fieldName,
		Required:   false,
		ReadOnly:   false,
		WriteOnly:  false,
		AllowNull:  false,
		AllowBlank: true,
	}
}

// GetFieldName returns the field name
func (f *BaseField) GetFieldName() string {
	return f.FieldName
}

// IsRequired returns whether the field is required
func (f *BaseField) IsRequired() bool {
	return f.Required
}

// IsReadOnly returns whether the field is read-only
func (f *BaseField) IsReadOnly() bool {
	return f.ReadOnly
}

// IsWriteOnly returns whether the field is write-only
func (f *BaseField) IsWriteOnly() bool {
	return f.WriteOnly
}

// ToRepresentation converts internal value to representation (default implementation)
func (f *BaseField) ToRepresentation(value interface{}) (interface{}, error) {
	return value, nil
}

// ToInternalValue converts representation to internal value (default implementation)
func (f *BaseField) ToInternalValue(data interface{}) (interface{}, error) {
	if data == nil && !f.AllowNull {
		if f.Required {
			return nil, fmt.Errorf("%s: This field is required", f.FieldName)
		}
		return f.Default, nil
	}
	return data, nil
}

// Validate validates a value (default implementation)
func (f *BaseField) Validate(value interface{}) error {
	if f.Required && value == nil {
		return fmt.Errorf("%s: This field is required", f.FieldName)
	}
	return nil
}
