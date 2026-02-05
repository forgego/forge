---
sidebar_position: 6
description: Validate data in forge models and APIs.
keywords:
  - forge validation
  - model validation
  - serializer validation
image: /forge-social-card.svg
---

# Validation

Validation in forge happens in two common layers: schema fields and API
serializers. Start with schema constraints, then add API validation where
needed.

## Schema field constraints

Use field options like `Required`, `MaxLength`, and `Default` to enforce data
rules at the model level. See the [Schema System feature](/docs/features/schema-system/)
for a full list of field options.

## API validation

Serializers validate incoming API payloads before they reach your models. The
[REST API guide](/docs/guides/rest-api/) covers serializer setup and request
handling.

## Next steps

- Review the [Models guide](/docs/guides/models/) for schema patterns.
- Use the [REST API guide](/docs/guides/rest-api/) to validate request data.
