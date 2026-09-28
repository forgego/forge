package fields

// SerializerMethodField represents a computed field using a method
type SerializerMethodField struct {
	*BaseField
	MethodName string
	MethodFunc func(interface{}) interface{}
}

// NewSerializerMethodField creates a new serializer method field
func NewSerializerMethodField(fieldName string, methodFunc func(interface{}) interface{}) *SerializerMethodField {
	field := &SerializerMethodField{
		BaseField:  NewBaseField(fieldName),
		MethodFunc: methodFunc,
	}
	field.ReadOnly = true
	return field
}

// ToRepresentation uses the method function to compute the value
func (f *SerializerMethodField) ToRepresentation(value interface{}) (interface{}, error) {
	if f.MethodFunc != nil {
		return f.MethodFunc(value), nil
	}
	return nil, nil
}

// ToInternalValue always returns nil (read-only)
func (f *SerializerMethodField) ToInternalValue(data interface{}) (interface{}, error) {
	return nil, nil
}
