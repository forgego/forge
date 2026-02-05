---
sidebar_position: 3
---

# Custom Fields

Create custom field types for specialized data.

## Field Interface

All fields implement the `Field` interface:

```go
type Field interface {
    Name() string
    Type() string
    SQLType() string
    Validate(value interface{}) error
    ToSQL(value interface{}) (string, []interface{}, error)
}
```

## Creating a Custom Field

### Basic Custom Field

```go
package fields

import (
    "github.com/forgego/forge/pkg/schema"
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
    return &IPAddressFieldImpl{
        name:     f.name,
        required: f.required,
    }
}

type IPAddressFieldImpl struct {
    name     string
    required bool
}

func (f *IPAddressFieldImpl) Name() string {
    return f.name
}

func (f *IPAddressFieldImpl) Type() string {
    return "ipaddress"
}

func (f *IPAddressFieldImpl) SQLType() string {
    return "INET"
}

func (f *IPAddressFieldImpl) Validate(value interface{}) error {
    if f.required && value == nil {
        return errors.New("IP address is required")
    }
    
    if value != nil {
        ip := value.(string)
        if net.ParseIP(ip) == nil {
            return errors.New("invalid IP address")
        }
    }
    
    return nil
}

func (f *IPAddressFieldImpl) ToSQL(value interface{}) (string, []interface{}, error) {
    if value == nil {
        return "NULL", nil, nil
    }
    return "$1", []interface{}{value}, nil
}
```

### Using Custom Field

```go
func (Server) Fields() []schema.Field {
    return []schema.Field{
        schema.Int64("id").Primary().AutoIncrement().Build(),
        schema.String("name").Required().Build(),
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
    return &JSONSchemaFieldImpl{
        name:     f.name,
        required: f.required,
        schema:   f.schema,
    }
}

type JSONSchemaFieldImpl struct {
    name     string
    required bool
    schema   map[string]interface{}
}

func (f *JSONSchemaFieldImpl) Validate(value interface{}) error {
    if f.required && value == nil {
        return errors.New("field is required")
    }
    
    if value != nil {
        // Validate against JSON schema
        data, _ := json.Marshal(value)
        if err := validateJSONSchema(data, f.schema); err != nil {
            return err
        }
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

