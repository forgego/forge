---
sidebar_position: 23
description: Built-in identity system for users, sessions, and permissions.
image: /social-card.png
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

Password reset delivery requires a `service.PasswordResetNotifier`. The password
service returned by `SetupIdentitySystem` has no notifier, so
`RequestPasswordReset` returns a configuration error. To enable reset requests,
construct a service with `service.NewPasswordServiceWithNotifier` and your
notifier, then assign it to `identitySystem.PasswordService`.

Custom user repositories can implement `repository.UserMutations` to support
column-targeted login, password, and profile writes. Implement
`repository.UserEmailMutations` as well to reject an email update when the
address changed since it was read, even when the submitted email is unchanged.
Without `UserMutations`, services use the existing full-row `Update` method for
compatibility; those non-atomic writes can overwrite a concurrent password,
lock, or email change. Without `UserEmailMutations`, email updates and
`email_verified` writes have no address condition and can apply after a
concurrent email change. Custom repositories should implement both interfaces.

A custom token repository uses a fail-closed sequence for password reset and
email verification: it consumes the token before changing the user, so a later
user write failure may leave the token consumed. The built-in SQL user and
token repositories commit both writes in one transaction when they share a
database. A custom token repository can implement `repository.TokenCommitRepository`
to own that atomic commit explicitly; a transaction claim inherited from a
wrapped repository does not opt it in.

## Next steps

- [API Overview](/docs/api/overview/)
- [Server Overview](/docs/server/overview/)
