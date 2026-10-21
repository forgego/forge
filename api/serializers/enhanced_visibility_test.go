package serializers

import (
	"github.com/forgego/forge/api/serializers/fields"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEnhancedSerializerHonorsWriteOnlyWithJSONOptions(t *testing.T) {
	value := struct {
		Name   string `json:"name,omitempty"`
		Secret string `json:"secret,omitempty"`
	}{Name: "public", Secret: "synthetic-private"}
	serializer := NewEnhancedSerializer()
	serializer.AddField("secret", fields.NewWriteOnlyField("secret", nil))
	data, err := serializer.ToRepresentation(&value)
	require.NoError(t, err)
	assert.Equal(t, "public", data["name"])
	assert.NotContains(t, data, "secret")
	assert.NotContains(t, data, "secret,omitempty")
}

type enhancedPrivateModel struct {
	schema.BaseSchema
	Name    string `json:"name"`
	Private string `json:"credential" db:"private_value"`
}

func (enhancedPrivateModel) Fields() []schema.Field {
	return []schema.Field{schema.StringField("name"), schema.StringField("private_value", schema.WriteOnly())}
}
func TestEnhancedSerializerHonorsSchemaVisibility(t *testing.T) {
	serializer := NewEnhancedSerializer()
	data, err := serializer.ToRepresentation(&enhancedPrivateModel{Name: "public", Private: "synthetic-private"})
	require.NoError(t, err)
	assert.Equal(t, "public", data["name"])
	assert.NotContains(t, data, "credential")
}

func TestEnhancedSerializerUsesDefaultWithNilData(t *testing.T) {
	s := NewEnhancedSerializer()
	field := fields.NewStringField("status")
	field.Default = "draft"
	s.AddField("status", field)
	require.NoError(t, s.Validate())
	require.Equal(t, "draft", s.Data["status"])
}

func TestEnhancedSerializerWrappedRequiredAndDefault(t *testing.T) {
	for _, wrapped := range []bool{false, true} {
		s := NewEnhancedSerializer()
		source := fields.NewStringField("secret")
		source.Required = true
		var field fields.Field = source
		if wrapped {
			field = fields.NewWriteOnlyField("secret", source)
		}
		s.AddField("secret", field)
		require.Error(t, s.Validate())
		require.Contains(t, s.Errors, "secret")

		source.Default = "fallback"
		require.NoError(t, s.Validate())
		require.Equal(t, "fallback", s.Data["secret"])
	}
}
