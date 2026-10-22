package validation

import (
	"testing"

	"github.com/forgego/forge/schema"
)

type referenceAddress struct {
	schema.BaseSchema
	Email   string `db:"email"`
	Website string `db:"website"`
}

func (referenceAddress) Meta() schema.Meta { return schema.Meta{TableName: "addresses"} }
func (referenceAddress) Fields() []schema.Field {
	return []schema.Field{schema.EmailField("email"), schema.URLField("website")}
}
func (referenceAddress) Relations() []schema.Relation { return nil }

func TestFieldReferenceEmailAndURLFormats_A157(t *testing.T) {
	v := FromSchema[referenceAddress](referenceAddress{})
	for _, tc := range []struct {
		name      string
		value     referenceAddress
		wantError bool
	}{
		{"valid", referenceAddress{Email: "user@example.com", Website: "https://example.com"}, false},
		{"invalid email", referenceAddress{Email: "not-an-email", Website: "https://example.com"}, true},
		{"invalid URL", referenceAddress{Email: "user@example.com", Website: "not-a-url"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(&tc.value)
			if (err != nil) != tc.wantError {
				t.Fatalf("validation error = %v, want error = %t", err, tc.wantError)
			}
		})
	}
}
