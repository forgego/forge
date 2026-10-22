package validation

import (
	"errors"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

type schemaConsumerRecord struct {
	schema.BaseSchema
	Display string `json:"display_name" db:"display_column"`
	Count   int64  `json:"count" db:"count"`
	Enabled bool   `json:"enabled" db:"enabled"`
	Email   string `json:"email" db:"email"`
	Status  string `json:"status" db:"status"`
}

func (schemaConsumerRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("display_name", schema.Required(), schema.MaxLength(8)),
		schema.Int64Field("count", schema.Required(), schema.MinValue(0), schema.MaxValue(10)),
		schema.BoolField("enabled", schema.Required()),
		schema.EmailField("email", schema.Required()),
		schema.StringField("status", schema.Required(), schema.ChoicesOpts(schema.NewChoice("in progress", "In progress"), schema.NewChoice("done", "Done"))),
	}
}
func TestFromSchemaValidatesConcreteFieldsAndAliases(t *testing.T) {
	validator := FromSchema[schemaConsumerRecord](schemaConsumerRecord{})
	valid := schemaConsumerRecord{Display: "Valid", Count: 0, Enabled: false, Email: "user@example.com", Status: "in progress"}
	assert.NoError(t, validator.Validate(&valid))
	for _, test := range []struct {
		name string
		edit func(*schemaConsumerRecord)
	}{
		{"required", func(v *schemaConsumerRecord) { v.Display = "" }},
		{"alias length", func(v *schemaConsumerRecord) { v.Display = "too-long-name" }},
		{"minimum", func(v *schemaConsumerRecord) { v.Count = -1 }},
		{"maximum", func(v *schemaConsumerRecord) { v.Count = 11 }},
		{"email", func(v *schemaConsumerRecord) { v.Email = "not-an-email" }},
		{"choices", func(v *schemaConsumerRecord) { v.Status = "invalid" }},
	} {
		t.Run(test.name, func(t *testing.T) { value := valid; test.edit(&value); assert.Error(t, validator.Validate(&value)) })
	}
}

type schemaRejectValue struct{}

func (schemaRejectValue) Validate(any) error { return errors.New("custom validation rejected") }

type schemaCustomRecord struct {
	schema.BaseSchema
	Name string
}

func (schemaCustomRecord) Fields() []schema.Field {
	field := schema.StringField("name")
	field.Validators = []schema.Validator{schemaRejectValue{}}
	return []schema.Field{field}
}
func TestFromSchemaRunsCustomValidators(t *testing.T) {
	assert.ErrorContains(t, FromSchema[schemaCustomRecord](schemaCustomRecord{}).Validate(&schemaCustomRecord{Name: "value"}), "custom validation rejected")
}
func TestFromSchemaRejectsNilInputWithoutPanic(t *testing.T) {
	assert.Error(t, FromSchema[schemaConsumerRecord](schemaConsumerRecord{}).Validate(nil))
}

type schemaPointerRecord struct {
	schema.BaseSchema
	Name  *string
	Count *int64
}

func (schemaPointerRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("name", schema.Required(), schema.MinLength(2)),
		schema.Int64Field("count", schema.Optional(), schema.MinValue(2)),
	}
}
func TestFromSchemaValidatesPointerContents(t *testing.T) {
	validator := FromSchema[schemaPointerRecord](schemaPointerRecord{})
	name := "ok"
	count := int64(1)
	assert.NoError(t, validator.Validate(&schemaPointerRecord{Name: &name}))
	assert.Error(t, validator.Validate(&schemaPointerRecord{Name: &name, Count: &count}))
	name = ""
	assert.Error(t, validator.Validate(&schemaPointerRecord{Name: &name}))
	assert.Error(t, validator.Validate(&schemaPointerRecord{}))
}

type optionalFormatRecord struct {
	schema.BaseSchema
	Email       *string
	URL         *string
	UUID        *string
	StrictEmail *string
}

func (optionalFormatRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.EmailField("email", schema.Optional(), schema.Blank()),
		schema.URLField("url", schema.Optional(), schema.Blank()),
		schema.UUIDField("uuid", schema.Optional(), schema.Blank()),
		schema.EmailField("strict_email", schema.Optional()),
	}
}
func TestFromSchemaRespectsExplicitBlankFormats(t *testing.T) {
	blank := ""
	validator := FromSchema[optionalFormatRecord](optionalFormatRecord{})
	assert.NoError(t, validator.Validate(&optionalFormatRecord{Email: &blank, URL: &blank, UUID: &blank}))
	assert.Error(t, validator.Validate(&optionalFormatRecord{StrictEmail: &blank}))
}
