package fields

// HiddenField represents a hidden field (never shown in output)
type HiddenField struct {
	*BaseField
	SourceField Field
}

// NewHiddenField creates a new hidden field
func NewHiddenField(fieldName string, sourceField Field) *HiddenField {
	field := &HiddenField{
		BaseField:   NewBaseField(fieldName),
		SourceField: sourceField,
	}
	field.ReadOnly = true
	field.WriteOnly = true
	return field
}

// ToRepresentation always returns nil for hidden fields
func (f *HiddenField) ToRepresentation(value interface{}) (interface{}, error) {
	return nil, nil
}

// ToInternalValue uses source field for internal value
func (f *HiddenField) ToInternalValue(data interface{}) (interface{}, error) {
	if f.SourceField != nil {
		return f.SourceField.ToInternalValue(data)
	}
	return data, nil
}
