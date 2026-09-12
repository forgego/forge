package models

import (
	"context"

	"github.com/forgego/forge/schema"
)

// Customer represents a customer in the ecommerce system
type Customer struct {
	schema.BaseSchema
}

// Fields returns all field definitions for Customer
func (Customer) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64("id").Primary().AutoIncrement().Build(),
		schema.UUID("uuid").Required().Unique().VerboseName("UUID").Build(),
		schema.String("email").Required().Unique().MaxLength(255).VerboseName("Email").Build(),
		schema.String("phone").MaxLength(20).VerboseName("Phone Number").Build(),
		schema.String("first_name").Required().MaxLength(100).VerboseName("First Name").Build(),
		schema.String("last_name").Required().MaxLength(100).VerboseName("Last Name").Build(),
		schema.String("password_hash").Required().MaxLength(255).WriteOnly().VerboseName("Password Hash").Build(),
		schema.Bool("is_active").Default(true).VerboseName("Is Active").Build(),
		schema.Bool("is_verified").Default(false).VerboseName("Is Verified").Build(),
		schema.Bool("is_premium").Default(false).VerboseName("Is Premium Member").Build(),
		schema.JSON("metadata").Optional().VerboseName("Additional Metadata").Build(),
		schema.Decimal("lifetime_value").MaxDigits(12).DecimalPlaces(2).Default(0.0).VerboseName("Lifetime Value").Build(),
		schema.Int64("total_orders").Default(0).VerboseName("Total Orders").Build(),
		schema.DateTime("last_login").Optional().VerboseName("Last Login").Build(),
		schema.DateTime("email_verified_at").Optional().VerboseName("Email Verified At").Build(),
		schema.DateTime("created_at").AutoNowAdd().Build(),
		schema.DateTime("updated_at").AutoNow().Build(),
	}
}

// Meta returns model metadata
func (Customer) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "customers",
		VerboseName:       "Customer",
		VerboseNamePlural: "Customers",
		OrderBy:           []string{"-created_at"},
		Indexes: []schema.Index{
			{Name: "idx_customer_email", Fields: []string{"email"}, Unique: true},
			{Name: "idx_customer_uuid", Fields: []string{"uuid"}, Unique: true},
			{Name: "idx_customer_is_active", Fields: []string{"is_active"}, Unique: false},
			{Name: "idx_customer_created_at", Fields: []string{"created_at"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"email"},
		},
	}
}

// Relations returns all relationship definitions
func (Customer) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKey("billing_address_id", "Address").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("billing_customers").Build(),
		schema.ForeignKey("shipping_address_id", "Address").Optional().OnDelete(schema.CascadeSET_NULL).RelatedName("shipping_customers").Build(),
		schema.OneToOne("profile", "CustomerProfile").Optional().OnDelete(schema.CascadeCASCADE).RelatedName("customer").Build(),
	}
}

// Hooks returns model lifecycle hooks
func (Customer) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			// Password hashing would be handled here if we had a password field
			// Since we only have password_hash, we assume it's already hashed
			// In a real app, you'd hash a plain password field before storing
			return nil
		},
		BeforeUpdate: func(ctx context.Context, instance interface{}) error {
			// If password_hash is being updated, ensure it's hashed
			// This would check if the value changed and hash it
			return nil
		},
	}
}
