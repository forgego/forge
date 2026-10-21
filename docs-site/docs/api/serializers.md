---
sidebar_position: 2
description: ModelSerializers, field whitelisting, input validation, and nested relational serialization.
image: /social-card.png
---

# ModelSerializers & Validation

Serializers control the representation of model data sent over HTTP and validate incoming request payloads before models are saved.

---

## Defining a ModelSerializer

```go
package catalog

import (
    "errors"
    "github.com/forgego/forge/api"
)

type ProductSerializer struct {
    *api.BaseSerializer
}

func NewProductSerializer() api.Serializer {
    return &ProductSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

// Fields controls which attributes are exposed in JSON
func (ProductSerializer) Fields() []string {
    return []string{
        "id",
        "sku",
        "name",
        "price",
        "tax_rate",
        "price_with_tax",
        "in_stock",
        "created_at",
    }
}

// ReadonlyFields prevents clients from overriding generated or audit fields
func (ProductSerializer) ReadonlyFields() []string {
    return []string{"id", "price_with_tax", "created_at"}
}

// Custom validation runs when Validate is called by a viewset.
func (s *ProductSerializer) Validate() error {
    price, _ := s.Get("price").(float64)
    if price < 0.01 {
        s.AddError("price", "price must be at least $0.01")
        return errors.New("validation failed")
    }
    return s.BaseSerializer.Validate()
}
```

Forge calls the serializer's `Validate()` method. Methods named `ValidatePrice` are not discovered automatically. For model-level checks, implement validation on the model or define schema field constraints.

---

## Serializer Configuration Options

| Method | Return Type | Description |
| :--- | :--- | :--- |
| `Fields()` | `[]string` | Whitelist of field names to serialize. |
| `Exclude()` | `[]string` | Blacklist of field names to omit. |
| `ReadonlyFields()` | `[]string` | Fields serialized in responses but ignored in `POST`/`PUT`/`PATCH`. |
| `WriteOnlyFields()`| `[]string` | Fields accepted in requests but never included in responses (e.g. passwords). |

---

## Handling Validation Errors

When a client submits invalid JSON or fails validation constraints, Forge automatically formats an RFC 7807 compliant error payload:

```json
{
  "status": 422,
  "error": "Unprocessable Entity",
  "detail": "Validation failed for 2 fields",
  "errors": {
    "price": ["price must be at least $0.01"],
    "tax_rate": ["tax rate must be between 0.0 and 1.0"]
  }
}
```

---

## Nested Relational Serialization

The response is built from the model's JSON-tagged fields. A populated nested model field is serialized as an object; embedding another serializer in the serializer struct does not add or fetch a relation. Load the relation explicitly (for example, with `SelectRelated`) before serializing it:

```go
type OrderSerializer struct {
    *api.BaseSerializer
}

type Order struct {
    Customer *Customer `json:"customer"`
}
```

---

## Next Steps

- **[ViewSets](/docs/api/viewsets/)**: Hook serializers into HTTP endpoints.
- **[Pagination](/docs/api/pagination/)**: Paginate list responses.
- **[OpenAPI](/docs/api/openapi/)**: What the OpenAPI document contains today.

