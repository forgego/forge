package renderers

import (
	"bytes"
	"html/template"
	"io"
)

// HTMLRenderer renders data as HTML (for browsable API)
type HTMLRenderer struct {
	// Template is the HTML template for rendering
	Template *template.Template
}

// NewHTMLRenderer creates a new HTML renderer
func NewHTMLRenderer() *HTMLRenderer {
	return &HTMLRenderer{
		Template: getDefaultTemplate(),
	}
}

// Render renders data to HTML bytes
func (r *HTMLRenderer) Render(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := r.RenderToWriter(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MediaType returns the HTML media type
func (r *HTMLRenderer) MediaType() string {
	return "text/html"
}

// RenderToWriter renders data directly to a writer
func (r *HTMLRenderer) RenderToWriter(w io.Writer, data interface{}) error {
	tmpl := r.Template
	if tmpl == nil {
		tmpl = getDefaultTemplate()
	}
	return tmpl.Execute(w, data)
}

// getDefaultTemplate returns a default HTML template
func getDefaultTemplate() *template.Template {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
	<title>API Response</title>
	<style>
		body { font-family: monospace; padding: 20px; }
		pre { background: #f5f5f5; padding: 10px; border-radius: 5px; }
	</style>
</head>
<body>
	<h1>API Response</h1>
	<pre>{{.}}</pre>
</body>
</html>
`
	return template.Must(template.New("default").Parse(tmpl))
}
