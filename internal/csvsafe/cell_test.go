package csvsafe

import (
	"encoding/json"
	"testing"
)

func TestCellEscapesFormulaPrefixesAndKeepsNumericValues(t *testing.T) {
	for _, tc := range []struct {
		name      string
		value     any
		formatted string
		want      string
	}{
		{"equals", "=1+1", "=1+1", "'=1+1"},
		{"plus", "+SUM(1)", "+SUM(1)", "'+SUM(1)"},
		{"minus text", "-5", "-5", "'-5"},
		{"at", "@SUM(1)", "@SUM(1)", "'@SUM(1)"},
		{"tab", "\t=1+1", "\t=1+1", "'\t=1+1"},
		{"carriage return", "\r=1+1", "\r=1+1", "'\r=1+1"},
		{"integer", -5, "-5", "-5"},
		{"float", -5.5, "-5.5", "-5.5"},
		{"json number", json.Number("-5"), "-5", "-5"},
		{"invalid json number", json.Number("=1+1"), "=1+1", "'=1+1"},
		{"plain", "hello", "hello", "hello"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Cell(tc.value, tc.formatted); got != tc.want {
				t.Fatalf("Cell(%q) = %q, want %q", tc.formatted, got, tc.want)
			}
		})
	}
}
