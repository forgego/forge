package schema

// IndexOrder represents the ordering direction for an index column
type IndexOrder string

const (
	IndexOrderASC  IndexOrder = "ASC"
	IndexOrderDESC IndexOrder = "DESC"
)

// IndexField represents a field in an index with optional ordering
type IndexField struct {
	Name  string
	Order IndexOrder // ASC or DESC
}

// IndexType represents the type of index
type IndexType string

const (
	IndexTypeBTree IndexType = "btree" // Default B-tree index
	IndexTypeHash  IndexType = "hash"  // Hash index
	IndexTypeGIN   IndexType = "gin"   // GIN index (PostgreSQL)
	IndexTypeGiST  IndexType = "gist"  // GiST index (PostgreSQL)
	IndexTypeBRIN  IndexType = "brin"  // BRIN index (PostgreSQL)
)

// Index represents a database index
type Index struct {
	Name        string
	Fields      []string     // Simple field names
	FieldSpecs  []IndexField // Field specifications with ordering
	Expressions []string     // Functional index expressions (e.g., "LOWER(title)")
	Unique      bool
	Type        IndexType         // Index type (btree, hash, gin, gist, brin)
	Condition   string            // WHERE condition for partial indexes
	Include     []string          // INCLUDE columns for covering indexes (PostgreSQL)
	OpClasses   []string          // Operator classes for PostgreSQL
	Tablespace  string            // Tablespace for index (PostgreSQL)
	Options     map[string]string // Index options (e.g., fillfactor)
}

// ConstraintType represents the type of constraint
type ConstraintType string

const (
	ConstraintTypeCheck      ConstraintType = "CHECK"
	ConstraintTypeUnique     ConstraintType = "UNIQUE"
	ConstraintTypePrimaryKey ConstraintType = "PRIMARY KEY"
	ConstraintTypeForeignKey ConstraintType = "FOREIGN KEY"
	ConstraintTypeExclusion  ConstraintType = "EXCLUDE" // PostgreSQL exclusion constraint
)

// DeferrableType represents constraint deferrability
type DeferrableType string

const (
	DeferrableNotDeferrable      DeferrableType = "NOT DEFERRABLE"
	DeferrableDeferrable         DeferrableType = "DEFERRABLE"
	DeferrableInitiallyDeferred  DeferrableType = "DEFERRABLE INITIALLY DEFERRED"
	DeferrableInitiallyImmediate DeferrableType = "DEFERRABLE INITIALLY IMMEDIATE"
)

// Constraint represents a database constraint
type Constraint struct {
	Name         string
	Type         ConstraintType
	Condition    string         // SQL expression for CHECK constraints
	Fields       []string       // Fields involved in constraint
	Deferrable   DeferrableType // Constraint deferrability
	NotValid     bool           // NOT VALID option (PostgreSQL) - constraint not validated immediately
	ExcludeUsing string         // Index method for EXCLUDE constraints (e.g., "gist", "btree")
	ExcludeWith  []string       // Operators for EXCLUDE constraints (PostgreSQL)
}

// Permission represents a custom permission
type Permission struct {
	Name        string
	Description string
}

// Meta contains model metadata with ALL Django Meta options
type Meta struct {
	TableName          string
	DefaultRelatedName string
	VerboseName        string
	VerboseNamePlural  string
	AppLabel           string
	DefaultManager     string
	BaseManager        string // Custom base manager name
	GetLatestBy        string // Field to use for latest() queries
	DBTablespace       string // Default tablespace for model (PostgreSQL)
	Constraints        []Constraint
	UniqueTogether     [][]string
	Indexes            []Index
	Permissions        []Permission
	OrderBy            []string
	Proxy              bool
	Abstract           bool
	Managed            bool
	SelectOnSave       bool
	Swappable          string            // Swappable model support (Django-specific)
	RequiredDBFeatures []string          // Database feature requirements
	RequiredDBVendor   string            // Database vendor requirements
	TableComment       string            // Table comment/description
	TableOptions       map[string]string // Table-specific options (e.g., ENGINE for MySQL)
}

// DefaultMeta returns a Meta with sensible defaults
func DefaultMeta() Meta {
	return Meta{
		Managed: true,
	}
}
