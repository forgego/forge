package utils

import (
	"github.com/iancoleman/strcase"
)

// ToSnake converts a string to snake_case
func ToSnake(s string) string {
	return strcase.ToSnake(s)
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

// ToLowerCamel converts a string to lowerCamelCase (same as ToCamel)
func ToLowerCamel(s string) string {
	return strcase.ToCamel(s)
}
