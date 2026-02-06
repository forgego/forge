---
sidebar_position: 25
description: Validation, error handling, and problem details.
image: /forge-social-card.svg
---

# Validation & Errors

Validation and errors are handled consistently across schema, API, and server layers.

## What you can do

- Field and schema validation with typed validators
- API errors with problem details
- Idempotency handling for safe retries
- Sanitization and request ID support

## Validator example

```go
v := validate.NewTypedValidator()
err := v.ValidateString("name", "forge", validate.MinLength(3))
```

## Error codes

Use the error code registry for consistent API responses.

## Next steps

- [API Errors](/docs/api/errors/)
- [Config Errors](/docs/config/errors/)
