package widgets

import (
	"fmt"
)

// SQLPreviewWidget shows SQL preview and estimated cost
type SQLPreviewWidget struct {
	SQLPreview    string
	EstimatedCost int
	EstimatedRows int64
}

// NewSQLPreviewWidget creates a new SQL preview widget
func NewSQLPreviewWidget(sqlPreview string, estimatedCost int, estimatedRows int64) *SQLPreviewWidget {
	return &SQLPreviewWidget{
		SQLPreview:    sqlPreview,
		EstimatedCost: estimatedCost,
		EstimatedRows: estimatedRows,
	}
}

// Render renders the SQL preview widget
func (w *SQLPreviewWidget) Render(name string, value interface{}, attrs map[string]string) (string, error) {
	html := `<div class="sql-preview">`
	html += `<h4>Query Preview</h4>`
	html += `<pre class="sql-code">` + fmt.Sprintf("%s", w.SQLPreview) + `</pre>`
	html += `<div class="query-stats">`
	html += `<span>Estimated Cost: ` + fmt.Sprintf("%d", w.EstimatedCost) + `</span>`
	html += `<span>Estimated Rows: ` + fmt.Sprintf("%d", w.EstimatedRows) + `</span>`
	html += `</div>`
	html += `</div>`
	return html, nil
}

// Parse is not applicable for preview widget
func (w *SQLPreviewWidget) Parse(value string) (interface{}, error) {
	return nil, fmt.Errorf("SQL preview widget does not parse values")
}
