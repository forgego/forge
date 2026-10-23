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

Inactive and locked accounts cannot authenticate through password login, token
refresh, session lookup, or API authentication. Deactivating or locking an
account revokes its existing sessions. `PasswordService.ChangePassword`
revokes all of the user's sessions except the one named by
`ChangePasswordRequest.CurrentSessionKey`; leave that field empty when the
password change should end every session.

Custom repositories that need the same atomic revocation must implement
`repository.AccountStateMutations` and `repository.PasswordSessionMutations`.
The built-in SQL repository implements both; older custom repositories remain
compatible with the existing repository interfaces.

## Components

- Repositories: user, session, token, permission, group
- Services: user, auth, password, permission
- Backends: password, token
- Config: password policy, lockout, rate limits

## User API access

All `/api/users` routes require authentication. Listing users requires staff or
superuser status. A user can retrieve their own record; staff and superusers can
retrieve any record. A user can update their own email, first name, and last name.
Only superusers can update another user or set `is_active`, `is_staff`, or
`is_locked`. Attempts to set those flags without superuser status return 403
without saving any part of the request. Creating and deleting users requires
superuser status. `/api/auth/logout` only revokes a session belonging to the
authenticated user.

## Setup example

### Rotating JWT signing keys

Set `security.secret_key` and `security.secret_key_id` to the new current key
and its public identifier. Add each retired key to
`security.previous_secret_keys` with its `id`, secret `key`, and RFC3339
`verify_until` deadline. Tokens include a `kid`; verification selects only that
key, rejects unknown IDs, and accepts a previous key only before its deadline.
New tokens must be signed with the current key. Set
`security.token_max_lifetime` to the maximum issued token lifetime and retain a
previous key for at least that long after switching issuers. Remove expired
keys after the overlap. Never log these secret values.

The `forge auth` scaffold passes `settings.Security` to its JWT authenticator.
When a key ID or previous key is configured, issued login tokens include the
current `kid`; API authentication accepts a previous key only until its
`verify_until` deadline. Without rotation keys, the scaffold keeps the existing
single-key behavior. `security.token_max_lifetime` controls the lifetime of
login tokens issued by the scaffold.

Changing `security.csrf_secret_key` invalidates outstanding CSRF form tokens;
users may need to reload forms before submitting them.

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
