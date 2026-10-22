package validation

import (
	"github.com/forgego/forge/orm"
	"testing"
)

type validatedRecord struct{ Name string }

func TestTypedValidatorValidatesConfiguredField(t *testing.T) {
	builder := For[validatedRecord]()
	FieldFor(builder, orm.NewField[string]("Name", "")).MinLength(3)
	validator := builder.Build()
	if err := validator.Validate(&validatedRecord{Name: "ab"}); err == nil {
		t.Fatal("short field should fail")
	}
	if err := validator.Validate(&validatedRecord{Name: "abcd"}); err != nil {
		t.Fatal(err)
	}
}
