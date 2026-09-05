package models

import "github.com/forgego/forge/schema"

// Product demonstrates various field types and constraints
type Product struct {
	schema.BaseSchema
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.Text("description").Build(),
		schema.Decimal("price").Required().Build(),
		schema.Int64("stock_quantity").Default(0).Build(),
		schema.Bool("is_active").Default(true).Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName: "products",
		Indexes: []schema.Index{
			schema.Index("name").Build(),
			schema.Index("is_active").Build(),
		},
	}
}

func (Product) Relations() []schema.Relation {
	return []schema.Relation{}
}

// Order demonstrates foreign keys and decimal precision
type Order struct {
	schema.BaseSchema
}

func (Order) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("order_number").Required().MaxLength(50).Unique().Build(),
		schema.Int64("customer_id").Required().Build(),
		schema.Decimal("total_amount").Required().Build(),
		schema.String("status").Required().MaxLength(50).Default("pending").Build(),
		schema.DateTime("created_at").Build(),
		schema.DateTime("updated_at").Build(),
	}
}

func (Order) Meta() schema.Meta {
	return schema.Meta{
		TableName: "orders",
		Indexes: []schema.Index{
			schema.Index("customer_id").Build(),
			schema.Index("status").Build(),
		},
	}
}

func (Order) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("customer_id", "Customer").OnDelete("RESTRICT").Build(),
	}
}

// Customer with unique email and default values
type Customer struct {
	schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.String("email").Required().MaxLength(255).Unique().Build(),
		schema.String("name").Required().MaxLength(255).Build(),
		schema.Bool("is_active").Default(true).Build(),
		schema.DateTime("created_at").Build(),
	}
}

func (Customer) Meta() schema.Meta {
	return schema.Meta{
		TableName: "customers",
	}
}

func (Customer) Relations() []schema.Relation {
	return []schema.Relation{}
}