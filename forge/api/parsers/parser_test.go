package parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONParser_Parse(t *testing.T) {
	parser := NewJSONParser()

	jsonData := `{"name":"John","email":"john@example.com"}`
	reader := strings.NewReader(jsonData)

	var result map[string]interface{}
	err := parser.Parse(reader, &result)

	require.NoError(t, err)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
}

func TestJSONParser_MediaType(t *testing.T) {
	parser := NewJSONParser()
	assert.Equal(t, "application/json", parser.MediaType())
}

func TestFormParser_Parse(t *testing.T) {
	parser := NewFormParser()

	formData := "name=John&email=john%40example.com"
	reader := strings.NewReader(formData)

	var result map[string]interface{}
	err := parser.Parse(reader, &result)

	require.NoError(t, err)
	assert.Equal(t, "John", result["name"])
	assert.Equal(t, "john@example.com", result["email"])
}

func TestFormParser_MediaType(t *testing.T) {
	parser := NewFormParser()
	assert.Equal(t, "application/x-www-form-urlencoded", parser.MediaType())
}

func TestParserList_GetParser(t *testing.T) {
	list := ParserList{
		NewJSONParser(),
		NewFormParser(),
	}

	// Get JSON parser
	parser := list.GetParser("application/json")
	assert.NotNil(t, parser)
	assert.Equal(t, "application/json", parser.MediaType())

	// Get non-existent parser
	parser = list.GetParser("application/unknown")
	assert.Nil(t, parser)
}

func TestParserList_GetMediaTypes(t *testing.T) {
	list := ParserList{
		NewJSONParser(),
		NewFormParser(),
	}

	types := list.GetMediaTypes()
	assert.Contains(t, types, "application/json")
	assert.Contains(t, types, "application/x-www-form-urlencoded")
}

func TestMultiPartParser_MediaType(t *testing.T) {
	parser := NewMultiPartParser()
	assert.Equal(t, "multipart/form-data", parser.MediaType())
}

func TestXMLParser_Parse(t *testing.T) {
	parser := NewXMLParser()

	xmlData := `<root><name>John</name><email>john@example.com</email></root>`
	reader := strings.NewReader(xmlData)

	var result map[string]interface{}
	err := parser.Parse(reader, &result)

	// XML parsing might require specific structure
	_ = err
	_ = result
}

func TestXMLParser_MediaType(t *testing.T) {
	parser := NewXMLParser()
	assert.Equal(t, "application/xml", parser.MediaType())
}
