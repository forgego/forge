package parsers

import (
	"encoding/xml"
	"io"
)

// XMLParser parses XML data
type XMLParser struct{}

// NewXMLParser creates a new XML parser
func NewXMLParser() *XMLParser {
	return &XMLParser{}
}

// Parse parses XML data from a reader
func (p *XMLParser) Parse(r io.Reader, v interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

// MediaType returns the XML media type
func (p *XMLParser) MediaType() string {
	return "application/xml"
}
