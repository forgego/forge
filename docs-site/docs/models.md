---
sidebar_position: 10
description: Define type-safe models, fields, constraints, generated columns, relations, and lifecycle hooks.
image: /social-card.png
---

# Models & Schema DSL

Models are the cornerstone of a Forge application. By implementing the `schema.Schema` interface, you define fields, validation rules, database options, relations, and lifecycle hooks in pure Go.

From these schema definitions, Forge automatically generates:
- Strongly typed model structs
- Compile-time checked QuerySet expression trees (`orm.Q`)
- Deterministic database migrations
- Instant React 19 Admin UI configuration
- REST API serializers and ViewSets (model generation does not emit OpenAPI paths or schemas)

---

## The Model Contract

Any model struct embeds `schema.BaseSchema` and implements the following methods. This example uses functional constructors supported by Go and generation:

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
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("sku", schema.MaxLength(64), schema.Unique()),
        schema.StringField("name", schema.MaxLength(255), schema.Required()),
        schema.DecimalField("price", schema.MaxDigits(10), schema.DecimalPlaces(2), schema.Required()),
        schema.Float64Field("tax_rate", schema.DBDefault("0.10")),
        schema.BoolField("in_stock", schema.DBDefault("true")),
        schema.Int64Field("category_id", schema.Optional()),
        schema.TimeField("created_at", schema.AutoNowAdd()),
    }
}

