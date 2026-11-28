---
title: Upgrading Forge
description: How to upgrade between Forge releases, regenerate code, and evolve migrations safely.
image: /social-card.png
---

# Upgrading Forge

Forge is currently pre-1.0 (`v0.2.x`). Follow this checklist whenever you upgrade the framework or CLI.

## Upgrade Steps

### 1. Update the CLI and Library Together

Keep the `forge` CLI at the same version as `github.com/forgego/forge` in `go.mod`:

```bash
go get github.com/forgego/forge@latest
go mod tidy
go install github.com/forgego/forge/cmd/forge@$(go list -m -f '{{.Version}}' github.com/forgego/forge)
forge version
```

### 2. Regenerate Model & API Code

Hand edits to `gen.go` and `api_gen.go` are not preserved. Re-run `forge generate` for every app package:

```bash
forge generate --models ./app/blog --output ./app/blog
# Or with REST API generation:
forge generate --models ./app/blog --output ./app/blog --api
```

### 3. Check Migrations & Bookkeeping

Applied migrations are never rewritten by Forge. After upgrading:

```bash
forge migrate status
forge migrate lint --verbose
forge migrate recover --verify
```

If `server.stores` is `database`, `forge migrate up` automatically applies any new framework store migrations (`forge_framework_migrations`) before your application migrations.

## Version-Specific Notes

See the [Changelog](/docs/changelog/) for breaking changes and release-specific migration steps (including upgrading from `v0.1.1` to `v0.2.x`).
