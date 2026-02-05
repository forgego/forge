package generator

import (
	_ "embed"
)

//go:embed templates/combined.tmpl
var combinedTemplate string
