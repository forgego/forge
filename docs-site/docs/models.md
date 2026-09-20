---
sidebar_position: 10
description: Define type-safe models, fields, constraints, generated columns, relations, and lifecycle hooks.
image: /forge-social-card.svg
---

# Models & Schema DSL

Models are the cornerstone of a Forge application. By implementing the `schema.Schema` interface, you define fields, validation rules, database options, relations, and lifecycle hooks in pure Go.

From these schema definitions, Forge automatically generates:
- Strongly typed model structs
- Compile-time checked QuerySet expression trees (`orm.Q`)
- Deterministic database migrations
- Instant React 19 Admin UI configuration
- REST API serializers and OpenAPI 3.0 documentation

---

## The Model Contract

Any model struct embeds `schema.BaseSchema` and implements the following methods:

```go
package models

import (
    "context"
    "github.com/forgego/forge/schema"
)

type Product struct {
    schema.BaseSchema
}

func (Product) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement(),
        schema.String("sku").MaxLength(64).Unique().DBIndex(),
        schema.String("name").MaxLength(255).Required(),
        schema.Decimal("price", 10, 2).Required(),
        schema.Float64("tax_rate").DBDefault("0.10"),
        schema.GeneratedColumn("price_with_tax", "price * (1 + tax_rate)", true),
        schema.Bool("in_stock").DBDefault("true"),
    }
}

func (Product) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKey("category_id", "Category", schema.CascadeSET_NULL).
            RelatedName("products"),
    }
}

func (Product) Meta() schema.Meta {
    return schema.Meta{
        TableName: "store_products",
        OrderBy:   []string{"-created_at"},
    }
}

func (Product) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithBeforeSave(func(ctx context.Context, instance interface{}) error {
            // Run pre-save calculations or validations
            return nil
        })
}
```

---

## Dual Syntax: Builder & Functional Options

Forge provides two equivalent, idiomatic styles for defining fields. Use whichever best fits your team's style guide:

### 1. Fluent Builder Style (Recommended)

```go
schema.String("email").MaxLength(255).Required().Unique().DBIndex()
schema.Decimal("balance", 12, 2).Required().MinValue(0.0)
schema.Time("created_at").AutoNowAdd()
```

### 2. Functional Options Style

```go
schema.StringField("email",
    schema.MaxLength(255),
    schema.Required(),
    schema.Unique(),
    schema.DBIndex(),
)

schema.DecimalField("balance",
    schema.MaxDigits(12),
    schema.DecimalPlaces(2),
    schema.Required(),
    schema.MinValue(0.0),
)

schema.TimeField("created_at", schema.AutoNowAdd())
```

---

## Supported Field Types

Forge includes first-class support for all common relational and modern data types:

| Field Type | Go Under-the-Hood | SQL Type (PostgreSQL / SQLite) | Builder Constructor |
| :--- | :--- | :--- | :--- |
| `Int64` | `int64` | `BIGINT` / `INTEGER` | `schema.Int64(name)` |
| `Int32` | `int32` | `INTEGER` | `schema.Int32(name)` |
| `String` | `string` | `VARCHAR(n)` | `schema.String(name)` |
| `Text` | `string` | `TEXT` | `schema.Text(name)` |
| `Bool` | `bool` | `BOOLEAN` / `INTEGER` | `schema.Bool(name)` |
| `Time` / `DateTime` | `time.Time` | `TIMESTAMP WITH TIME ZONE` | `schema.Time(name)` / `schema.DateTime(name)` |
| `Date` | `time.Time` | `DATE` | `schema.Date(name)` |
| `Float64` / `Float32` | `float64` / `float32` | `DOUBLE PRECISION` / `REAL` | `schema.Float64(name)` |
| `Decimal` | `string` / `shopspring.Decimal` | `NUMERIC(p, s)` / `DECIMAL` | `schema.Decimal(name, digits, places)` |
| `Email` | `string` (validated email format) | `VARCHAR(254)` | `schema.Email(name)` |
| `URL` | `string` (validated URL format) | `VARCHAR(2048)` | `schema.URL(name)` |
| `UUID` | `string` / `uuid.UUID` | `UUID` / `VARCHAR(36)` | `schema.UUID(name)` |
| `JSON` | `[]byte` / `any` | `JSONB` / `JSON` / `TEXT` | `schema.JSON(name)` |
| `Bytes` | `[]byte` | `BYTEA` / `BLOB` | `schema.Bytes(name)` |

