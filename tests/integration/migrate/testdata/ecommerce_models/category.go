package models

import (
	"github.com/forgego/forge/schema"
)

// Category represents a product category with hierarchical structure
type Category struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Category
func (Category) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.UUID("uuid").Required().Unique().VerboseName("UUID").Build(),
		schema.Int64("parent_id").Optional().VerboseName("Parent Category ID").Build(),
		schema.String("name").Required().MaxLength(200).VerboseName("Category Name").Build(),
		schema.String("slug").Required().Unique().MaxLength(250).DBIndex().VerboseName("URL Slug").Build(),
		schema.Text("description").Optional().VerboseName("Description").Build(),
		schema.URL("image_url").Optional().MaxLength(500).VerboseName("Image URL").Build(),
		schema.Int32("sort_order").Default(0).VerboseName("Sort Order").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_featured").Default(false).VerboseName("Is Featured").Build(),
		schema.Int32("level").Default(0).VerboseName("Category Level").Build(),
		schema.String("path").MaxLength(500).VerboseName("Category Path").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Category) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "categories",
		VerboseName:       "Category",
		VerboseNamePlural: "Categories",
		OrderBy:           []string{"sort_order", "name"},
		Indexes: []schema.Index{
			{Name: "idx_category_slug", Fields: []string{"slug"}, Unique: true},
			{Name: "idx_category_parent_id", Fields: []string{"parent_id"}, Unique: false},
			{Name: "idx_category_is_active", Fields: []string{"is_active"}, Unique: false},
			{Name: "idx_category_level", Fields: []string{"level"}, Unique: false},
			{Name: "idx_category_path", Fields: []string{"path"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Category) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("parent_id", "Category").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("children").Build(),
		schema.ManyToMany("products", "Product").RelatedName("categories").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Category) Hooks() *schema.ModelHooks {
	return nil
}
