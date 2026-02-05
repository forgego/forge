package widgets

import (
	"fmt"
)

// AutosuggestWidget provides autosuggest functionality for filter inputs
type AutosuggestWidget struct {
	Suggestions []string
	MinLength   int
}

// NewAutosuggestWidget creates a new autosuggest widget
func NewAutosuggestWidget(suggestions []string, minLength int) *AutosuggestWidget {
	return &AutosuggestWidget{
		Suggestions: suggestions,
		MinLength:   minLength,
	}
}

// Render renders the autosuggest widget
func (w *AutosuggestWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	valueStr := ""
	if value != nil {
		valueStr = fmt.Sprintf("%v", value)
	}

	html := fmt.Sprintf(`<input type="text" name="%s" value="%s"`, name, valueStr)
	html += ` class="form-control autosuggest" data-min-length="` + fmt.Sprintf("%d", w.MinLength) + `"`
	
	for k, v := range attrs {
		html += fmt.Sprintf(` %s="%s"`, k, v)
	}
	
	html += `>`
	html += `<div class="autosuggest-dropdown"></div>`
	
	// Add JavaScript for autosuggest
	html += `<script>initAutosuggest('` + name + `', ` + fmt.Sprintf("%v", w.Suggestions) + `);</script>`
	
	return html, nil
}

// Parse parses the widget value
func (w *AutosuggestWidget) Parse(value string) (interface{}, error) {
	return value, nil
}