func (Product) Relations() []schema.Relation {
    return []schema.Relation{
        schema.ForeignKeyField("category_id", "Category", schema.OnDelete(schema.CascadeSET_NULL), schema.RelatedName("products")),
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

## Functional Field Options

Use `schema.*Field(name, options...)` constructors. The builder-chain examples once shown here do not compile against the current `schema` package.

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

| Field Type | Go field | SQL Type (PostgreSQL / SQLite) | Constructor |
| :--- | :--- | :--- | :--- |
| `Int64` | `int64` | `BIGINT` / `INTEGER` | `schema.Int64Field(name, options...)` |
| `Int32` | `int32` | `INTEGER` | `schema.Int32Field(name, options...)` |
| `String` | `string` | `VARCHAR(n)` with `MaxLength(n)`, otherwise `TEXT` | `schema.StringField(name, options...)` |
| `Text` | `string` | `TEXT` | `schema.TextField(name, options...)` |
| `Bool` | `bool` | `BOOLEAN` / `INTEGER` | `schema.BoolField(name, options...)` |
| `Time` / `DateTime` | `time.Time` | `TIMESTAMP` | `schema.TimeField(name, options...)` / `schema.DateTimeField(name, options...)` |
| `Date` | `time.Time` | `DATE` | `schema.DateField(name, options...)` |
| `Float64` / `Float32` | `float64` / `float32` | `DOUBLE PRECISION` / `REAL` | `schema.Float64Field(name, options...)` |
| `Decimal` | `float64` | `NUMERIC(p, s)` | `schema.DecimalField(name, options...)` |
| `Email` | `string` (validated email format) | `TEXT` by default | `schema.EmailField(name, options...)` |
| `URL` | `string` (validated URL format) | `TEXT` by default | `schema.URLField(name, options...)` |
| `UUID` | `string` | `TEXT` | `schema.UUIDField(name, options...)` |
| `JSON` | `[]byte` | `JSONB` / `TEXT` | `schema.JSONField(name, options...)` |
| `Bytes` | `[]byte` | `BYTEA` / `BLOB` | `schema.BytesField(name, options...)` |

---

## Generated Columns

`schema.GeneratedColumn(expr, stored)` is a field option, but source generation does not yet carry it into migrations. Define the column in a manual migration if you need it:

```go
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
    schema.DBIndex(),                      // Metadata only; use Meta.Indexes for a migration index
)
```

### Choices & Enums

`ChoicesOpts` attaches human-friendly values at runtime. Source generation currently reports it as an unsupported helper, so omit it from models used by `forge generate`:

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

### Defaults, zero values and timestamps

Forge follows Django: `Create` writes the value each field of the struct
holds, and a `Default` is applied when an instance is *constructed*, not when
it is inserted. A Go struct cannot tell an unset `bool` from `false`, so an
explicit `false`, `0` or `""` is stored as is, even when the field declares
`Default(true)`, `Default(5)` or `Default("draft")`.

To start from the declared defaults, build the instance with the manager and
then set your values:

```go
post, err := PostObjects.New() // applies every schema Default
post.Title = "Hello"
err = PostObjects.Create(ctx, post) // status is "draft", published is false
```

The REST API and the admin do the same: a key missing from a create request
gets the field's `Default`, and a key sent as `false`, `0` or `""` is
written. The admin create form also starts each field at its `Default`.

`Create` leaves a column out of the `INSERT`, so the database fills it, only
when:

| Column | Omitted when |
| :--- | :--- |
| Auto-increment primary key, generated column | always |
| `AutoNow` / `AutoNowAdd` timestamp | the Go value is zero (the column defaults to the current time) |
| Field with a `DBDefault` | the Go value is zero; use a pointer type such as `*bool` to store an explicit zero |
| Pointer, slice, map or interface | it is `nil` (`NULL`, or the column default) |
| Struct value such as `time.Time` | it is the zero value |
| Foreign key column | it is `0` or `""` (`NULL` instead of a reference to row 0) |
| Unique, optional field | it is the zero value (`NULL`, so two blank rows do not collide) |
| Optional field without a `Default` | it is the zero value (`NULL`, as the column is nullable and has no `DEFAULT`) |
| UUID, JSON, decimal, date or time field, or one with a non-text custom `DBType` | it is `""`, which is not a valid value for the column |

Apart from the first three rows, a required field is always written, even when zero.
So the explicit `false`, `0` or `""` that `Create` writes is the value of a
required field or of a field with a `Default`, where omitting it would let the
column default replace it. Use a pointer field such as `*string` to store an
explicit `""` in an optional field without a `Default`.

`BulkCreate` builds one `INSERT` for all rows, so every row must write the
same columns: rows where one has a value and another leaves the same optional
field at its zero value fail with "bulk insert requires consistent columns".
Give such fields a pointer type or a `Default`, or insert those rows
separately.

`AutoNow` fields are set to the current time on every `Update`, `Save` and
`UpdateFields`, both in the database and on the struct (`time.Time`,
`*time.Time` or `string` fields). `AutoNowAdd` fields are set once, by the
database, on insert; an update never clears them. Like Django's
`QuerySet.update()`, `QuerySet.Update` and `BulkUpdate` write only the
columns you name.

---

## Relationships & Cascade Behaviors

Forge generates foreign-key constraints for `ForeignKeyField` and `OneToOneField`. `ManyToManyField` describes a relation, but migration generation does not create its through table; create that table manually:

```go
func (Order) Relations() []schema.Relation {
    return []schema.Relation{
        // Many-to-One: Foreign Key with cascade protection
        schema.ForeignKeyField("customer_id", "Customer", schema.OnDelete(schema.CascadePROTECT), schema.RelatedName("orders")),

        // One-to-One: Profile linked to User with cascade deletion
        schema.OneToOneField("profile_id", "CustomerProfile", schema.OnDelete(schema.CascadeCASCADE), schema.RelatedName("customer")),

        // Many-to-Many: Order items linked via through table
        schema.ManyToManyField("tags", "Tag", schema.Through("order_tags"), schema.RelatedName("orders")),
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

### Supported schema DSL

Two fixture packages in the repository define the supported subset, and tests check both:

- [`codegen/testdata/dsl/supported/models.go`](https://github.com/forgego/forge/blob/master/codegen/testdata/dsl/supported/models.go)
  covers every form listed below. Generating it twice gives byte-identical `gen.go` and
  `api_gen.go`, the generated package passes `go build` and `go vet`, and its
  `makemigrations` output is pinned column by column.
- [`codegen/testdata/dsl/unsupported/models.go`](https://github.com/forgego/forge/blob/master/codegen/testdata/dsl/unsupported/models.go)
  holds a helper call, a helper option, a loop, a conditional, a computed `append`, and a
  computed `Meta` value. Each one produces exactly one diagnostic at its own line.

Supported forms:

- **Fields:** `Int64Field`, `Int32Field`, `StringField`, `TextField`, `EmailField`,
  `URLField`, `UUIDField`, `BoolField`, `Float64Field`, `Float32Field`, `DecimalField`,
  `JSONField`, `BytesField`, `DateField`, `TimeField`, and `DateTimeField`. Each takes
  literal options: `Primary`, `AutoIncrement`, `Required`, `Optional`, `Unique`,
  `Default`, `MaxLength`, `MinLength`, `MinValue`, `MaxValue`, `MaxDigits`,
  `DecimalPlaces`, `HelpText`, `VerboseName`, `AutoNow`, and `AutoNowAdd`.
- **Relations:** `ForeignKeyField` and `OneToOneField` with `OnDelete(schema.Cascade...)`
  and `RelatedName`. Also declare the key column (for example `author_id`) in `Fields()`,
  or no foreign key is created.
- **Meta:** literal `TableName`, `VerboseName`, `VerboseNamePlural`, `OrderBy`, `Indexes`
  as `schema.Index{...}` literals or `schema.IndexOn` / `schema.UniqueIndexOn` with
  literal name/field arguments, and `Constraints` as `schema.Constraint{...}` literals,
  either `CHECK` with a `Condition` or `UNIQUE` with `Fields`.
- **Hooks:** `schema.NewModelHooks().With...(...)`.

Generation reads the following forms without warning, but `makemigrations` does not
apply them yet: `DBIndex()` (use `Meta.Indexes`), `UniqueTogether` (use a `UNIQUE`
constraint), and `ManyToManyField` (no join table is created). `MaxLength` controls
`VARCHAR(n)` for a `StringField` column. `schema.ChoicesOpts(...)` is still an
unsupported helper call and produces a diagnostic. Arbitrary metadata expressions
also produce diagnostics; automatic migration generation refuses any reported
unsupported expression rather than diffing a partial schema.

Functional constructors map to their declared schema type. Existing databases created
with the older Go-type fallback need an explicit migration before relying on the new types.

### Strict mode

`forge generate --strict` parses the models before writing anything. If any expression
cannot be evaluated, it prints each one as `file:line:column: Model.Method: message`, exits
with an error, and leaves the existing `gen.go` untouched. Use `--strict` in CI so a model
that generation would silently misread fails the build.

### Regeneration and upgrades

- `gen.go` and `api_gen.go` start with `// Code generated by forge. DO NOT EDIT.` and are
  rewritten in full each time. Generation leaves every other file in the directory alone,
  so keep your own code in separate files.
- Identical models produce identical output, so a diff in a generated file comes from the
  models or from a Forge upgrade. After upgrading Forge, run `forge generate` and
  `forge makemigrations <name> --auto`, then review both diffs before you commit.
- Running `makemigrations` on unchanged models writes no migration. When an upgrade makes
  Forge emit DDL that an older release skipped, the first `makemigrations` afterwards adds
  it. For example, foreign keys declared with `ForeignKeyField` and table constraints were
  not emitted before, so that migration can fail on existing rows that violate them. Review
  and test it like any other migration (see [safe changes](/docs/migrations/#safe-schema-changes)).

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
  `GetID`/`SetID` methods declared on or promoted into the model. Pointer-embedded helpers
  are not detected.

Generation stops with an error naming the model when either is missing.

Generated endpoints follow the schema:

- Fields marked `Serialize(false)` never appear in responses under any of their names (schema name, column, or JSON tag), and cannot be used to filter, order, or search.
- Non-editable fields are ignored in create and update bodies.
- A list request without `ordering` uses the model's `Meta().OrderBy`, or the primary key when
  none is set. Ordering fields that are hidden (`Serialize(false)`) are dropped, so a
  `Meta().OrderBy` built only from hidden fields falls back to the primary key.
- The request body must be a JSON object; `null` or any other value returns 400.
- `Time` and `DateTime` fields are returned as RFC 3339 timestamps (`2006-01-02T15:04:05Z07:00`) and `Date` fields as `2006-01-02`, the layouts requests accept.

Both files are rendered and staged before either is replaced, and the previous contents are
backed up first. If replacing `api_gen.go` fails, the generator restores `gen.go` on a
best-effort basis. A failure during that restore, or a crash between the two renames, can
still leave one file new and the other old; rerun the generator after such a failure.

---

## Next Steps

- **[Type-Safe ORM & QuerySet](/docs/orm/)**: Learn how to query, filter with `orm.Q`, and aggregate data.
- **[AST Migrations](/docs/migrations/)**: Automatically generate and apply schema migrations.
- **[Admin Console](/docs/admin/overview/)**: Expose your models with zero frontend code.
