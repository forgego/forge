---
title: Security Settings
description: Application, CSRF, and session secrets and CSRF path exemptions.
image: /social-card.png
---

# Security Settings

`SecuritySettings` (`config/settings.go`) configures secrets and CSRF protection in `server.NewServer`.

## Keys & Defaults

| Key | Env Override | Default | Description |
|---|---|---|---|
| `security.secret_key` | `FORGE_SECURITY_SECRET_KEY` | `""` | Primary application secret |
| `security.csrf_secret_key` | `FORGE_SECURITY_CSRF_SECRET_KEY` | `""` | Enables CSRF middleware when non-empty |
| `security.session_secret` | `FORGE_SECURITY_SESSION_SECRET` | `""` | Enables `scs` session middleware when non-empty |
| `security.csrf_exempt_paths` | `FORGE_SECURITY_CSRF_EXEMPT_PATHS` | `nil` | Path prefixes exempt from CSRF checks (e.g. token-authenticated API routes) |

## Production Validation

When `app.env` is `production`, `Server.Start` / `StartWithGracefulShutdown` fails closed unless `secret_key`, `csrf_secret_key`, and `session_secret` are all:
1. Non-empty and not placeholder strings (such as `change-me`).
2. At least 32 characters long.
3. Distinct from one another.

Generate independent secrets with:

```bash
openssl rand -hex 32
```
