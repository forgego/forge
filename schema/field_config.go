package schema

// FieldConfig groups field configuration by concern to improve organization
// and reduce the monolithic Field struct

// DBConfig contains database-level configuration
type DBConfig struct {
	Column        string // Custom database column name
	Type          string // Explicit database column type override
	Collation     string // Character set collation (PostgreSQL, MySQL)
	Comment       string // Column comment/description
	Tablespace    string // Tablespace for field storage (PostgreSQL)
	Default       string // Database-level default (SQL expression)
	Index         bool   // Create database index on this field
	Generated     bool   // Generated column (STORED/VIRTUAL)
	GeneratedExpr string // Expression for generated column
	IsStored      bool   // For generated columns: STORED vs VIRTUAL
}

// ValidationConfig contains validation-related configuration
type ValidationConfig struct {
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
}

// PresentationConfig contains presentation/UI-related configuration
type PresentationConfig struct {
	VerboseName string // Human-readable name
	HelpText    string // Help text for forms/admin
	Editable    bool   // Control field editability in forms/admin
	Serialize   bool   // Control field serialization in forms/API
}

// FieldOptions contains all field configuration organized by concern
type FieldOptions struct {
	DB           DBConfig
	Validation   ValidationConfig
	Presentation PresentationConfig
}

// ApplyToField applies the options to a Field struct
func (o *FieldOptions) ApplyToField(f *Field) {
	// DB config
	f.DBColumn = o.DB.Column
	f.DBType = o.DB.Type
	f.DBCollation = o.DB.Collation
	f.DBComment = o.DB.Comment
	f.DBTablespace = o.DB.Tablespace
	f.DBDefault = o.DB.Default
	f.DBIndex = o.DB.Index
	f.Generated = o.DB.Generated
	f.GeneratedExpr = o.DB.GeneratedExpr
	f.IsStored = o.DB.IsStored

	// Validation config
	f.Required = o.Validation.Required
	f.Unique = o.Validation.Unique
	f.Blank = o.Validation.Blank
	f.MinLength = o.Validation.MinLength
	f.MaxLength = o.Validation.MaxLength
	f.MinValue = o.Validation.MinValue
	f.MaxValue = o.Validation.MaxValue
	f.MaxDigits = o.Validation.MaxDigits
	f.DecimalPlaces = o.Validation.DecimalPlaces
	f.Choices = o.Validation.Choices
	f.Validators = o.Validation.Validators
	f.ValidationTag = o.Validation.ValidationTag
	f.ErrorMessages = o.Validation.ErrorMessages
	f.UniqueForDate = o.Validation.UniqueForDate
	f.UniqueForMonth = o.Validation.UniqueForMonth
	f.UniqueForYear = o.Validation.UniqueForYear

	// Presentation config
	f.VerboseName = o.Presentation.VerboseName
	f.HelpText = o.Presentation.HelpText
	f.Editable = o.Presentation.Editable
	f.Serialize = o.Presentation.Serialize
}

// NewFieldOptions creates a new FieldOptions with sensible defaults
func NewFieldOptions() *FieldOptions {
	return &FieldOptions{
		DB: DBConfig{
			Index: false,
		},
		Validation: ValidationConfig{
			Required: false,
			Unique:   false,
			Blank:    false,
		},
		Presentation: PresentationConfig{
			Editable:  true,
			Serialize: true,
		},
	}
}
