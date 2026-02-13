---
sidebar_position: 4
description: Authentication classes and defaults.
image: /forge-social-card.svg
---

# Authentication

Authentication classes include token, session, JWT, basic, and API key auth.

## Set defaults

```go
api.SetDefaultAuthentication(
    authentication.NewTokenAuthentication(tokenResolver),
)
```

## Built-in auth classes

- TokenAuthentication
- SessionAuthentication
- JWTAuthentication
- BasicAuthentication
- APIKeyAuthentication

## Next steps

- [Permissions](/docs/api/permissions/)
- [Throttling](/docs/api/throttling/)
