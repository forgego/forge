---
name: forge-identity
description: Implement authentication, password hashing, users, groups, permissions, sessions, and tokens in Forge using `github.com/forgego/forge/identity` and `forge auth`. Use when setting up user auth, password resets, `RequirePermission`, or `SetupIdentitySystem`. Do NOT use for REST API ViewSet serializer fields (use `forge-api`) or admin `ListDisplay` config (use `forge-admin`).
license: MIT
compatibility: Go 1.26.10+, PostgreSQL 15
metadata:
  author: forgego
  version: "0.2"
---

# Forge Identity & Authentication

`github.com/forgego/forge/identity` implements a repository + service + backend architecture for users, bcrypt password hashing, sessions, tokens, password resets, groups, and permissions.

## Golden Path

Scaffold auth into a project:

```bash
forge auth
```

Or wire `identity.SetupIdentitySystem` programmatically:

```go
idSys, err := identity.SetupIdentitySystem(database, nil)
if err != nil {
	return err
}
```

Password hashing & verification (`identity/password.go`):

```go
hashed, err := identity.HashPassword("s3cr3t-password")
ok := identity.VerifyPassword("s3cr3t-password", hashed)
```

## Hard Guardrails

- **No OAuth/OIDC, social login, or field-level permissions**: Only password/token/session auth and user/group/model permissions are implemented (`docs-site/docs/status.md`).
- **No custom user model**: `identity/repository/user.go` uses `identity/models.User`.
- **Avoid double-hashing**: `identity` guards against re-hashing strings with `$2a$` prefix, and `forge auth` + `forge generate --api` can duplicate user serializers if both are run in the same package — check generated `api_gen.go` if customizing `app/auth`.
- **`RequirePermission`**: Enforces `PermissionService.CheckPermission` with superuser (`IsSuperuser`) bypass and staff fallback.

## References

- [Identity repositories, services & backends](references/identity.md)
