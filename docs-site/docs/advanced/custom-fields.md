---
sidebar_position: 3
---

# Custom Fields

Create custom field types for specialized data.

## Field Structure

Custom fields are built on the `schema.Field` struct:

```go
type Field struct {
    Name       string
    Type       FieldType
    DBType     string
    Required   bool
    Validators []Validator
}
```

## Creating a Custom Field

### Basic Custom Field

```go
package fields

import (
    "github.com/forgego/forge/schema"
)

type IPAddressField struct {
    name     string
    required bool
}

func IPAddress(name string) *IPAddressField {
    return &IPAddressField{
        name: name,
    }
}

func (f *IPAddressField) Required() *IPAddressField {
    f.required = true
    return f
}

func (f *IPAddressField) Build() schema.Field {
    opts := []schema.FieldOpt{}
    if f.required {
        opts = append(opts, schema.Required())
    }
    field := schema.StringField(f.name, opts...)
    field.DBType = "INET"
    field.Validators = append(field.Validators, IPAddressValidator{required: f.required})
    return field
}

type IPAddressValidator struct {
    required bool
}

func (v IPAddressValidator) Validate(value interface{}) error {
    if value == nil {
        if v.required {
            return errors.New("IP address is required")
        }
        return nil
    }
    ip, ok := value.(string)
    if !ok {
        return errors.New("expected string")
    }
    if net.ParseIP(ip) == nil {
        return errors.New("invalid IP address")
    }
    return nil
}
```

### Using Custom Field

```go
func (Server) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
        schema.StringField("name", schema.Required()),
        fields.IPAddress("ip_address").Required().Build(),
        fields.IPAddress("gateway").Build(),
    }
}
```

## Advanced Custom Field

### JSON Field with Validation

```go
type JSONSchemaField struct {
    name     string
    required bool
    schema   map[string]interface{}
}

func JSONSchema(name string, schema map[string]interface{}) *JSONSchemaField {
    return &JSONSchemaField{
        name:   name,
        schema: schema,
    }
}

func (f *JSONSchemaField) Required() *JSONSchemaField {
    f.required = true
    return f
}

func (f *JSONSchemaField) Build() schema.Field {
    opts := []schema.FieldOpt{}
    if f.required {
        opts = append(opts, schema.Required())
    }
    field := schema.JSONField(f.name, opts...)
    field.Validators = append(field.Validators, JSONSchemaValidator{
        required: f.required,
        schema:   f.schema,
    })
    return field
}

type JSONSchemaValidator struct {
    required bool
    schema   map[string]interface{}
}

func (v JSONSchemaValidator) Validate(value interface{}) error {
    if value == nil {
        if v.required {
            return errors.New("field is required")
        }
        return nil
    }

    // Validate against JSON schema
    data, _ := json.Marshal(value)
    if err := validateJSONSchema(data, v.schema); err != nil {
        return err
    }

    return nil
}
```

## Field Options

Add options to your custom field:

```go
type IPAddressField struct {
    name     string
    required bool
    version  int // IPv4 or IPv6
}

func (f *IPAddressField) IPv4() *IPAddressField {
    f.version = 4
    return f
}

func (f *IPAddressField) IPv6() *IPAddressField {
    f.version = 6
    return f
}
```

## Best Practices

1. **Validate Input** - Always validate field values
2. **Handle NULL** - Properly handle NULL values
3. **Use Appropriate SQL Types** - Choose the right SQL type
4. **Document Fields** - Document your custom fields
5. **Test Fields** - Write tests for custom fields

## See Also

- [Fields Reference](/docs/api-reference/fields) - Built-in field types
- [Models Guide](/docs/guides/models) - Using fields in models

