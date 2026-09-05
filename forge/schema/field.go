package schema

// FieldType represents the type of a field
type FieldType int

const (
	TypeInt64 FieldType = iota
	TypeInt32
	TypeString
	TypeBool
	TypeTime
	TypeDate
	TypeDateTime
	TypeFloat32
	TypeFloat64
	TypeDecimal
	TypeText
	TypeEmail
	TypeURL
	TypeUUID
	TypeJSON
	TypeBytes
	TypeForeignKey
	TypeManyToMany
	TypeOneToOne
)

func (t FieldType) String() string {
	switch t {
	case TypeInt64:
		return "Int64"
	case TypeInt32:
		return "Int32"
	case TypeString:
		return "String"
	case TypeBool:
		return "Bool"
	case TypeTime:
		return "Time"
	case TypeDate:
		return "Date"
	case TypeDateTime:
		return "DateTime"
	case TypeFloat32:
		return "Float32"
	case TypeFloat64:
		return "Float64"
	case TypeDecimal:
		return "Decimal"
	case TypeText:
		return "Text"
	case TypeEmail:
		return "Email"
	case TypeURL:
		return "URL"
	case TypeUUID:
		return "UUID"
	case TypeJSON:
		return "JSON"
	case TypeBytes:
		return "Bytes"
	case TypeForeignKey:
		return "ForeignKey"
	case TypeManyToMany:
		return "ManyToMany"
	case TypeOneToOne:
		return "OneToOne"
	default:
		return "Unknown"
	}
}

// Field represents a model field with ALL Django field options.
// This struct maintains compatibility while the internal
// organization is improved through FieldOptions (see field_config.go).
type Field struct {
	// Core field properties
	CustomField CustomField
	Name        string
	Type        FieldType
	Default     interface{}
	PrimaryKey  bool

	// Database configuration (organized in FieldOptions.DBConfig)
	DBDefault     string // Database-level default (SQL expression)
	DBColumn      string // Custom database column name
	DBType        string // Explicit database column type override
	DBCollation   string // Character set collation (PostgreSQL, MySQL)
	DBComment     string // Column comment/description
	DBTablespace  string // Tablespace for field storage (PostgreSQL)
	DBIndex       bool   // Create database index on this field
	Generated     bool   // Generated column (STORED/VIRTUAL)
	GeneratedExpr string // Expression for generated column
	IsStored      bool   // For generated columns: STORED vs VIRTUAL

	// Validation configuration (organized in FieldOptions.ValidationConfig)
	Required       bool              // NOT NULL constraint
	Unique         bool              // UNIQUE constraint
	Blank          bool              // Allow blank/empty values
	MinLength      *int              // Minimum length (for strings/arrays)
	MaxLength      *int              // Maximum length (for strings/arrays)
	MinValue       *float64          // Minimum value (for numeric types)
	MaxValue       *float64          // Maximum value (for numeric types)
	MaxDigits      *int              // Maximum digits (for Decimal fields)
	DecimalPlaces  *int              // Decimal places (for Decimal fields)
	Choices        []Choice          // Predefined choices
	Validators     []Validator       // Custom validators
	ValidationTag  string            // Validation tag (e.g., for struct tags)
	ErrorMessages  map[string]string // Custom error messages per validation type
	UniqueForDate  string            // Unique constraint scoped to date field
	UniqueForMonth string            // Unique constraint scoped to month
	UniqueForYear  string            // Unique constraint scoped to year

	// Presentation configuration (organized in FieldOptions.PresentationConfig)
	VerboseName string // Human-readable name
	HelpText    string // Help text for forms/admin
	Editable    bool   // Control field editability in forms/admin
	Serialize   bool   // Control field serialization in forms/API

	// Type-specific behaviors
	AutoIncrement bool // Auto-increment (for integer fields)
	AutoNow       bool // Auto-update on save (for temporal fields)
	AutoNowAdd    bool // Auto-set on creation (for temporal fields)
}

// Choice represents a choice option for ChoiceField
type Choice struct {
	Value string
	Label string
}

// Validator is the interface for field validators
type Validator interface {
	Validate(value interface{}) error
}

// CustomField is the interface for custom field types
type CustomField interface {
	GetName() string
	GetType() FieldType
	Validate(value interface{}) error
}
