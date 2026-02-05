package schema

// Common field building helpers and convenience functions

// NewChoice creates a new Choice with value and label
func NewChoice(value, label string) Choice {
	return Choice{
		Value: value,
		Label: label,
	}
}

// Choices creates multiple choices from value-label pairs
// Usage: Choices("active", "Active", "inactive", "Inactive")
func Choices(pairs ...string) []Choice {
	if len(pairs)%2 != 0 {
		// If odd number, treat as values only (label = value)
		choices := make([]Choice, len(pairs))
		for i, val := range pairs {
			choices[i] = Choice{Value: val, Label: val}
		}
		return choices
	}

	choices := make([]Choice, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		choices[i/2] = Choice{
			Value: pairs[i],
			Label: pairs[i+1],
		}
	}
	return choices
}

// Common index creation helpers

// IndexOn creates a simple index on the given fields
func IndexOn(name string, fields ...string) Index {
	return Index{
		Name:   name,
		Fields: fields,
		Type:   IndexTypeBTree,
	}
}

// UniqueIndexOn creates a unique index on the given fields
func UniqueIndexOn(name string, fields ...string) Index {
	return Index{
		Name:   name,
		Fields: fields,
		Unique: true,
		Type:   IndexTypeBTree,
	}
}

// GINIndex creates a GIN index (PostgreSQL) - useful for JSONB, arrays, full-text search
func GINIndex(name string, fields ...string) Index {
	return Index{
		Name:   name,
		Fields: fields,
		Type:   IndexTypeGIN,
	}
}

// GiSTIndex creates a GiST index (PostgreSQL) - useful for geometric/spatial data
func GiSTIndex(name string, fields ...string) Index {
	return Index{
		Name:   name,
		Fields: fields,
		Type:   IndexTypeGiST,
	}
}

// Common constraint creation helpers

// Check creates a CHECK constraint
func Check(name string, condition string) Constraint {
	return Constraint{
		Name:      name,
		Type:      ConstraintTypeCheck,
		Condition: condition,
	}
}

// UniqueOn creates a UNIQUE constraint on the given fields
func UniqueOn(name string, fields ...string) Constraint {
	return Constraint{
		Name:   name,
		Type:   ConstraintTypeUnique,
		Fields: fields,
	}
}

// Common Meta helpers

// WithTableName sets the table name in Meta
func WithTableName(tableName string) func(*Meta) {
	return func(m *Meta) {
		m.TableName = tableName
	}
}

// WithOrderBy sets the default ordering in Meta
func WithOrderBy(orderBy ...string) func(*Meta) {
	return func(m *Meta) {
		m.OrderBy = orderBy
	}
}

// WithIndexes adds indexes to Meta
func WithIndexes(indexes ...Index) func(*Meta) {
	return func(m *Meta) {
		m.Indexes = append(m.Indexes, indexes...)
	}
}

// WithConstraints adds constraints to Meta
func WithConstraints(constraints ...Constraint) func(*Meta) {
	return func(m *Meta) {
		m.Constraints = append(m.Constraints, constraints...)
	}
}

// WithUniqueTogether sets unique together constraints
func WithUniqueTogether(fieldGroups ...[]string) func(*Meta) {
	return func(m *Meta) {
		m.UniqueTogether = fieldGroups
	}
}

// BuildMeta builds a Meta struct using functional options
func BuildMeta(options ...func(*Meta)) Meta {
	meta := DefaultMeta()
	for _, option := range options {
		option(&meta)
	}
	return meta
}
