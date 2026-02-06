---
sidebar_position: 4
---

# Fields

Field expressions provide type-safe access to model fields. Use them in queries to ensure compile-time type checking.

Complete reference for all field types and options.

## Field Types

### Numeric Fields

#### Int64

64-bit integer:

```go
schema.Int64("id").Primary().AutoIncrement().Build()
schema.Int64("age").Build()
```

#### Int32

32-bit integer:

```go
schema.Int32("count").Build()
```

#### Float64

64-bit floating point:

```go
schema.Float64("price").Build()
```

#### Decimal

Decimal with precision:

```go
schema.Decimal("amount").
    MaxDigits(10).
    DecimalPlaces(2).
    Build()
```

### String Fields

#### String

Variable-length string:

```go
schema.String("username").
    MaxLength(150).
    Required().
    Build()
```

#### Text

Unlimited text:

```go
schema.Text("content").Build()
```

#### Email

Email field with validation:

```go
schema.Email("email").
    Unique().
    Required().
    Build()
```

#### URL

URL field with validation:

```go
schema.URL("website").Build()
```

#### Slug

URL-friendly string:

```go
schema.Slug("slug").
    MaxLength(200).
    Unique().
    Build()
```

### Boolean Fields

#### Bool

Boolean field:

```go
schema.Bool("is_active").
    Default(true).
    Build()
```

### Date and Time Fields

#### Time

Timestamp:

```go
schema.Time("created_at").
    AutoNowAdd().
    Build()
```

#### Date

Date only:

```go
schema.Date("birth_date").Build()
```

#### DateTime

Date and time:

```go
schema.DateTime("last_login").Build()
```

### Special Fields

#### UUID

UUID field:

```go
schema.UUID("id").Primary().Build()
```

#### JSON

JSON field:

```go
schema.JSON("metadata").Build()
```

#### Bytes

Binary data:

```go
schema.Bytes("avatar").Build()
```

## Field Options

All field types support these options:

### Required

Make field required (NOT NULL):

```go
schema.String("username").Required().Build()
```

### Unique

Add unique constraint:

```go
schema.String("email").Unique().Build()
```

### Primary

Set as primary key:

```go
schema.Int64("id").Primary().Build()
```

### Index

Create database index:

```go
schema.String("username").Index().Build()
```

### DBColumn

Custom column name:

```go
schema.String("firstName").DBColumn("first_name").Build()
```

### Default

Default value:

```go
schema.Bool("is_active").Default(true).Build()
schema.String("status").Default("pending").Build()
```

### MaxLength

Maximum length (strings):

```go
schema.String("username").MaxLength(150).Build()
```

### MinLength

Minimum length (strings):

```go
schema.String("password").MinLength(8).Build()
```

### MaxDigits

Maximum digits (decimal):

```go
schema.Decimal("amount").
    MaxDigits(10).
    DecimalPlaces(2).
    Build()
```

### DecimalPlaces

Decimal places (decimal):

```go
schema.Decimal("price").
    MaxDigits(10).
    DecimalPlaces(2).
    Build()
```

### Choices

Predefined choices:

```go
schema.String("status").
    Choices("active", "pending", "deleted").
    Build()
```

### Null

Allow NULL:

```go
schema.String("middle_name").Null().Build()
```

### Blank

Allow blank (strings):

```go
schema.String("description").Blank().Build()
```

### HelpText

Help text for admin:

```go
schema.String("username").
    HelpText("Required. 150 characters or fewer.").
    Build()
```

### VerboseName

Human-readable name:

```go
schema.String("firstName").
    VerboseName("First Name").
    Build()
```

### Editable

Make read-only in admin:

```go
schema.Time("created_at").
    Editable(false).
    Build()
```

### AutoNow

Set on save:

```go
schema.Time("updated_at").AutoNow().Build()
```

### AutoNowAdd

Set on create only:

```go
schema.Time("created_at").AutoNowAdd().Build()
```

### AutoIncrement

Auto-incrementing (primary keys):

```go
schema.Int64("id").
    Primary().
    AutoIncrement().
    Build()
```

## Field Type Reference Table

| Field Type | Go Type | SQL Type | Builder |
|------------|---------|----------|---------|
| Int64 | `int64` | `BIGINT` | `schema.Int64()` |
| Int32 | `int32` | `INTEGER` | `schema.Int32()` |
| String | `string` | `VARCHAR(n)` | `schema.String()` |
| Text | `string` | `TEXT` | `schema.Text()` |
| Bool | `bool` | `BOOLEAN` | `schema.Bool()` |
| Float64 | `float64` | `DOUBLE PRECISION` | `schema.Float64()` |
| Decimal | `decimal.Decimal` | `NUMERIC` | `schema.Decimal()` |
| Time | `time.Time` | `TIMESTAMP` | `schema.Time()` |
| Date | `time.Time` | `DATE` | `schema.Date()` |
| DateTime | `time.Time` | `TIMESTAMP` | `schema.DateTime()` |
| Email | `string` | `VARCHAR(255)` | `schema.Email()` |
| URL | `string` | `VARCHAR(255)` | `schema.URL()` |
| UUID | `string` | `UUID` | `schema.UUID()` |
| JSON | `[]byte` | `JSONB` | `schema.JSON()` |
| Bytes | `[]byte` | `BYTEA` | `schema.Bytes()` |

## See Also

- [Schema Reference](/docs/api-reference/schema) - Model definition
- [Models Guide](/docs/guides/models) - Model usage guide

