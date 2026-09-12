package models

import (
	"github.com/forgego/forge/schema"
)

// Warehouse represents a warehouse location
type Warehouse struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Warehouse
func (Warehouse) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(200).VerboseName("Warehouse Name").Build(),
		schema.String("code").Required().Unique().MaxLength(50).VerboseName("Warehouse Code").Build(),
		schema.String("address_line1").Required().MaxLength(255).VerboseName("Address Line 1").Build(),
		schema.String("address_line2").MaxLength(255).VerboseName("Address Line 2").Build(),
		schema.String("city").Required().MaxLength(100).VerboseName("City").Build(),
		schema.String("state").Required().MaxLength(100).VerboseName("State/Province").Build(),
		schema.String("postal_code").Required().MaxLength(20).VerboseName("Postal Code").Build(),
		schema.String("country").Required().MaxLength(2).VerboseName("Country Code").Build(),
		schema.String("phone").MaxLength(20).VerboseName("Phone Number").Build(),
		schema.String("email").MaxLength(255).VerboseName("Email").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_primary").Default(false).VerboseName("Is Primary Warehouse").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Warehouse) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "warehouses",
		VerboseName:       "Warehouse",
		VerboseNamePlural: "Warehouses",
		OrderBy:           []string{"-is_primary", "name"},
		Indexes: []schema.Index{
			{Name: "idx_warehouse_code", Fields: []string{"code"}, Unique: true},
			{Name: "idx_warehouse_is_active", Fields: []string{"is_active"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Warehouse) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToMany("inventory", "Inventory", "warehouse_id").CascadeOnDelete().Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Warehouse) Hooks() *schema.ModelHooks {
	return nil
}
