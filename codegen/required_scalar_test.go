package generator

import (
	"strings"
	"testing"
)

func TestGeneratedRequiredScalarValidationAllowsZero(t *testing.T) {
	parser := NewASTParser()
	for _, kind := range []string{"Int64", "Int32", "Float64", "Float32", "Decimal", "Bool"} {
		t.Run(kind, func(t *testing.T) {
			tags := parser.buildValidationTag(kind, map[string]interface{}{"required": true})
			for _, tag := range strings.Split(tags, ",") {
				if tag == "required" {
					t.Fatalf("zero/false is a present scalar value; generated %s tag %q rejects it", kind, tags)
				}
			}
		})
	}
	if got := parser.buildValidationTag("String", map[string]interface{}{"required": true}); got != "required" {
		t.Fatalf("required text must still reject an empty string: %q", got)
	}
}
