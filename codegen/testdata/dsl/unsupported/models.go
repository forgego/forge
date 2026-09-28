// Package blog is the unsupported schema DSL fixture for code generation.
//
// Each line marked `// want:` holds an expression that generation cannot
// evaluate from source. `forge generate` warns about each one and
// `forge generate --strict` rejects the package. codegen/dsl_contract_test.go
// asserts exactly one diagnostic per marker, at that marker's line, with the
// marker's text in the message.
package blog

import "github.com/forgego/forge/schema"

type Article struct {
	schema.BaseSchema
}

var translated bool

func titleField() schema.Field { return schema.StringField("title") }

func lengthOption() schema.FieldOpt { return schema.MaxLength(200) }

func auditFields() []schema.Field { return nil }

func tableName() string { return "articles" }

func (Article) Fields() []schema.Field {
	fields := []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		titleField(), // want: helper call titleField
		schema.StringField("slug", lengthOption()), // want: helper call lengthOption
	}
	for _, locale := range []string{"en", "fr"} { // want: loops cannot be evaluated
		fields = append(fields, schema.StringField("title_"+locale))
	}
	if translated { // want: conditional assembly cannot be evaluated
		fields = append(fields, schema.StringField("translation"))
	}
	fields = append(fields, auditFields()...) // want: append of a computed slice
	return fields
}

func (Article) Meta() schema.Meta {
	return schema.Meta{
		TableName: tableName(), // want: helper call tableName
	}
}

func (Article) Relations() []schema.Relation {
	return []schema.Relation{}
}
