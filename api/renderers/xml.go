package renderers

import (
	"encoding/xml"
	"io"
)

// XMLRenderer renders data as XML
type XMLRenderer struct {
	// Indent is the indentation string (default: "  ")
	Indent string
}

// NewXMLRenderer creates a new XML renderer
func NewXMLRenderer() *XMLRenderer {
	return &XMLRenderer{
		Indent: "  ",
	}
}

// Render renders data to XML bytes
func (r *XMLRenderer) Render(data interface{}) ([]byte, error) {
	return xml.MarshalIndent(data, "", r.Indent)
}

// MediaType returns the XML media type
func (r *XMLRenderer) MediaType() string {
	return "application/xml"
}

// RenderToWriter renders data directly to a writer
func (r *XMLRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", r.Indent)
	return encoder.Encode(data)
}
