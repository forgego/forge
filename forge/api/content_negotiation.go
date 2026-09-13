package api

import (
	"net/http"
	"strings"

	"github.com/forgego/forge/api/parsers"
	"github.com/forgego/forge/api/renderers"
)

// ContentNegotiator handles content negotiation
type ContentNegotiator struct {
	Renderers []renderers.Renderer
	Parsers   []parsers.Parser
}

// NewContentNegotiator creates a new content negotiator
func NewContentNegotiator(renderers []renderers.Renderer, parsers []parsers.Parser) *ContentNegotiator {
	return &ContentNegotiator{
		Renderers: renderers,
		Parsers:   parsers,
	}
}

// SelectRenderer selects a renderer based on Accept header
func (cn *ContentNegotiator) SelectRenderer(r *http.Request) renderers.Renderer {
	accept := r.Header.Get("Accept")
	if accept == "" {
		// Default to JSON
		return cn.Renderers[0]
	}

	// Parse Accept header
	mediaTypes := parseAcceptHeader(accept)

	// Find matching renderer
	for _, mt := range mediaTypes {
		for _, renderer := range cn.Renderers {
			if renderer.MediaType() == mt {
				return renderer
			}
		}
	}

	// Default to first renderer
	if len(cn.Renderers) > 0 {
		return cn.Renderers[0]
	}

	return nil
}

// SelectParser selects a parser based on Content-Type header
func (cn *ContentNegotiator) SelectParser(r *http.Request) parsers.Parser {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		// Default to JSON
		return cn.Parsers[0]
	}

	// Remove parameters (e.g., "application/json; charset=utf-8" -> "application/json")
	mediaType := strings.Split(contentType, ";")[0]
	mediaType = strings.TrimSpace(mediaType)

	// Find matching parser
	for _, parser := range cn.Parsers {
		if parser.MediaType() == mediaType {
			return parser
		}
	}

	// Default to first parser
	if len(cn.Parsers) > 0 {
		return cn.Parsers[0]
	}

	return nil
}

// parseAcceptHeader parses the Accept header
func parseAcceptHeader(accept string) []string {
	var mediaTypes []string

	parts := strings.Split(accept, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// Remove quality values (e.g., "application/json;q=0.9" -> "application/json")
		if idx := strings.Index(part, ";"); idx != -1 {
			part = part[:idx]
		}
		mediaTypes = append(mediaTypes, strings.TrimSpace(part))
	}

	return mediaTypes
}
