package serializers

import (
	"github.com/forgego/forge/schema"
	"testing"
)

type typedDocumentedModel struct {
	schema.BaseSchema
	Name string `json:"name" db:"name"`
}

func (typedDocumentedModel) Fields() []schema.Field {
	return []schema.Field{schema.StringField("name", schema.Required())}
}

// Schema-backed typed serializers populate a concrete model.
func TestTypedSerializerDeserializesModel(t *testing.T) {
	serializer := FromSchema[typedDocumentedModel](typedDocumentedModel{})
	var model typedDocumentedModel
	if err := serializer.Deserialize(map[string]any{"name": "Ada"}, &model); err != nil {
		t.Fatal(err)
	}
	if model.Name != "Ada" {
		t.Fatalf("name=%q", model.Name)
	}
}
