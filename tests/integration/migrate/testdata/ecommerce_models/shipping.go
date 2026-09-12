package models

import (
	"github.com/forgego/forge/schema"
)

// Shipping represents shipping information for an order
type Shipping struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Shipping
func (Shipping) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.Int64("order_id").Required().Unique().VerboseName("Order ID").Build(),
		schema.String("carrier").MaxLength(100).VerboseName("Shipping Carrier").Build(),
		schema.String("service").MaxLength(100).VerboseName("Shipping Service").Build(),
		schema.String("tracking_number").MaxLength(100).VerboseName("Tracking Number").Build(),
		schema.URL("tracking_url").Optional().MaxLength(500).VerboseName("Tracking URL").Build(),
		schema.String("status").MaxLength(50).Choices(
			schema.Choice{Value: "pending", Label: "Pending"},
			schema.Choice{Value: "label_created", Label: "Label Created"},
			schema.Choice{Value: "in_transit", Label: "In Transit"},
			schema.Choice{Value: "out_for_delivery", Label: "Out for Delivery"},
			schema.Choice{Value: "delivered", Label: "Delivered"},
			schema.Choice{Value: "exception", Label: "Exception"},
			schema.Choice{Value: "returned", Label: "Returned"},
		).Default("pending").VerboseName("Shipping Status").Build(),
		schema.Decimal("cost").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Shipping Cost").Build(),
		schema.Float64("weight").Optional().VerboseName("Package Weight (kg)").Build(),
		schema.String("dimensions").MaxLength(100).VerboseName("Package Dimensions").Build(),
		schema.DateTime("shipped_at").Optional().VerboseName("Shipped At").Build(),
		schema.DateTime("estimated_delivery_at").Optional().VerboseName("Estimated Delivery At").Build(),
		schema.DateTime("delivered_at").Optional().VerboseName("Delivered At").Build(),
		schema.JSON("carrier_data").Optional().VerboseName("Carrier Response Data").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Shipping) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "shipping",
		VerboseName:       "Shipping",
		VerboseNamePlural: "Shipping Records",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_shipping_order_id", Fields: []string{"order_id"}, Unique: true},
			{Name: "idx_shipping_tracking_number", Fields: []string{"tracking_number"}, Unique: false},
			{Name: "idx_shipping_status", Fields: []string{"status"}, Unique: false},
		},
	}
}

// Relations returns all relationship definitions
func (Shipping) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("order_id", "Order").Required().OnDelete(schema.CascadeCASCADE).RelatedName("shipping").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Shipping) Hooks() *schema.ModelHooks {
	return nil
}
