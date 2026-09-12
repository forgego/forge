package models

import (
	"github.com/forgego/forge/schema"
)

// Address represents a customer address
type Address struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Address
func (Address) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("customer_id").Required().VerboseName("Customer ID").Build(),
		schema.String("type").Required().MaxLength(20).Choices(
			schema.Choice{Value: "billing", Label: "Billing"},
			schema.Choice{Value: "shipping", Label: "Shipping"},
			schema.Choice{Value: "both", Label: "Both"},
		).Default("shipping").VerboseName("Address Type").Build(),
		schema.String("first_name").Required().MaxLength(100).VerboseName("First Name").Build(),
		schema.String("last_name").Required().MaxLength(100).VerboseName("Last Name").Build(),
		schema.String("company").MaxLength(200).VerboseName("Company").Build(),
		schema.String("address_line1").Required().MaxLength(255).VerboseName("Address Line 1").Build(),
		schema.String("address_line2").MaxLength(255).VerboseName("Address Line 2").Build(),
		schema.String("city").Required().MaxLength(100).VerboseName("City").Build(),
		schema.String("state").Required().MaxLength(100).VerboseName("State/Province").Build(),
		schema.String("postal_code").Required().MaxLength(20).VerboseName("Postal Code").Build(),
		schema.String("country").Required().MaxLength(2).VerboseName("Country Code").Build(),
		schema.String("phone").MaxLength(20).VerboseName("Phone Number").Build(),
		schema.Bool("is_default").Default(false).VerboseName("Is Default Address").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Address) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "addresses",
		VerboseName:       "Address",
		VerboseNamePlural: "Addresses",
		OrderBy:           []string{"-is_default", "-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_address_customer_id", Fields: []string{"customer_id"}, Unique: false},
			{Name: "idx_address_type", Fields: []string{"type"}, Unique: false},
			{Name: "idx_address_country", Fields: []string{"country"}, Unique: false},
			{Name: "idx_address_postal_code", Fields: []string{"postal_code"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Address) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("customer_id", "Customer").Required().OnDelete(schema.CascadeCASCADE).RelatedName("addresses").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Address) Hooks() *schema.ModelHooks {
	return nil
}
