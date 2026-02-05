package schema

// RelationType represents the type of a relationship
type RelationType int

const (
	RelationForeignKey RelationType = iota
	RelationOneToOne
	RelationManyToMany
)

func (t RelationType) String() string {
	switch t {
	case RelationForeignKey:
		return "ForeignKey"
	case RelationOneToOne:
		return "OneToOne"
	case RelationManyToMany:
		return "ManyToMany"
	default:
		return "Unknown"
	}
}

// CascadeType represents cascade behavior
type CascadeType string

const (
	CascadeCASCADE     CascadeType = "CASCADE"
	CascadeSET_NULL    CascadeType = "SET_NULL"
	CascadePROTECT     CascadeType = "PROTECT"
	CascadeSET_DEFAULT CascadeType = "SET_DEFAULT"
	CascadeDO_NOTHING  CascadeType = "DO_NOTHING"
)

// FKMatchType represents foreign key match type
type FKMatchType string

const (
	FKMatchFull    FKMatchType = "FULL"
	FKMatchPartial FKMatchType = "PARTIAL"
	FKMatchSimple  FKMatchType = "SIMPLE"
)

// Relation represents a model relationship with full Django relationship features
type Relation struct {
	LimitChoicesTo   interface{}
	CustomRelation   CustomRelation
	Name             string
	To               string
	RelatedName      string
	RelatedQueryName string
	OnDelete         CascadeType
	OnUpdate         CascadeType
	Through          string
	Type             RelationType
	Symmetrical      bool
	ParentLink       bool
	DBConstraint     bool           // Control FK constraint creation (default: true)
	ConstraintName   string         // Explicit FK constraint name
	Deferrable       DeferrableType // FK constraint deferrability
	Match            FKMatchType    // FK match type (FULL, PARTIAL, SIMPLE)
	Swappable        string         // Swappable model support (Django-specific)
}

// CustomRelation is the interface for custom relation types
type CustomRelation interface {
	GetName() string
	GetType() RelationType
	GetTarget() string
}
