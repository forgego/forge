# Models, Meta, and Constraints

## Defining a Model

Embed `<Model>Generated` (generated into `gen.go` by `forge generate`) and implement `Fields()`, `Meta()`, `Relations()`, and `Hooks()` in the **same `.go` file**:

```go
package catalog

import "github.com/forgego/forge/schema"

type Product struct {
	ProductGenerated
}

func (Product) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(200)),
		schema.DecimalField("price", schema.Required(), schema.MaxDigits(10), schema.DecimalPlaces(2)),
		schema.BoolField("is_active", schema.Default(true)),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Product) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "products",
		OrderBy:           []string{"-created_at"},
		VerboseName:       "Product",
		VerboseNamePlural: "Products",
	}
}

func (Product) Relations() []schema.Relation { return nil }
func (Product) Hooks() *schema.ModelHooks  { return nil }
```

## `schema.Meta` Options

```go
func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName:         "users",
		OrderBy:           []string{"-date_joined"},
		VerboseName:       "User",
		VerboseNamePlural: "Users",
		Indexes: []schema.Index{
			{Name: "idx_user_email", Fields: []string{"email"}, Unique: false},
		},
		UniqueTogether: [][]string{
			{"username", "email"},
		},
		Constraints: []schema.Constraint{
			{
				Name:      "check_age",
				Type:      schema.ConstraintTypeCheck,
				Condition: "age >= 0",
			},
		},
	}
}
```

You can also construct check constraints with `schema.Check("check_age", "age >= 0")`.

## Composite Literals with Embedded `<Model>Generated`

Go does not allow promoted fields in struct literals. Initialize embedded fields explicitly:

```go
p := &catalog.Product{
	ProductGenerated: catalog.ProductGenerated{
		Name:     "Desk",
		Price:    199.99,
		IsActive: true,
	},
}
```
