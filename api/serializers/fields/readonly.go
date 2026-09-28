package fields

// ReadOnlyField represents a read-only field
type ReadOnlyField struct {
	*BaseField
	SourceField Field
}

// NewReadOnlyField creates a new read-only field
func NewReadOnlyField(fieldName string, sourceField Field) *ReadOnlyField {
	field := &ReadOnlyField{
		BaseField:   NewBaseField(fieldName),
		SourceField: sourceField,
	}
	field.ReadOnly = true
	return field
}

// ToInternalValue always returns nil for read-only fields
func (f *ReadOnlyField) ToInternalValue(data interface{}) (interface{}, error) {
	// Read-only fields are not set from input
	return nil, nil
}

// ToRepresentation uses source field for representation
func (f *ReadOnlyField) ToRepresentation(value interface{}) (interface{}, error) {
	if f.SourceField != nil {
		return f.SourceField.ToRepresentation(value)
	}
	return value, nil
}
