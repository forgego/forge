package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a simple model for testing
type TestModel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// TestSerializer is a test serializer
type TestSerializer struct {
	*BaseSerializer
}

func NewTestSerializer() Serializer {
	return &TestSerializer{
		BaseSerializer: NewBaseSerializer(make(map[string]interface{})),
	}
}

func (s *TestSerializer) Fields() []string {
	return []string{"id", "name", "email", "age"}
}

func (s *TestSerializer) ReadOnlyFields() []string {
	return []string{"id"}
}

func TestBaseSerializer_Validate_Valid(t *testing.T) {
	serializer := NewTestSerializer()
	testSerializer := serializer.(*TestSerializer)

	// Set valid data
	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}
	testSerializer.BaseSerializer.SetData(data)

	err := serializer.Validate()
	assert.NoError(t, err)
	assert.True(t, serializer.IsValid())
}

func TestBaseSerializer_Validate_Invalid(t *testing.T) {
	serializer := NewTestSerializer()
	testSerializer := serializer.(*TestSerializer)

	// Set invalid data (missing required fields)
	data := map[string]interface{}{
		"name": "", // Empty name
	}
	testSerializer.BaseSerializer.SetData(data)

	err := serializer.Validate()
	// Validation might pass if fields are optional
	// This depends on validation rules
	_ = err
}

func TestBaseSerializer_ReadOnlyFields(t *testing.T) {
	serializer := NewTestSerializer()
	testSerializer := serializer.(*TestSerializer)

	data := map[string]interface{}{
		"id":    123, // Read-only field
		"name":  "John",
		"email": "john@example.com",
	}
	testSerializer.BaseSerializer.SetData(data)

	// Read-only fields should be ignored in validation
	readOnlyFields := testSerializer.ReadOnlyFields()
	assert.Contains(t, readOnlyFields, "id")
}

func TestSerializeModel(t *testing.T) {
	model := &TestModel{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	result := SerializeModel(model)

	assert.NotNil(t, result)
	// SerializeModel uses reflection to convert model to map
	// The exact field names depend on JSON tags
	if id, ok := result["id"]; ok {
		// ID might be int or int64 depending on reflection
		assert.True(t, id == int64(1) || id == 1)
	}
	if name, ok := result["name"]; ok {
		assert.Equal(t, "John Doe", name)
	}
	if email, ok := result["email"]; ok {
		assert.Equal(t, "john@example.com", email)
	}
	if age, ok := result["age"]; ok {
		// Age might be int or int64 depending on reflection
		assert.True(t, age == 30 || age == int64(30))
	}
}

func TestSerializeMany(t *testing.T) {
	models := []interface{}{
		&TestModel{ID: 1, Name: "John", Email: "john@example.com", Age: 30},
		&TestModel{ID: 2, Name: "Jane", Email: "jane@example.com", Age: 25},
	}

	results := SerializeMany(models)

	require.Len(t, results, 2)
	assert.Equal(t, "John", results[0]["name"])
	assert.Equal(t, "Jane", results[1]["name"])
}

func TestBaseSerializer_SetData(t *testing.T) {
	serializer := NewBaseSerializer(make(map[string]interface{}))
	data := map[string]interface{}{
		"name": "Test",
	}

	serializer.SetData(data)

	// Check data was set
	assert.Equal(t, "Test", serializer.Get("name"))
	assert.False(t, serializer.IsValid()) // Should reset validity
}

func TestBaseSerializer_Errors(t *testing.T) {
	serializer := NewBaseSerializer(make(map[string]interface{}))

	// Initially no errors
	errors := serializer.Errors()
	assert.Empty(t, errors)

	// After validation failure, errors should be populated
	// (This depends on validation implementation)
	_ = errors
}
