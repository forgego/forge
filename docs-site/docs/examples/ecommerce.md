---
sidebar_position: 2
---

# E-commerce Example

A comprehensive e-commerce application demonstrating complex models and relationships.

## Overview

This example demonstrates:

- Complex model relationships (OneToOne, ForeignKey, ManyToMany)
- All PostgreSQL data types
- Hierarchical categories
- Product variants and inventory
- Orders and payments

## Key Models

### Product

```go
type Product struct {
    schema.BaseSchema
}

func (Product) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().MaxLength(200).Build(),
        schema.String("sku").Unique().MaxLength(100).Build(),
        schema.Text("description").Build(),
        schema.Decimal("price").
            MaxDigits(10).
            DecimalPlaces(2).
            Required().
            Build(),
        schema.Bool("active").Default(true).Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
    }
}

func (Product) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("brand", "Brand").
            OnDelete(schema.SetNull),
        relations.ForeignKey("supplier", "Supplier").
            OnDelete(schema.SetNull),
        relations.ManyToMany("categories", "Category").
            Through("product_categories"),
        relations.OneToMany("variants", "ProductVariant").
            RelatedName("product"),
    }
}
```

### Order

```go
type Order struct {
    schema.BaseSchema
}

func (Order) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("order_number").Unique().MaxLength(50).Build(),
        schema.Decimal("total").
            MaxDigits(10).
            DecimalPlaces(2).
            Required().
            Build(),
        schema.String("status").
            Choices("pending", "processing", "shipped", "delivered", "cancelled").
            Default("pending").
            Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
    }
}

func (Order) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("customer", "Customer").
            Required().
            OnDelete(schema.Cascade),
        relations.ForeignKey("billing_address", "Address").
            Required(),
        relations.ForeignKey("shipping_address", "Address").
            Required(),
        relations.OneToMany("items", "OrderItem").
            RelatedName("order"),
        relations.OneToOne("payment", "Payment"),
        relations.OneToOne("shipping", "Shipping"),
    }
}
```

### Customer with Profile

```go
type Customer struct {
    schema.BaseSchema
}

func (Customer) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("email").Unique().Required().MaxLength(255).Build(),
        schema.String("password_hash").Required().MaxLength(255).Build(),
        schema.Bool("is_active").Default(true).Build(),
        schema.Time("created_at").AutoNowAdd().Build(),
    }
}

func (Customer) Relations() []schema.Relation {
    return []schema.Relation{
        relations.OneToOne("profile", "CustomerProfile").
            OnDelete(schema.Cascade),
        relations.OneToMany("addresses", "Address").
            RelatedName("customer"),
        relations.OneToMany("orders", "Order").
            RelatedName("customer"),
    }
}

type CustomerProfile struct {
    schema.BaseSchema
}

func (CustomerProfile) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("first_name").MaxLength(100).Build(),
        schema.String("last_name").MaxLength(100).Build(),
        schema.String("phone").MaxLength(20).Build(),
        schema.Date("date_of_birth").Build(),
    }
}

func (CustomerProfile) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("customer", "Customer").
            Required().
            Unique().
            OnDelete(schema.Cascade),
    }
}
```

## Hierarchical Categories

```go
type Category struct {
    schema.BaseSchema
}

func (Category) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().MaxLength(100).Build(),
        schema.String("slug").Unique().MaxLength(100).Build(),
        schema.Text("description").Build(),
    }
}

func (Category) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("parent", "Category").
            OnDelete(schema.Cascade).
            RelatedName("children"),
    }
}
```

## Product Variants and Inventory

```go
type ProductVariant struct {
    schema.BaseSchema
}

func (ProductVariant) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("sku").Unique().MaxLength(100).Build(),
        schema.String("name").MaxLength(200).Build(),
        schema.Decimal("price").
            MaxDigits(10).
            DecimalPlaces(2).
            Build(),
        schema.JSON("attributes").Build(), // Size, color, etc.
    }
}

func (ProductVariant) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("product", "Product").
            Required().
            OnDelete(schema.Cascade),
        relations.OneToMany("inventory", "Inventory").
            RelatedName("variant"),
    }
}

type Inventory struct {
    schema.BaseSchema
}

func (Inventory) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.Int32("quantity").Required().Default(0).Build(),
        schema.Int32("reserved").Default(0).Build(),
    }
}

func (Inventory) Relations() []schema.Relation {
    return []schema.Relation{
        relations.ForeignKey("product", "Product").
            Required().
            OnDelete(schema.Cascade),
        relations.ForeignKey("variant", "ProductVariant").
            OnDelete(schema.Cascade),
        relations.ForeignKey("warehouse", "Warehouse").
            Required().
            OnDelete(schema.Cascade),
    }
}
```

## Usage Examples

### Create Order with Items

```go
order := &models.Order{
    OrderNumber: generateOrderNumber(),
    Customer:    customer,
    BillingAddress: billingAddress,
    ShippingAddress: shippingAddress,
    Status:      "pending",
    Total:       calculateTotal(items),
}

err := models.Order.Objects.Create(ctx, order)

// Create order items
for _, item := range items {
    orderItem := &models.OrderItem{
        Order:   order,
        Product: item.Product,
        Variant: item.Variant,
        Quantity: item.Quantity,
        Price:     item.Price,
    }
    err := models.OrderItem.Objects.Create(ctx, orderItem)
}
```

### Get Products by Category

```go
products, err := models.Product.Objects.
    Filter(models.Product.Fields.Categories.Contains(categoryID)).
    Filter(models.Product.Fields.Active.Equals(true)).
    PrefetchRelated("brand", "variants").
    All(ctx)
```

### Check Inventory

```go
inventory, err := models.Inventory.Objects.
    Filter(models.Inventory.Fields.Product.Equals(productID)).
    Filter(models.Inventory.Fields.Variant.Equals(variantID)).
    Filter(models.Inventory.Fields.Warehouse.Equals(warehouseID)).
    First(ctx)

available := inventory.Quantity - inventory.Reserved
```

## See Also

- [E-commerce Website Sample](/docs/examples/ecommerce-website) - Run the full sample locally and explore the UI.
- [Models Guide](/docs/guides/models) - Model definitions
- [Relations Reference](/docs/api-reference/relations) - Relationship types
- [Library Example](/docs/examples/library) - Another complex example
