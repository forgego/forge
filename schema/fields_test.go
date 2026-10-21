package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Schema supports listed scalar field types
func TestSchemaSupportsListedScalarFieldTypes(t *testing.T) {
	testCases := []struct {
		field    Field
		name     string
		expected FieldType
	}{
		{Int64Field("id"), "id", TypeInt64},
		{Int32Field("count32"), "count32", TypeInt32},
		{StringField("title"), "title", TypeString},
		{TextField("body"), "body", TypeText},
		{BoolField("is_active"), "is_active", TypeBool},
		{DateField("published_date"), "published_date", TypeDate},
		{TimeField("delivery_time"), "delivery_time", TypeTime},
		{DateTimeField("created_at"), "created_at", TypeDateTime},
		{Float64Field("score"), "score", TypeFloat64},
		{Float32Field("scale32"), "scale32", TypeFloat32},
		{DecimalField("amount"), "amount", TypeDecimal},
		{EmailField("user_email"), "user_email", TypeEmail},
		{URLField("profile_url"), "profile_url", TypeURL},
		{UUIDField("identifier"), "identifier", TypeUUID},
		{JSONField("payload_json"), "payload_json", TypeJSON},
		{BytesField("raw_bytes"), "raw_bytes", TypeBytes},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.name, tc.field.Name)
		assert.Equal(t, tc.expected, tc.field.Type, "field %s should have type %v", tc.name, tc.expected)
	}
}

// Schema supports generated SQL columns
func TestSchemaSupportsGeneratedSQLColumns(t *testing.T) {
	storedField := StringField("full_name", GeneratedColumn("first_name || ' ' || last_name", true))
	assert.True(t, storedField.Generated)
	assert.Equal(t, "first_name || ' ' || last_name", storedField.GeneratedExpr)
	assert.True(t, storedField.IsStored)

	virtualField := Int64Field("doubled_val", GeneratedColumn("val * 2", false))
	assert.True(t, virtualField.Generated)
	assert.Equal(t, "val * 2", virtualField.GeneratedExpr)
	assert.False(t, virtualField.IsStored)
}

// Schema supports database column options
func TestSchemaSupportsDatabaseColumnOptions(t *testing.T) {
	f := StringField("slug",
		DBColumn("post_slug"),
		DBDefault("'default_slug'"),
		DBCollation("utf8mb4_unicode_ci"),
		DBComment("Slug for URLs"),
		DBTablespace("index_ts"),
		DBIndex(),
	)

	assert.Equal(t, "post_slug", f.DBColumn)
	assert.Equal(t, "'default_slug'", f.DBDefault)
	assert.Equal(t, "utf8mb4_unicode_ci", f.DBCollation)
	assert.Equal(t, "Slug for URLs", f.DBComment)
	assert.Equal(t, "index_ts", f.DBTablespace)
	assert.True(t, f.DBIndex)

	opts := NewFieldOptions()
	require.NotNil(t, opts)
	opts.DB.Column = "configured_col"
	opts.DB.Default = "NOW()"
	opts.DB.Collation = "en_US"
	opts.DB.Comment = "Column comment"
	opts.DB.Tablespace = "pg_default"
	opts.DB.Index = true

	var target Field
	opts.ApplyToField(&target)
	assert.Equal(t, "configured_col", target.DBColumn)
	assert.Equal(t, "NOW()", target.DBDefault)
	assert.Equal(t, "en_US", target.DBCollation)
	assert.Equal(t, "Column comment", target.DBComment)
	assert.Equal(t, "pg_default", target.DBTablespace)
	assert.True(t, target.DBIndex)
}

// Schema supports value-label choices
func TestSchemaSupportsValueLabelChoices(t *testing.T) {
	choice := NewChoice("active", "Active Status")
	assert.Equal(t, "active", choice.Value)
	assert.Equal(t, "Active Status", choice.Label)

	choices := Choices("p", "Pending", "c", "Completed", "f", "Failed")
	require.Len(t, choices, 3)
	assert.Equal(t, "p", choices[0].Value)
	assert.Equal(t, "Pending", choices[0].Label)
	assert.Equal(t, "c", choices[1].Value)
	assert.Equal(t, "Completed", choices[1].Label)
	assert.Equal(t, "f", choices[2].Value)
	assert.Equal(t, "Failed", choices[2].Label)

	oddChoices := Choices("singleton")
	require.Len(t, oddChoices, 1)
	assert.Equal(t, "singleton", oddChoices[0].Value)
	assert.Equal(t, "singleton", oddChoices[0].Label)
}
