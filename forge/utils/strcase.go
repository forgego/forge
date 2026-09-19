package utils

import (
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

// ToSnake converts a string to snake_case
func ToSnake(s string) string {
	return strcase.ToSnake(s)
}

// Pluralize applies the common English pluralization rules used by generated
// route slugs. Callers should normalize compound names before pluralizing.
func Pluralize(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") {
		return s + "es"
	}
	if strings.HasSuffix(lower, "y") && len(lower) > 1 {
		previous := rune(lower[len(lower)-2])
		if !strings.ContainsRune("aeiou", unicode.ToLower(previous)) {
			return s[:len(s)-1] + "ies"
		}
	}
	return s + "s"
}

// ToCamel converts a string to camelCase
func ToCamel(s string) string {
	return strcase.ToCamel(s)
}

// ToPascal converts a string to PascalCase
func ToPascal(s string) string {
	camel := strcase.ToCamel(s)
	if camel == "" {
		return camel
	}
	// Capitalize first letter if it's lowercase
	if camel[0] >= 'a' && camel[0] <= 'z' {
		return string(camel[0]-32) + camel[1:]
	}
	return camel
}

// ToKebab converts a string to kebab-case
func ToKebab(s string) string {
	return strcase.ToKebab(s)
}

// ToKebabPlural converts a model name to its URL-slug plural form: the name
// is first normalized to kebab-case (the established form used by
// hand-written routes, e.g. product-variants) and then pluralized with
// Pluralize, which remains the single pluralizer.
func ToKebabPlural(s string) string {
	return Pluralize(ToKebab(s))
}

// ToLowerCamel converts a string to lowerCamelCase (same as ToCamel)
func ToLowerCamel(s string) string {
	return strcase.ToCamel(s)
}
