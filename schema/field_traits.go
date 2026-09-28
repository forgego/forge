package schema

// FieldTrait represents a capability that a field type can have
// Traits are used for documentation and runtime validation
type FieldTrait string

const (
	TraitNumeric  FieldTrait = "numeric"  // Supports MinValue, MaxValue, AutoIncrement
	TraitString   FieldTrait = "string"   // Supports MinLength, MaxLength, Choices
	TraitTemporal FieldTrait = "temporal" // Supports AutoNow, AutoNowAdd
	TraitDecimal  FieldTrait = "decimal"  // Supports MaxDigits, DecimalPlaces
)

// HasTrait checks if a field type has a specific trait
func HasTrait(fieldType FieldType, trait FieldTrait) bool {
	info := GetFieldTypeInfo(fieldType)
	for _, t := range info.Traits {
		if t == trait {
			return true
		}
	}
	return false
}

// SupportsNumericOperations checks if a field type supports numeric operations
func SupportsNumericOperations(fieldType FieldType) bool {
	return HasTrait(fieldType, TraitNumeric)
}

// SupportsStringOperations checks if a field type supports string operations
func SupportsStringOperations(fieldType FieldType) bool {
	return HasTrait(fieldType, TraitString)
}

// SupportsTemporalOperations checks if a field type supports temporal operations
func SupportsTemporalOperations(fieldType FieldType) bool {
	return HasTrait(fieldType, TraitTemporal)
}

// SupportsDecimalOperations checks if a field type supports decimal operations
func SupportsDecimalOperations(fieldType FieldType) bool {
	return HasTrait(fieldType, TraitDecimal)
}

// FieldTypeInfo contains metadata about a field type
type FieldTypeInfo struct {
	Type         FieldType
	Traits       []FieldTrait
	DefaultValue interface{}
}

// GetFieldTypeInfo returns the FieldTypeInfo for a given FieldType
func GetFieldTypeInfo(fieldType FieldType) FieldTypeInfo {
	infos := map[FieldType]FieldTypeInfo{
		TypeInt64: {
			Type:   TypeInt64,
			Traits: []FieldTrait{TraitNumeric},
		},
		TypeInt32: {
			Type:   TypeInt32,
			Traits: []FieldTrait{TraitNumeric},
		},
		TypeString: {
			Type:   TypeString,
			Traits: []FieldTrait{TraitString},
		},
		TypeText: {
			Type:   TypeText,
			Traits: []FieldTrait{TraitString},
		},
		TypeEmail: {
			Type:   TypeEmail,
			Traits: []FieldTrait{TraitString},
		},
		TypeURL: {
			Type:   TypeURL,
			Traits: []FieldTrait{TraitString},
		},
		TypeBool: {
			Type:   TypeBool,
			Traits: []FieldTrait{},
		},
		TypeTime: {
			Type:   TypeTime,
			Traits: []FieldTrait{TraitTemporal},
		},
		TypeDate: {
			Type:   TypeDate,
			Traits: []FieldTrait{TraitTemporal},
		},
		TypeDateTime: {
			Type:   TypeDateTime,
			Traits: []FieldTrait{TraitTemporal},
		},
		TypeFloat32: {
			Type:   TypeFloat32,
			Traits: []FieldTrait{TraitNumeric},
		},
		TypeFloat64: {
			Type:   TypeFloat64,
			Traits: []FieldTrait{TraitNumeric},
		},
		TypeDecimal: {
			Type:   TypeDecimal,
			Traits: []FieldTrait{TraitNumeric, TraitDecimal},
		},
		TypeUUID: {
			Type:   TypeUUID,
			Traits: []FieldTrait{},
		},
		TypeJSON: {
			Type:   TypeJSON,
			Traits: []FieldTrait{},
		},
		TypeBytes: {
			Type:   TypeBytes,
			Traits: []FieldTrait{TraitString}, // Bytes can have length constraints
		},
	}

	if info, ok := infos[fieldType]; ok {
		return info
	}
	return FieldTypeInfo{Type: fieldType}
}
