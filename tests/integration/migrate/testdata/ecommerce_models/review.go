package models

import (
	"github.com/forgego/forge/schema"
)

// Review represents a product review
type Review struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Review
func (Review) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("product_id").Required().VerboseName("Product ID").Build(),
		schema.Int64("customer_id").Required().VerboseName("Customer ID").Build(),
		schema.Int64("order_id").Optional().VerboseName("Order ID").Build(),
		schema.Int32("rating").Required().MaxValue(5).MinValue(1).VerboseName("Rating").Build(),
		schema.String("title").MaxLength(200).VerboseName("Review Title").Build(),
		schema.Text("comment").Optional().VerboseName("Review Comment").Build(),
		schema.Bool("is_verified_purchase").Default(false).VerboseName("Is Verified Purchase").Build(),
		schema.Bool("is_approved").Default(false).VerboseName("Is Approved").Build(),
		schema.Int32("helpful_count").Default(0).VerboseName("Helpful Count").Build(),
		schema.JSON("images").Optional().VerboseName("Review Images").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
		schema.DateTime("approved_at").Optional().VerboseName("Approved At").Build(),
	}
}

// Meta returns model metadata
func (Review) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "reviews",
		VerboseName:       "Review",
		VerboseNamePlural: "Reviews",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_review_product_id", Fields: []string{"product_id"}, Unique: false},
			{Name: "idx_review_customer_id", Fields: []string{"customer_id"}, Unique: false},
			{Name: "idx_review_order_id", Fields: []string{"order_id"}, Unique: false},
			{Name: "idx_review_rating", Fields: []string{"rating"}, Unique: false},
			{Name: "idx_review_is_approved", Fields: []string{"is_approved"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"product_id", "customer_id"},
		},
	}
}

// Relations returns all relationship definitions
func (Review) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("product_id", "Product").Required().OnDelete(schema.CascadeCASCADE).RelatedName("reviews").Build(),
		schema.ForeignKey("customer_id", "Customer").Required().OnDelete(schema.CascadeCASCADE).RelatedName("reviews").Build(),
		schema.ForeignKey("order_id", "Order").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("reviews").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Review) Hooks() *schema.ModelHooks {
	return nil
}
