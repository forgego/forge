package parsers

import (
	"io"
)

// Parser is the interface for request parsers
type Parser interface {
	// Parse parses data from a reader
	Parse(r io.Reader, v interface{}) error
	// MediaType returns the media type this parser handles
	MediaType() string
}

// ParserList is a list of parsers
type ParserList []Parser

// GetParser returns a parser by media type
func (pl ParserList) GetParser(mediaType string) Parser {
	for _, p := range pl {
		if p.MediaType() == mediaType {
			return p
		}
	}
	return nil
}

// GetMediaTypes returns all media types supported by this list
func (pl ParserList) GetMediaTypes() []string {
	types := make([]string, len(pl))
	for i, p := range pl {
		types[i] = p.MediaType()
	}
	return types
}
