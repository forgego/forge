// Package csvsafe prevents spreadsheet formula interpretation of CSV cells.
package csvsafe

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Cell prefixes formula-like text with a single quote. Numeric Go values keep
// their representation, including a leading minus sign.
func Cell(value any, formatted string) string {
	v := reflect.ValueOf(value)
	for v.IsValid() && (v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer) {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	if v.IsValid() {
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Uintptr, reflect.Float32, reflect.Float64:
			return formatted
		}
	}
	if number, ok := value.(json.Number); ok && len(number) > 0 &&
		(number[0] == '-' || number[0] >= '0' && number[0] <= '9') && json.Valid([]byte(number)) {
		return formatted
	}
	if strings.HasPrefix(formatted, "=") || strings.HasPrefix(formatted, "+") ||
		strings.HasPrefix(formatted, "-") || strings.HasPrefix(formatted, "@") ||
		strings.HasPrefix(formatted, "\t") || strings.HasPrefix(formatted, "\r") {
		return "'" + formatted
	}
	return formatted
}
