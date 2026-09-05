---
sidebar_position: 6
description: Input validation with serializers and schema rules.
keywords:
  - forge validation
  - serializers
image: /forge-social-card.svg
---

# Validation Guide

Validate input at the serializer layer.

## Example

```go
s.AddField("email", api.EmailField())
s.AddField("age", api.IntegerField(api.Min(18)))
```

## Next steps

- [REST API guide](/docs/guides/rest-api)
- [Error handling](/docs/guides/error-handling)
