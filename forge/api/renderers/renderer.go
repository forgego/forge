package renderers

import (
	"io"
)

// Renderer is the interface for response renderers
type Renderer interface {
	// Render renders data to bytes
	Render(data interface{}) ([]byte, error)
	// MediaType returns the media type for this renderer
	MediaType() string
	// RenderToWriter renders data directly to a writer
	RenderToWriter(w io.Writer, data interface{}) error
}

// RendererList is a list of renderers
type RendererList []Renderer

// GetRenderer returns a renderer by media type
func (rl RendererList) GetRenderer(mediaType string) Renderer {
	for _, r := range rl {
		if r.MediaType() == mediaType {
			return r
		}
	}
	return nil
}

// GetMediaTypes returns all media types supported by this list
func (rl RendererList) GetMediaTypes() []string {
	types := make([]string, len(rl))
	for i, r := range rl {
		types[i] = r.MediaType()
	}
	return types
}
