package models

import (
	"github.com/forgego/forge/schema"
)

// Brand represents a product brand
type Brand struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Brand
func (Brand) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().Unique().MaxLength(200).VerboseName("Brand Name").Build(),
		schema.String("slug").Required().Unique().MaxLength(250).DBIndex().VerboseName("URL Slug").Build(),
		schema.Text("description").Optional().VerboseName("Description").Build(),
		schema.URL("logo_url").Optional().MaxLength(500).VerboseName("Logo URL").Build(),
		schema.URL("website_url").Optional().MaxLength(500).VerboseName("Website URL").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_featured").Default(false).VerboseName("Is Featured").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Brand) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "brands",
		VerboseName:       "Brand",
		VerboseNamePlural: "Brands",
		OrderBy:           []string{"name"},
		Indexes: []schema.Index{
			{Name: "idx_brand_slug", Fields: []string{"slug"}, Unique: true},
			{Name: "idx_brand_is_active", Fields: []string{"is_active"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Brand) Relations() []schema.Relation {
	return []schema.Relation{
		schema.OneToMany("products", "Product", "brand_id").CascadeOnDelete().Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Brand) Hooks() *schema.ModelHooks {
	return nil
}
