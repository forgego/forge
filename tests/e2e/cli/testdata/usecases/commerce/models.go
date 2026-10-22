package work

import "github.com/forgego/forge/schema"

type Member struct{ MemberGenerated }

func (Member) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("email", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.StringField("api_token", schema.Required(), schema.WriteOnly(), schema.Editable(false)),
		schema.BoolField("is_staff", schema.Default(false), schema.WriteOnly(), schema.Editable(false)),
	}
}
func (Member) Meta() schema.Meta { return schema.Meta{TableName: "members", VerboseName: "Member"} }
func (Member) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Member) Hooks() *schema.ModelHooks { return nil }

type Product struct{ ProductGenerated }

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("sku", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.Int64Field("price_cents", schema.Required(), schema.MinValue(0)),
		schema.Int64Field("stock", schema.Required(), schema.Default(0), schema.MinValue(0)),
		schema.BoolField("active", schema.Default(true)),
		schema.StringField("internal_token", schema.Default("system-only"), schema.WriteOnly(), schema.Editable(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
		// ACCEPTANCE EXTRA FIELD
	}
}
func (Product) Meta() schema.Meta {
	return schema.Meta{TableName: "products", VerboseName: "Product", Indexes: []schema.Index{{Name: "idx_products_name", Fields: []string{"name"}}}, OrderBy: []string{"name"}}
}
func (Product) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Product) Hooks() *schema.ModelHooks { return nil }

type Customer struct{ CustomerGenerated }

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
	}
}
func (Customer) Meta() schema.Meta {
	return schema.Meta{TableName: "customers", VerboseName: "Customer"}
}
func (Customer) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Customer) Hooks() *schema.ModelHooks { return nil }

type Order struct{ OrderGenerated }

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("customer_id", schema.Required()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("status", schema.Required(), schema.MaxLength(80), schema.Default("draft")),
		schema.Int64Field("total_cents", schema.Required(), schema.Default(0), schema.MinValue(0)),
	}
}
func (Order) Meta() schema.Meta { return schema.Meta{TableName: "orders", VerboseName: "Order"} }
func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadePROTECT)),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Order) Hooks() *schema.ModelHooks { return nil }

type OrderLine struct{ OrderLineGenerated }

func (OrderLine) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("order_id", schema.Required()),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("quantity", schema.Required(), schema.MinValue(1)),
		schema.Int64Field("unit_price_cents", schema.Required(), schema.MinValue(0)),
	}
}
func (OrderLine) Meta() schema.Meta {
	return schema.Meta{TableName: "order_lines", VerboseName: "OrderLine"}
}
func (OrderLine) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("order_id", "Order", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("product_id", "Product", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (OrderLine) Hooks() *schema.ModelHooks { return nil }

type StockMovement struct{ StockMovementGenerated }

func (StockMovement) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("product_id", schema.Required()),
		schema.Int64Field("delta", schema.Required()),
		schema.StringField("reason", schema.Required(), schema.MaxLength(80)),
	}
}
func (StockMovement) Meta() schema.Meta {
	return schema.Meta{TableName: "stock_movements", VerboseName: "StockMovement"}
}
func (StockMovement) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("product_id", "Product", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (StockMovement) Hooks() *schema.ModelHooks { return nil }
