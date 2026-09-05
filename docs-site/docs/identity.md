---
sidebar_position: 23
description: Built-in identity system for users, sessions, and permissions.
image: /forge-social-card.svg
---

# Identity

forge includes a full identity system for user management and authentication.

## What you can do

- Users, groups, permissions, sessions, and tokens
- Pluggable auth backends
- Password policies, lockout, and rate limits
- Middleware for auth requirements

## Components

- Repositories: user, session, token, permission, group
- Services: user, auth, password, permission
- Backends: password, token
- Config: password policy, lockout, rate limits

## Setup example

```go
identitySystem, err := identity.SetupIdentitySystem(db, nil)
if err != nil {
    panic(err)
}
_ = identitySystem
```

## Next steps

- [API Overview](/docs/api/overview/)
- [Server Overview](/docs/server/overview/)
