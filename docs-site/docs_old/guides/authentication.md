---
sidebar_position: 5
description: Authentication options for forge APIs.
keywords:
  - forge auth
  - authentication
image: /forge-social-card.svg
---

# Authentication Guide

Choose an auth backend and apply it to your routes.

## Common setup

```go
router.Use(api.TokenAuthenticationMiddleware)
```

## Next steps

- [REST API guide](/docs/guides/rest-api)
- [Security guide](/docs/guides/security)
