---
sidebar_position: 4
description: API Reference for model fields, options, and constraints.
---

# Fields API Reference

The `github.com/forgego/forge/schema` package provides constructor functions and options for defining typed model fields.

---

## Field Types & Constructors

Every field type can be constructed via fluent builder constructors or functional options:

### Numeric Fields
- `schema.Int64("id")` / `schema.Int64Field("id", ...)`: 64-bit integer (`BIGINT`).
- `schema.Int32("age")` / `schema.Int32Field("age", ...)`: 32-bit integer (`INTEGER`).
- `schema.Float64("score")` / `schema.Float64Field("score", ...)`: Double precision float.
- `schema.Float32("ratio")` / `schema.Float32Field("ratio", ...)`: Single precision float.
- `schema.Decimal("amount", 12, 2)` / `schema.DecimalField("amount", ...)`: Fixed-point decimal number.

### Text & String Fields
- `schema.String("name")` / `schema.StringField("name", ...)`: Variable length string with optional `MaxLength`.
- `schema.Text("body")` / `schema.TextField("body", ...)`: Unbounded long text (`TEXT`).
- `schema.Email("email")` / `schema.EmailField("email", ...)`: Email address with automatic format validation.
- `schema.URL("website")` / `schema.URLField("website", ...)`: Web URL with URL format validation.
- `schema.UUID("identifier")` / `schema.UUIDField("identifier", ...)`: Universally unique identifier.

### Temporal Fields
- `schema.Time("clock")` / `schema.TimeField("clock", ...)`: Time without timezone.
- `schema.Date("day")` / `schema.DateField("day", ...)`: Calendar date without time.
- `schema.DateTime("timestamp")` / `schema.DateTimeField("timestamp", ...)`: Full timestamp with timezone.

### Boolean & Binary Fields
- `schema.Bool("is_active")` / `schema.BoolField("is_active", ...)`: Boolean flag (`BOOLEAN` or `INTEGER 0/1`).
- `schema.JSON("payload")` / `schema.JSONField("payload", ...)`: Structured JSON/JSONB document.
- `schema.Bytes("raw_data")` / `schema.BytesField("raw_data", ...)`: Binary payload (`BYTEA` / `BLOB`).

---

## Field Configuration Options

### Database Layout & Constraints
- `Primary()`: Marks the field as the primary key.
- `AutoIncrement()`: Enables database sequence / auto-increment.
- `Required()`: Adds a `NOT NULL` constraint.
- `Unique()`: Adds a `UNIQUE` constraint and index.
- `DBDefault(expr)`: Sets a database-level default SQL expression (e.g. `DBDefault("NOW()")` or `DBDefault("'active'")`).
- `DBColumn(name)`: Explicit SQL column name override.
- `DBType(sqlType)`: Explicit SQL column type override.
- `DBIndex()`: Instructs migrations to create a B-Tree index.
- `DBCollation(name)`: Specifies character collation.
- `DBComment(text)`: Adds a database comment to the column.

### Generated Columns
- `GeneratedColumn(expression, isStored)`: Defines a computed database column.
  ```go
  // PostgreSQL STORED generated column
  schema.GeneratedColumn("price * (1 + tax_rate)", true)
  ```

### Validation Options
- `MaxLength(n int)`: Maximum string or array length.
- `MinLength(n int)`: Minimum string or array length.
- `MaxValue(v float64)`: Upper numerical bound.
- `MinValue(v float64)`: Lower numerical bound.
- `MaxDigits(n int)` / `DecimalPlaces(n int)`: Precision and scale for decimal numbers.
- `ChoicesOpts(choices ...Choice)`: Permitted enumerated choices.

### Temporal Helpers
- `AutoNowAdd()`: Sets timestamp automatically on record creation (like `created_at`).
- `AutoNow()`: Sets timestamp automatically on every record save (like `updated_at`).

### UI & Admin Visibility
- `VerboseName(label string)`: Human-friendly label displayed in the Admin UI.
- `HelpText(text string)`: Tooltip/help text displayed below input in the Admin UI.
- `Editable(bool)`: If `false`, disables field editing in the Admin UI and serializers.
- `Serialize(bool)`: If `false`, marks field as write-only (e.g. passwords).

