package renderers

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONRenderer_Render(t *testing.T) {
	renderer := NewJSONRenderer()

	data := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	result, err := renderer.Render(data)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(result, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "John", decoded["name"])
	assert.Equal(t, "john@example.com", decoded["email"])
}

func TestJSONRenderer_MediaType(t *testing.T) {
	renderer := NewJSONRenderer()
	assert.Equal(t, "application/json", renderer.MediaType())
}

func TestJSONRenderer_RenderToWriter(t *testing.T) {
	renderer := NewJSONRenderer()
	var buf bytes.Buffer

	data := map[string]interface{}{
		"name": "John",
	}

	err := renderer.RenderToWriter(&buf, data)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, "John", decoded["name"])
}

func TestXMLRenderer_Render(t *testing.T) {
	renderer := NewXMLRenderer()

	// XML renderer requires a struct or a type that implements xml.Marshaler
	// Maps don't work directly with xml.Marshal, so we use a struct
	type Person struct {
		XMLName xml.Name `xml:"person"`
		Name    string   `xml:"name"`
		Email   string   `xml:"email"`
	}

	data := Person{
		Name:  "John",
		Email: "john@example.com",
	}

	result, err := renderer.Render(data)
	require.NoError(t, err)

	// Check it's valid XML (starts with <)
	assert.True(t, len(result) > 0)
	assert.Contains(t, string(result), "<person>")
}

func TestXMLRenderer_MediaType(t *testing.T) {
	renderer := NewXMLRenderer()
	assert.Equal(t, "application/xml", renderer.MediaType())
}

func TestHTMLRenderer_MediaType(t *testing.T) {
	renderer := NewHTMLRenderer()
	assert.Equal(t, "text/html", renderer.MediaType())
}

func TestHTMLRenderer_Render(t *testing.T) {
	renderer := NewHTMLRenderer()

	data := map[string]interface{}{
		"name": "John",
	}

	result, err := renderer.Render(data)
	require.NoError(t, err)

	// Should contain HTML
	assert.Contains(t, string(result), "<html")
	assert.Contains(t, string(result), "<body")
}

func TestRendererList_GetRenderer(t *testing.T) {
	list := RendererList{
		NewJSONRenderer(),
		NewXMLRenderer(),
		NewHTMLRenderer(),
	}

	// Get JSON renderer
	renderer := list.GetRenderer("application/json")
	assert.NotNil(t, renderer)
	assert.Equal(t, "application/json", renderer.MediaType())

	// Get non-existent renderer
	renderer = list.GetRenderer("application/unknown")
	assert.Nil(t, renderer)
}

func TestRendererList_GetMediaTypes(t *testing.T) {
	list := RendererList{
		NewJSONRenderer(),
		NewXMLRenderer(),
		NewHTMLRenderer(),
	}

	types := list.GetMediaTypes()
	assert.Contains(t, types, "application/json")
	assert.Contains(t, types, "application/xml")
	assert.Contains(t, types, "text/html")
}
