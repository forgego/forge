package templates

import (
	"fmt"
	"strings"
	"text/template"
	"unicode"
)

// GoIdentifier converts a user-supplied name such as a URL resource
// ("blog-posts", "order_items") into an exported CamelCase Go identifier
// ("BlogPosts", "OrderItems"). Every character that cannot appear in an
// identifier separates words, and each word starts upper-case. It fails when
// no letter or digit is left or the result would start with a digit.
func GoIdentifier(name string) (string, error) {
	var b strings.Builder
	startWord := true
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			startWord = true
			continue
		}
		if startWord {
			r = unicode.ToUpper(r)
			startWord = false
		}
		b.WriteRune(r)
	}
	ident := b.String()
	if ident == "" {
		return "", fmt.Errorf("%q has no letters or digits to form a Go identifier", name)
	}
	if first := []rune(ident)[0]; !unicode.IsLetter(first) {
		return "", fmt.Errorf("%q cannot form a Go identifier: it must start with a letter", name)
	}
	return ident, nil
}

// funcMap holds the functions available to templates.
var funcMap = template.FuncMap{
	"goName": GoIdentifier,
}