---

## Generated Columns

Forge natively supports **Database Generated Columns** (stored or virtual computed columns). The database computes and indexes values automatically:

```go
// Stored generated column (PostgreSQL STORED, SQLite GENERATED ALWAYS AS ... STORED)
schema.GeneratedColumn("price_with_tax", "price * (1 + tax_rate)", true)

// Functional syntax:
schema.DecimalField("price_with_tax",
    schema.GeneratedColumn("price * (1 + tax_rate)", true),
)
```

---

## Database Constraints & Options

Forge allows fine-grained control over database physical layout, indexes, defaults, and collations:

```go
schema.StringField("title",
    schema.Required(),
    schema.DBColumn("article_title"),      // Custom column name
    schema.DBDefault("'Untitled'"),        // Database-level SQL default expression
    schema.DBCollation("en_US.utf8"),       // Collation for collation-sensitive sorting
    schema.DBComment("Public headline"),   // Table column comment
    schema.DBIndex(),                      // Create B-Tree index
)
```

### Choices & Enums

Add human-friendly enumerated values:

```go
schema.StringField("status",
    schema.ChoicesOpts(
        schema.NewChoice("draft", "Draft Order"),
        schema.NewChoice("processing", "In Processing"),
        schema.NewChoice("shipped", "Shipped & In Transit"),
        schema.NewChoice("delivered", "Delivered"),
    ),
    schema.DBDefault("'draft'"),
)
```

---

## Relationships & Cascade Behaviors

Forge handles `ForeignKey`, `OneToOne`, and `ManyToMany` with configurable referential actions:

```go
func (Order) Relations() []schema.Relation {
    return []schema.Relation{
        // Many-to-One: Foreign Key with cascade protection
        schema.ForeignKey("customer_id", "Customer", schema.CascadePROTECT).
            RelatedName("orders"),

        // One-to-One: Profile linked to User with cascade deletion
        schema.OneToOne("profile_id", "CustomerProfile", schema.CascadeCASCADE).
            RelatedName("customer"),

        // Many-to-Many: Order items linked via through table
        schema.ManyToMany("tags", "Tag").
            Through("order_tags").
            RelatedName("orders"),
    }
}
```

### Supported Cascade Actions
- `schema.CascadeCASCADE`: Automatically delete child rows when the referenced parent is deleted.
- `schema.CascadePROTECT`: Prevent deletion of the parent if any child row references it (raises integrity error).
- `schema.CascadeSET_NULL`: Set the foreign key column to `NULL` when the referenced parent is deleted.
- `schema.CascadeSET_DEFAULT`: Set the foreign key column to its SQL default value.
- `schema.CascadeDO_NOTHING`: Take no action at database level.

---

## Lifecycle Hooks

Execute business logic, audit recording, or validation during ORM lifecycle events:

```go
func (Order) Hooks() *schema.ModelHooks {
    return schema.NewModelHooks().
        WithBeforeCreate(func(ctx context.Context, instance interface{}) error {
            order := instance.(*Order)
            // Assign sequential invoice number
            return nil
        }).
        WithBeforeSave(func(ctx context.Context, instance interface{}) error {
            order := instance.(*Order)
            // Compute total amount
            order.Total = order.Subtotal + order.TaxAmount + order.ShippingAmount
            return nil
        }).
        WithAfterCreate(func(ctx context.Context, instance interface{}) error {
            // Trigger asynchronous transactional email
            return nil
        })
}
```

Available hooks:
- `BeforeCreate` / `AfterCreate`
- `BeforeUpdate` / `AfterUpdate`
- `BeforeSave` / `AfterSave`
- `BeforeDelete` / `AfterDelete`
- `Clean` (for model-level multi-field validation)

