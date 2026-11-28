---
title: App Settings
description: Application-level settings (name, env, version, debug).
image: /social-card.png
---

# App Settings

`AppSettings` (`config/settings.go`) controls application identity and environment mode.

## Keys & Defaults

| Key | Env Override | Default | Description |
|---|---|---|---|
| `app.name` | `FORGE_APP_NAME` | `"forge"` | Application display name |
| `app.env` | `FORGE_APP_ENV` | `"development"` | Environment (`development`, `staging`, `production`) |
| `app.version` | `FORGE_APP_VERSION` | `"0.1.0"` | Application version string |
| `app.debug` | `FORGE_APP_DEBUG` | `false` | Enables debug behavior (e.g. profiling routes when `server.enable_profiling` is true) |

## Production Rules

When `app.env` is `production` (case-insensitive):
- `Server.Start` / `StartWithGracefulShutdown` refuses to listen if `app.debug` is `true`.
- Session and CSRF cookies set `Secure: true`.
- All three `security.*` secrets must be explicitly set, at least 32 characters long, and distinct.

```yaml
app:
  name: store
  env: production
  version: 1.0.0
  debug: false
```
