package models

import (
	"github.com/forgego/forge/schema"
)

// Supplier represents a product supplier
type Supplier struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Supplier
func (Supplier) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().Unique().MaxLength(200).VerboseName("Supplier Name").Build(),
		schema.String("code").Required().Unique().MaxLength(50).VerboseName("Supplier Code").Build(),
		schema.String("contact_name").MaxLength(200).VerboseName("Contact Name").Build(),
		schema.Email("email").MaxLength(255).VerboseName("Email").Build(),
		schema.String("phone").MaxLength(20).VerboseName("Phone Number").Build(),
		schema.String("address_line1").MaxLength(255).VerboseName("Address Line 1").Build(),
		schema.String("address_line2").MaxLength(255).VerboseName("Address Line 2").Build(),
		schema.String("city").MaxLength(100).VerboseName("City").Build(),
		schema.String("state").MaxLength(100).VerboseName("State/Province").Build(),
		schema.String("postal_code").MaxLength(20).VerboseName("Postal Code").Build(),
		schema.String("country").MaxLength(2).VerboseName("Country Code").Build(),
		schema.URL("website_url").Optional().MaxLength(500).VerboseName("Website URL").Build(),
		schema.String("payment_terms").MaxLength(100).VerboseName("Payment Terms").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Supplier) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "suppliers",
		VerboseName:       "Supplier",
		VerboseNamePlural: "Suppliers",
		OrderBy:           []string{"name"},
		Indexes: []schema.Index{
			{Name: "idx_supplier_code", Fields: []string{"code"}, Unique: true},
			{Name: "idx_supplier_email", Fields: []string{"email"}, Unique: false},
			{Name: "idx_supplier_is_active", Fields: []string{"is_active"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Supplier) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToMany("products", "Product", "supplier_id").CascadeOnDelete().Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Supplier) Hooks() *schema.ModelHooks {
	return nil
}
