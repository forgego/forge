---
sidebar_position: 7
description: Add and order middleware in forge applications.
keywords:
  - forge middleware
  - request pipeline
  - http middleware
image: /forge-social-card.svg
---

# Middleware

Middleware lets you run logic before and after request handlers. forge uses a
standard chi-style middleware stack, so ordering matters.

## Add middleware

Register middleware with your router in the order you want it applied:

```go
router.Use(middleware.RequestID)
router.Use(middleware.Logger)
router.Use(middleware.Recoverer)
```

## Common use cases

- Authentication and authorization
- Logging and tracing
- Rate limiting and throttling
- Request/response transformations

## Related docs

- [Request lifecycle](/docs/learn/request-lifecycle/)
- [Security guide](/docs/guides/security/)
