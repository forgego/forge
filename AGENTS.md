# AGENTS

Portable repo map (relative paths only):

- `forge/` — runtime: `admin/`, `api/`, `cli/`, `codegen/`, `db/migrate/`, `filter/`, `identity/`, `orm/`, `schema/`, `server/`
- `tests/` — cross-module integration; `examples/` — usage patterns
- `docs/` — `DESIGN.md`, `PRD.md`, `ROADMAP.md`, `TECH-DEBT.md`, `BUGS.md`, `REVIEWING.md`
- `skills/` — reusable contributor guidance (see below)
- `docs-site/` — user documentation; `scripts/`, `.github/workflows/` — tooling and CI

## Available Skills

- forge-admin: `skills/forge-admin/SKILL.md`
- forge-api: `skills/forge-api/SKILL.md`
- forge-cli: `skills/forge-cli/SKILL.md`
- forge-migrations-schema: `skills/forge-migrations-schema/SKILL.md`
- forge-models: `skills/forge-models/SKILL.md`
- forge-contributor: `skills/forge-contributor/SKILL.md`

## Structure Rules

- Each skill must live in its own directory under `skills/`.
- Each skill directory must include a `SKILL.md` with YAML frontmatter that matches the directory name.
- Place supporting docs under `references/` within the skill directory.
- Keep skill instructions concise and task-trigger focused.

## Collaboration

- Shared workspace: preserve others' edits; own only what your task assigns.
- State-changing Git operations require current explicit user authorization; read-only inspection is permitted. Never assume standing approval.

## Code Review Rules

- Trace affected runtime flows from entry point through authorization, validation, business operation, persistence, response, and UI. Trace schema generation and migrations separately.
- Check auth, domain, and transaction invariants against `docs/DESIGN.md`.
- Check compatibility: API, CLI, migration state, supported drivers.
- Check naming, ownership, necessity, and duplication before adding code.
- Cite evidence: tests run, files/lines observed, and gaps left uncovered.
- Use [the full review passes](docs/REVIEWING.md) and [documentation index](docs/README.md).
- For admin UI changes, read [the visual system](docs/design/admin-ui-system.md).
