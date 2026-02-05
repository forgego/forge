package fields

// WriteOnlyField represents a write-only field (e.g., password)
type WriteOnlyField struct {
	*BaseField
	SourceField Field
}

// NewWriteOnlyField creates a new write-only field
func NewWriteOnlyField(fieldName string, sourceField Field) *WriteOnlyField {
	field := &WriteOnlyField{
		BaseField:   NewBaseField(fieldName),
		SourceField: sourceField,
	}
	field.WriteOnly = true
	return field
}

// ToRepresentation always returns nil for write-only fields
func (f *WriteOnlyField) ToRepresentation(value interface{}) (interface{}, error) {
	// Write-only fields are not included in output
	return nil, nil
}

// ToInternalValue uses source field for internal value
func (f *WriteOnlyField) ToInternalValue(data interface{}) (interface{}, error) {
	if f.SourceField != nil {
		return f.SourceField.ToInternalValue(data)
	}
	return data, nil
}
