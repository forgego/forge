# Field Constructors and Options

All field definitions in `github.com/forgego/forge/schema` use functional constructors (`schema.<Type>Field(name, opts...)`) defined in `schema/fields_functional.go`.

## Field Constructors

### Numeric
- `schema.Int64Field(name, opts...)` -> Go `int64`
- `schema.Int32Field(name, opts...)` -> Go `int32`
- `schema.Float64Field(name, opts...)` -> Go `float64`
- `schema.Float32Field(name, opts...)` -> Go `float32`
- `schema.DecimalField(name, opts...)` -> Go `float64` (codegen emits `float64`)

### String & Text
- `schema.StringField(name, opts...)` -> Go `string`
- `schema.TextField(name, opts...)` -> Go `string`
- `schema.EmailField(name, opts...)` -> Go `string`
- `schema.URLField(name, opts...)` -> Go `string`

### Boolean, Time & Special
- `schema.BoolField(name, opts...)` -> Go `bool`
- `schema.TimeField(name, opts...)` -> Go `time.Time`
- `schema.DateField(name, opts...)` -> Go `time.Time`
- `schema.DateTimeField(name, opts...)` -> Go `time.Time`
- `schema.UUIDField(name, opts...)` -> Go `string`
- `schema.JSONField(name, opts...)` -> Go `[]byte` / JSON
- `schema.BytesField(name, opts...)` -> Go `[]byte`

## Field Options

```go
schema.StringField(
	"username",
	schema.Required(),
	schema.Unique(),
	schema.DBIndex(),
	schema.DBColumn("user_name"),
	schema.Default("guest"),
	schema.MinLength(3),
	schema.MaxLength(150),
	schema.VerboseName("Username"),
	schema.HelpText("Public handle"),
	schema.ChoicesFromPairsOpts("admin", "Admin", "user", "User"),
	schema.Editable(true),
	schema.Serialize(true),
)
```

Common options:
- **Keys & Constraints**: `schema.Primary()`, `schema.AutoIncrement()`, `schema.Required()`, `schema.Unique()`, `schema.Blank()`
- **Database**: `schema.DBIndex()`, `schema.DBColumn(col)`, `schema.DBType(sqlType)`, `schema.DBDefault(expr)`, `schema.Default(val)`, `schema.GeneratedColumn(expr, stored)`
- **Validation**: `schema.MinLength(n)`, `schema.MaxLength(n)`, `schema.MinValue(v)`, `schema.MaxValue(v)`, `schema.MaxDigits(d)`, `schema.DecimalPlaces(p)`, `schema.ChoicesOpts(choices...)`, `schema.ChoicesFromPairsOpts(pairs...)`
- **Temporal**: `schema.AutoNowAdd()`, `schema.AutoNow()`
- **Admin / API Visibility**: `schema.VerboseName(s)`, `schema.HelpText(s)`, `schema.Editable(bool)`, `schema.Serialize(bool)`

## Go & SQL Type Mapping (`codegen/ast_fields.go`, `db/migrate/sql/common.go`)

| Constructor | Go Type | PostgreSQL | SQLite |
|---|---|---|---|
| `Int64Field` | `int64` | `BIGINT` (`BIGSERIAL` if AutoIncrement) | `INTEGER` |
| `Int32Field` | `int32` | `INTEGER` (`SERIAL` if AutoIncrement) | `INTEGER` |
| `Float64Field` | `float64` | `DOUBLE PRECISION` | `REAL` |
| `Float32Field` | `float32` | `REAL` | `REAL` |
| `DecimalField` | `float64` | `DECIMAL(p, s)` | `REAL` / `NUMERIC` |
| `StringField` | `string` | `VARCHAR(n)` or `TEXT` | `TEXT` |
| `TextField` | `string` | `TEXT` | `TEXT` |
| `EmailField` / `URLField` | `string` | `TEXT` (or `VARCHAR(n)` with `MaxLength`) | `TEXT` |
| `BoolField` | `bool` | `BOOLEAN` | `INTEGER` |
| `TimeField` | `time.Time` | `TIME` | `TEXT` |
| `DateField` | `time.Time` | `DATE` | `TEXT` |
| `DateTimeField` | `time.Time` | `TIMESTAMP` / `TIMESTAMP WITH TIME ZONE` | `TEXT` |
| `UUIDField` | `string` | `UUID` | `CHAR(36)` |
| `JSONField` | `json.RawMessage` / `[]byte` | `JSONB` | `TEXT` |
| `BytesField` | `[]byte` | `BYTEA` | `BLOB` |

Note: `orm.Manager` and `forge generate --api` only support `int64` auto-increment primary keys named `id`.
