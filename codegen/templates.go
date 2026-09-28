package generator

import (
	_ "embed"
)

//go:embed templates/combined.tmpl
var combinedTemplate string

//go:embed templates/api.tmpl
var apiTemplate string
