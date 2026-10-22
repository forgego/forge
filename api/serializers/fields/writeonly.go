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

func (f *WriteOnlyField) IsRequired() bool {
	return f.BaseField.IsRequired() || (f.SourceField != nil && f.SourceField.IsRequired())
}

func (f *WriteOnlyField) DefaultValue() interface{} {
	if f.Default != nil {
		return f.Default
	}
	if source, ok := f.SourceField.(interface{ DefaultValue() interface{} }); ok {
		return source.DefaultValue()
	}
	return nil
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

// Validate applies the wrapped field's constraints to input values.
func (f *WriteOnlyField) Validate(value interface{}) error {
	if err := f.BaseField.Validate(value); err != nil {
		return err
	}
	if f.SourceField != nil {
		return f.SourceField.Validate(value)
	}
	return nil
}
