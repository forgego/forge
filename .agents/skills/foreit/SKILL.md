---
name: foreit-conventions
description: Development conventions, architecture, and testing patterns for Forge (foreit) in Go.
---

# Foreit / Forge Conventions

Forge is a type-safe, full-stack Go web framework with batteries included (ORM, migrations, admin, REST API, identity, and security).

## Tech Stack

- **Primary Language**: Go (1.21+)
- **Architecture**: Modular framework packages under `forge/` with reference application under `examples/ecommerce/`
- **Docs**: Docusaurus site under `docs-site/`

## Core Directory Layout

- `forge/`: Framework modules
  - `admin/`: Auto-generated web admin interface and REST metadata API
  - `api/`: REST API framework (serializers, viewsets, filters, permissions, throttling, OpenAPI)
  - `config/`: Viper-based configuration and ephemeral secret generation
  - `db/`: Database abstraction (PostgreSQL, SQLite), connection pooling, migrations
  - `identity/`: Authentication, session management, password hashing (bcrypt), token repositories
  - `orm/`: Generic QuerySet builder, schema registry, type-safe field accessors
  - `schema/`: Model schema definition, field types, relation definitions, and hooks
  - `server/`: HTTP server, chi router, CSRF, secure cookies, and middleware
- `examples/ecommerce/`: Comprehensive reference store demonstrating all Forge subsystems
- `docs/`: Authoritative architectural and roadmap docs (`DESIGN.md`, `ROADMAP.md`, `TECH-DEBT.md`, `BUGS.md`)

## Testing & Verification

- Framework tests: `go test ./...` inside `forge/`
- Reference app tests: `go test ./...` inside `examples/ecommerce/`
- Always verify tests pass before committing.

## Commit Conventions

Follow Conventional Commits:
- `feat(subsystem): description`
- `fix(subsystem): description`
- `test(subsystem): description`
- `docs(subsystem): description`
- `refactor(subsystem): description`