---

## Code Generation

After defining models, run:

```bash
forge generate
```

This compiles your schema into high-performance Go types with zero runtime reflection overhead in your query execution paths.

### What generation can read

Generation reads direct schema constructor calls in a returned slice, a `var` slice literal, an assignment, or an individual `append`. It cannot evaluate helper calls, values computed by functions, loops, conditional (`if` or `switch`) assembly, or appending computed slices. These constructs are reported as warnings; use `forge generate --strict` to turn any warning into an error.

### Generating a REST API

```bash
forge generate --api
```

With `--api`, generation also writes `api_gen.go` next to `gen.go`. For each model it contains a serializer, a ViewSet, and a `Register<Model>Routes` function, plus `RegisterAPIRoutes` for the whole package.

Nothing is served until you register the routes yourself, during server setup:

```go
import blog "myapp/app/blog"

// Generated managers are created without a database connection, so bind them first.
blog.PostObjects.SetDB(database)
blog.CategoryObjects.SetDB(database)

blog.RegisterAPIRoutes(router) // serves /api/v1/posts/, /api/v1/categories/, ...
```

Each model is served under `/api/v1/<kebab-case plural of the model name>/`. The collection
routes are registered with a trailing slash, so the list and create URL is `/api/v1/posts/`,
and the slashless `/api/v1/posts` returns 404. Detail routes are `/api/v1/posts/{id}`, without
a trailing slash. Use the exact forms above in clients.

:::warning Secure the generated endpoints before registering them
Generated ViewSets declare no authentication or permission classes, and
`api.DefaultSettings()` starts with both lists empty, which permits anonymous
requests. Registering them as-is exposes unauthenticated list, create, update, and
delete for every model. Call `api.SetDefaultAuthentication(...)` and
`api.SetDefaultPermissions(...)` at startup, or set `Authentication` and `Permissions`
on each ViewSet, before mounting the routes.
:::

Running without `--api` does not undo anything: a previously generated `api_gen.go` is
left untouched, so it keeps compiling and keeps serving wherever it is registered.
Delete the file to remove the generated API.

Each model with API generation needs:

- exactly one primary key, named `id`, declared as an auto-increment `int64` field in `Fields()`;
- a writable `int64` ID: a declared `ID`/`Id` field, an embedded `<Model>Generated`, or
  `GetID`/`SetID` methods declared on the model itself. Methods promoted from another
  embedded helper type are not detected today, even though they satisfy `orm.ModelWithID`.

Generation stops with an error naming the model when either is missing.

Generated endpoints follow the schema:

- Fields marked `Serialize(false)` never appear in responses under any of their names (schema name, column, or JSON tag), and cannot be used to filter, order, or search.
- Non-editable fields are ignored in create and update bodies.
- A list request without `ordering` uses the model's `Meta().OrderBy`, or the primary key when
  none is set. Ordering fields that are hidden (`Serialize(false)`) are dropped, so a
  `Meta().OrderBy` built only from hidden fields currently leaves the list unordered, and paging
  through it can repeat or skip rows; keep at least one visible field in `Meta().OrderBy`.
- The request body must be a JSON object; `null` or any other value returns 400.
- `Time` fields are returned as `15:04:05` and `Date` fields as `2006-01-02`, the layouts requests accept.

Both files are rendered and staged before either is replaced, and the previous contents are
backed up first. If replacing `api_gen.go` fails, the generator restores `gen.go` on a
best-effort basis. A failure during that restore, or a crash between the two renames, can
still leave one file new and the other old; rerun the generator after such a failure.

---

## Next Steps

- **[Type-Safe ORM & QuerySet](/docs/orm/)**: Learn how to query, filter with `orm.Q`, and aggregate data.
- **[AST Migrations](/docs/migrations/)**: Automatically generate and apply schema migrations.
- **[Admin Console](/docs/admin/overview/)**: Expose your models with zero frontend code.
