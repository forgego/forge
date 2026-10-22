package orm

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
)

type documentedProduct struct {
	schema.BaseSchema
	ID      int64   `db:"id"`
	Name    string  `db:"name"`
	Price   float64 `db:"price"`
	InStock bool    `db:"in_stock"`
	events  *[]string
	db      *db.DB
}

type documentedFields struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Email     string    `db:"email"`
	Balance   float64   `db:"balance"`
	CreatedAt time.Time `db:"created_at"`
	Status    string    `db:"status"`
}

type documentedOrder struct {
	schema.BaseSchema
	ID             int64   `db:"id"`
	Subtotal       float64 `db:"subtotal"`
	TaxAmount      float64 `db:"tax_amount"`
	ShippingAmount float64 `db:"shipping_amount"`
	TotalAmount    float64 `db:"total_amount"`
	events         *[]string
}

func (documentedOrder) Meta() schema.Meta { return schema.Meta{TableName: "documented_orders"} }
func (documentedOrder) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Float64Field("subtotal"), schema.Float64Field("tax_amount"),
		schema.Float64Field("shipping_amount"), schema.Float64Field("total_amount"),
	}
}
func (documentedOrder) Hooks() *schema.ModelHooks {
	return schema.NewModelHooks().
		WithClean(func(instance interface{}) error {
			order := instance.(*documentedOrder)
			if order.ShippingAmount < 0 {
				return fmt.Errorf("shipping amount cannot be negative")
			}
			return nil
		}).
		WithBeforeSave(func(_ context.Context, instance interface{}) error {
			order := instance.(*documentedOrder)
			order.TotalAmount = order.Subtotal + order.TaxAmount + order.ShippingAmount
			return nil
		}).
		WithAfterCreate(func(_ context.Context, instance interface{}) error {
			order := instance.(*documentedOrder)
			if order.events != nil {
				*order.events = append(*order.events, "created")
			}
			return nil
		})
}

func TestDocumentedOrderHooksComputeAndValidateBeforeInsert(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE documented_orders (id INTEGER PRIMARY KEY AUTOINCREMENT, subtotal REAL, tax_amount REAL, shipping_amount REAL, total_amount REAL)`); err != nil {
		t.Fatal(err)
	}
	manager, err := NewManagerWithDB[documentedOrder]("documented_orders", database)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	var events []string
	order := &documentedOrder{Subtotal: 100, TaxAmount: 10, ShippingAmount: 5, events: &events}
	if err := manager.Create(ctx, order); err != nil {
		t.Fatal(err)
	}
	got, err := manager.Get(ctx, order.ID)
	if err != nil || got.TotalAmount != 115 || !reflect.DeepEqual(events, []string{"created"}) {
		t.Fatalf("stored order = %+v, events = %v, err = %v", got, events, err)
	}
	if err := manager.Create(ctx, &documentedOrder{ShippingAmount: -1}); err == nil {
		t.Fatal("Clean accepted a negative shipping amount")
	}
	count, err := manager.Count(ctx)
	if err != nil || count != 1 {
		t.Fatalf("invalid order reached database: count = %d, err = %v", count, err)
	}
}

func (documentedFields) Meta() schema.Meta { return schema.Meta{TableName: "documented_fields"} }
func (documentedFields) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("email", schema.MaxLength(255), schema.Required(), schema.Unique(), schema.DBIndex()),
		schema.DecimalField("balance", schema.MaxDigits(12), schema.DecimalPlaces(2), schema.Required(), schema.MinValue(0.0)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
		schema.StringField("status", schema.ChoicesOpts(
			schema.NewChoice("draft", "Draft Order"),
			schema.NewChoice("processing", "In Processing"),
			schema.NewChoice("shipped", "Shipped & In Transit"),
			schema.NewChoice("delivered", "Delivered")), schema.DBDefault("'draft'")),
	}
}

func TestDocumentedFunctionalFieldSamplesRoundTrip(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE documented_fields (id INTEGER PRIMARY KEY AUTOINCREMENT, email VARCHAR(255) NOT NULL UNIQUE, balance NUMERIC(12, 2) NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, status TEXT DEFAULT 'draft')`); err != nil {
		t.Fatal(err)
	}
	manager, err := NewManagerWithDB[documentedFields]("documented_fields", database)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	entry := &documentedFields{Email: "buyer@example.com", Balance: 12.5, Status: "draft"}
	if err := manager.Create(ctx, entry); err != nil {
		t.Fatal(err)
	}
	got, err := manager.Get(ctx, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("required_unique_email", func(t *testing.T) {
		if got.Email != entry.Email {
			t.Fatalf("email = %q", got.Email)
		}
		duplicate := &documentedFields{Email: entry.Email, Balance: 1, Status: "draft"}
		if err := manager.Create(ctx, duplicate); err == nil {
			t.Fatal("unique email allowed a duplicate")
		}
	})
	t.Run("decimal_balance", func(t *testing.T) {
		if got.Balance != entry.Balance {
			t.Fatalf("balance = %v", got.Balance)
		}
	})
	t.Run("automatic_creation_time", func(t *testing.T) {
		if got.CreatedAt.IsZero() {
			t.Fatal("created_at was not filled by the database")
		}
	})
	t.Run("choice_labels_and_status", func(t *testing.T) {
		if got.Status != "draft" {
			t.Fatalf("status = %q", got.Status)
		}
		if len(entry.Fields()[4].Choices) != 4 || entry.Fields()[4].Choices[2].Label != "Shipped & In Transit" {
			t.Fatalf("documented choices lost: %+v", entry.Fields()[4].Choices)
		}
	})
}

