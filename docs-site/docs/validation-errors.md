---
sidebar_position: 25
description: Validation, error handling, and problem details.
image: /social-card.png
---

# Validation & Errors

Validation and errors are handled consistently across schema, API, and server layers.

## What you can do

- Field and schema validation with typed validators
- API errors with problem details
- Idempotency key hashing and response stores for applications implementing safe retries (Forge does not automatically replay HTTP requests)
- Sanitization and request ID support

## Validator example

```go
v := validate.NewValidator()
err := v.ValidateField("forge", "min=3")
```

## Error codes

Use the error code registry for consistent API responses.

## Next steps

- [API Errors](/docs/api/errors/)
- [Config Errors](/docs/config/errors/)
