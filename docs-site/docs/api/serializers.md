---
sidebar_position: 2
description: ModelSerializers, field whitelisting, input validation, and nested relational serialization.
image: /forge-social-card.svg
---

# ModelSerializers & Validation

Serializers control the representation of model data sent over HTTP and validate incoming request payloads before models are saved.

---

## Defining a ModelSerializer

```go
package catalog

import (
    "context"
    "errors"
    "github.com/forgego/forge/api"
    "myapp/models"
)

type ProductSerializer struct {
    api.ModelSerializer[models.Product]
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

// Custom field validation: Validate<FieldName>
func (s *ProductSerializer) ValidatePrice(price float64) error {
    if price < 0.01 {
        return errors.New("price must be at least $0.01")
    }
    return nil
}

// Cross-field object validation
func (s *ProductSerializer) Validate(ctx context.Context, p *models.Product) error {
    if p.TaxRate < 0 || p.TaxRate > 1.0 {
        return errors.New("tax rate must be between 0.0 and 1.0")
    }
    return nil
}
```

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

Serialize related entities by embedding their serializers:

```go
type OrderSerializer struct {
    api.ModelSerializer[models.Order]
    Customer   CustomerSummarySerializer   `json:"customer"`
    OrderItems []OrderItemDetailSerializer `json:"items"`
}
```

---

## Next Steps

- **[ViewSets](/docs/api/viewsets/)**: Hook serializers into HTTP endpoints.
- **[Pagination](/docs/api/pagination/)**: Paginate list responses.
- **[OpenAPI 3.0](/docs/api/openapi/)**: Generate schemas from serializers.