func (documentedProduct) Meta() schema.Meta { return schema.Meta{TableName: "documented_products"} }
func (documentedProduct) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.Float64Field("price", schema.Required()),
		schema.BoolField("in_stock", schema.Required()),
	}
}
func (documentedProduct) Hooks() *schema.ModelHooks {
	return schema.NewModelHooks().
		WithBeforeDelete(func(_ context.Context, instance interface{}) error {
			p := instance.(*documentedProduct)
			if p.events == nil {
				return nil
			}
			var count int
			if err := p.db.QueryRow(`SELECT count(*) FROM documented_products WHERE id = ?`, p.ID).Scan(&count); err != nil {
				return err
			}
			if count != 1 {
				return fmt.Errorf("before delete saw %d rows", count)
			}
			*p.events = append(*p.events, "before")
			return nil
		}).
		WithAfterDelete(func(_ context.Context, instance interface{}) error {
			p := instance.(*documentedProduct)
			if p.events == nil {
				return nil
			}
			var count int
			if err := p.db.QueryRow(`SELECT count(*) FROM documented_products WHERE id = ?`, p.ID).Scan(&count); err != nil {
				return err
			}
			if count != 0 {
				return fmt.Errorf("after delete saw %d rows", count)
			}
			*p.events = append(*p.events, "after")
			return nil
		})
}

func documentedProductManager(t *testing.T) *Manager[documentedProduct] {
	t.Helper()
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	if _, err := database.Exec(`CREATE TABLE documented_products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, price REAL NOT NULL, in_stock INTEGER NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	manager, err := NewManagerWithDB[documentedProduct]("documented_products", database)
	if err != nil {
		t.Fatal(err)
	}
	return manager
}

func TestDocumentedManagerCreateGetUpdateDelete(t *testing.T) {
	ProductManager := documentedProductManager(t)
	ctx := context.Background()
	product := &documentedProduct{Name: "Mechanical Keyboard", Price: 129.99, InStock: true}
	if err := ProductManager.Create(ctx, product); err != nil {
		t.Fatal(err)
	}
	if product.ID == 0 {
		t.Fatal("Create did not assign an ID")
	}
	loaded, err := ProductManager.Get(ctx, product.ID)
	if err != nil || loaded.Name != product.Name {
		t.Fatalf("Get = %+v, %v", loaded, err)
	}
	product.Price = 119.99
	if err := ProductManager.Update(ctx, product); err != nil {
		t.Fatal(err)
	}
	loaded, err = ProductManager.Get(ctx, product.ID)
	if err != nil || loaded.Price != 119.99 {
		t.Fatalf("updated Get = %+v, %v", loaded, err)
	}
	if err := ProductManager.Delete(ctx, product); err != nil {
		t.Fatal(err)
	}
	count, err := ProductManager.Count(ctx)
	if err != nil || count != 0 {
		t.Fatalf("after Delete count = %d, %v", count, err)
	}
}

func TestDocumentedQuerySetBranchesExecuteIndependently(t *testing.T) {
	ProductManager := documentedProductManager(t)
	ctx := context.Background()
	for _, product := range []*documentedProduct{
		{Name: "budget", Price: 20, InStock: true},
		{Name: "premium", Price: 600, InStock: true},
		{Name: "retired", Price: 5, InStock: false},
	} {
		if err := ProductManager.Create(ctx, product); err != nil {
			t.Fatal(err)
		}
	}
	// Adapt the documented ProductExpr fields to the public typed field constructor.
	price := NewField[float64]("price", "documented_products")
	inStock := NewField[bool]("in_stock", "documented_products")
	activeProducts := ProductManager.QuerySet().Filter(inStock.Eq(true))
	budgetItems, err := activeProducts.Filter(price.Lte(25)).OrderBy(Asc("price")).Limit(10).All(ctx)
	if err != nil || len(budgetItems) != 1 || budgetItems[0].Name != "budget" {
		t.Fatalf("budget branch = %+v, %v", budgetItems, err)
	}
	premiumItems, err := activeProducts.Filter(price.Gte(500)).All(ctx)
	if err != nil || len(premiumItems) != 1 || premiumItems[0].Name != "premium" {
		t.Fatalf("premium branch = %+v, %v", premiumItems, err)
	}
	count, err := activeProducts.Count(ctx)
	if err != nil || count != 2 {
		t.Fatalf("base queryset changed: count = %d, %v", count, err)
	}
	name := NewField[string]("name", "documented_products")
	remaining, err := activeProducts.Exclude(name.Eq("budget")).OrderBy("-price").All(ctx)
	if err != nil || len(remaining) != 1 || remaining[0].Name != "premium" {
		t.Fatalf("excluded ordered branch = %+v, %v", remaining, err)
	}
}

func TestDocumentedDeleteHooksSurroundManagerDelete(t *testing.T) {
	ProductManager := documentedProductManager(t)
	ctx := context.Background()
	var events []string
	product := &documentedProduct{Name: "hooked", Price: 1, InStock: true, events: &events, db: ProductManager.db.db}
	if err := ProductManager.Create(ctx, product); err != nil {
		t.Fatal(err)
	}
	if err := ProductManager.Delete(ctx, product); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(events, []string{"before", "after"}) {
		t.Fatalf("delete hook events = %v", events)
	}
}
