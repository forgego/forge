package schema

// Schema is the core interface that all models must implement.
// It defines the structure and behavior of a model.
type Schema interface {
	// Fields returns all field definitions for the model
	Fields() []Field

	// Relations returns all relationship definitions (Django-style, not Ent edges)
	Relations() []Relation

	// Meta returns model metadata (full Django Meta options)
	Meta() Meta

	// Hooks returns model lifecycle hooks
	Hooks() *ModelHooks
}

// BaseSchema provides a default implementation that can be embedded
type BaseSchema struct{}

func (b BaseSchema) Fields() []Field {
	return []Field{}
}

func (b BaseSchema) Relations() []Relation {
	return []Relation{}
}

func (b BaseSchema) Meta() Meta {
	return Meta{}
}

func (b BaseSchema) Hooks() *ModelHooks {
	return nil
}
