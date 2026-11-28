---
name: forge-cli
description: Use the Forge CLI (`forge`) for project scaffolding, app/model/api generation, AST codegen, SQL migrations, dev server, checks, and superuser creation. Trigger on `forge new`, `forge add`, `forge auth`, `forge generate`, `forge makemigrations`, `forge migrate`, `forge runserver`, `forge createsuperuser`, `forge check`, `forge test`, or `forge version`. Do NOT use for modifying the Forge framework repository itself (use `forge-contributor`).
license: MIT
compatibility: Go 1.26.10+, github.com/forgego/forge v0.2.x
metadata:
  author: forgego
  version: "0.2"
---

# Forge CLI

`forge` is the command-line workflow for scaffolding projects, generating type-safe model/API code, managing SQL migrations, and running the server.

## Command Surface

| Command | Purpose | Key Flags |
|---|---|---|
| `forge version` | Print CLI version | — |
| `forge new <name>` | Scaffold a new project | `--database` (`postgres`/`sqlite`), `--template` (`simple`/`full`), `--docker=false` |
| `forge add app <name>` | Create `app/<name>/` (`models.go`, `admin.go`, `api.go`) | `--example` |
| `forge add model <name>` | Add a model to an app | `--app <name>` |
| `forge add api <name>` | Add an API file to an app | `--app <name>` |
| `forge add handler <name>` | Add an HTTP handler | `--app <name>` |
| `forge add service <name>` | Add a service | `--app <name>` |
| `forge auth` | Scaffold `app/auth` and wire auth routes in `cmd/server/main.go` | — |
| `forge generate` | AST-parse models and emit `gen.go` (and `api_gen.go` with `--api`) | `--models`, `--output`, `--api`, `--strict` |
| `forge makemigrations <name>` | Diff models vs migration SQL history and write `.up.sql`/`.down.sql` | `--auto`, `--models`, `--migrations` |
| `forge migrate up` | Apply all pending migrations and framework store tables | `--migrations` |
| `forge migrate rollback` | Roll back the last applied migration | `--migrations` |
| `forge migrate status` | Show applied, pending, and dirty migrations | `--migrations` |
| `forge migrate show` | Preview planned migration SQL from current models | `--sql`, `--verbose` |
| `forge migrate lint` | Lint migration files | `--verbose` |
| `forge migrate fake [ver]` | Mark pending or `--fake-initial` migrations as applied | `--fake-initial` |
| `forge migrate force <ver>` | Force migration version in DB | — |
| `forge migrate recover` | Inspect/clean dirty migration state and verify checksums | `--clean`, `--verify` |
| `forge migrate baseline` | Record a baseline migration version | — |
| `forge createsuperuser` | Create an admin/superuser account | — |
| `forge runserver` | Run `cmd/server/main.go` | — |
| `forge check` | Run project config/health checks | — |
| `forge test` | Run project Go tests | — |

## Golden-Path Workflow

```bash
forge new myapp --database postgres --template simple --docker=false
cd myapp
forge add app blog --example
forge generate --models ./app/blog --output ./app/blog
forge makemigrations initial --auto --models ./app/blog
forge migrate up
go run ./cmd/server
```

## Hard Guardrails

- **Version command**: Use `forge version` (there is no `--version` flag).
- **`makemigrations` requires `<name>`**: Always pass a migration name (e.g., `forge makemigrations initial --auto`).
- **`migrate up` requires `up`**: Always run `forge migrate up` (not bare `forge migrate`). There is no `migrate down` or target-version flag; `forge migrate squash` returns `NotImplemented`.
- **Pass `--models` and `--output` together**: `forge generate` defaults `--models` to `./app` if it exists (else `./models`) and `--output` to `./models`. When targeting an app directory, pass both: `forge generate --models ./app/blog --output ./app/blog`.
- **`--api` PK rule**: `forge generate --api` rejects any model whose primary key is not a single auto-increment `int64` named `id`.
- **`forge shell`**: Currently a minimal 5-command loop (`help`, `history`, `clear`, `models`, `dbinfo` where `models`/`dbinfo` print `"pending"`); do not rely on it as a Go REPL.

## References

- [Installation & project layout](references/installation.md)
- [Quickstart walkthrough](references/quickstart.md)
- [AST code generation details](references/code-generation.md)
