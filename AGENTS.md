# AGENTS

Portable repo map (relative paths only):

- repository root — Go module `github.com/forgego/forge` (`go 1.26.10+`): `admin/`, `api/`, `cli/`, `cmd/forge/`, `codegen/`, `config/`, `db/`, `errors/`, `filter/`, `identity/`, `internal/`, `log/`, `netutil/`, `orm/`, `registry/`, `schema/`, `server/`, `stores/`, `utils/`, `validate/`
- `tests/` — cross-module integration, CLI E2E, doc samples, and Playwright admin E2E; `examples/` — reference applications (`examples/ecommerce/`)
- `docs/` — `DESIGN.md`, `PRD.md`, `ROADMAP.md`, `TECH-DEBT.md`, `BUGS.md`, `REVIEWING.md`, `RELEASING.md`, `DOCUMENTATION-POLICY.md`, `design/admin-ui-system.md`
- `docs-site/` — user documentation website (Docusaurus); `.github/workflows/` — CI
- `skills/` — public Agent Skills (`agentskills.io` standard); `.agents/skills/foreit/` — repo contributor conventions skill

## Available Skills

- `forge-cli`: `skills/forge-cli/SKILL.md`
- `forge-models`: `skills/forge-models/SKILL.md`
- `forge-orm`: `skills/forge-orm/SKILL.md`
- `forge-migrations-schema`: `skills/forge-migrations-schema/SKILL.md`
- `forge-admin`: `skills/forge-admin/SKILL.md`
- `forge-api`: `skills/forge-api/SKILL.md`
- `forge-identity`: `skills/forge-identity/SKILL.md`
- `forge-server`: `skills/forge-server/SKILL.md`
- `forge-contributor`: `skills/forge-contributor/SKILL.md`
- `foreit`: `.agents/skills/foreit/SKILL.md`

## Structure Rules

- Each public skill must live in its own directory under `skills/`.
- Each skill directory must include a `SKILL.md` with YAML frontmatter (`name` matching the directory name, third-person trigger-rich `description` including negative `Do NOT use for ...` triggers, `license`, `compatibility`, `metadata`).
- Place pure-Markdown supporting docs under `references/` within the skill directory (no Docusaurus frontmatter or MDX syntax).
- Include `agents/openai.yaml` (using the top-level `interface:` block) and `eval_queries.json` (positive and negative trigger eval cases).
- Keep `SKILL.md` concise (< 500 lines) and task-trigger focused.

## Engineering & Verification Guardrails

- **Toolchain & DB**: Go `1.26.10+`; PostgreSQL 15 is the primary CI-gated database (`FORGE_REQUIRE_DB=1`). SQLite is partially tested (SQLite migration apply is experimental). See `docs-site/docs/status.md` as the canonical support contract.
- **Primary Keys**: Only `int64` auto-increment primary keys named `id` are supported by `orm.Manager` and `forge generate --api`.
- **Schema DSL**: Use functional builders (`schema.StringField("name", schema.Required())`), not fluent `.Build()` chains. Embed `<Model>Generated` in model structs.
- **Hook Order**: `BeforeCreate`/`BeforeUpdate` -> `BeforeSave` -> `Clean` -> SQL -> `AfterCreate`/`AfterSave` (`orm/hook_order_contract_test.go`).
- **Verification**:
  - Root Go tests: `go test -race ./...`
  - Doc links: `go run ./internal/tools/doclinks -root .`
  - Docs site: `cd docs-site && npm run build`
  - Admin UI: `cd admin/ui/web && npm run build && npm run lint && npm run test -- --run`
  - Example app: `cd examples/ecommerce && go test ./...`

## Collaboration

- Shared workspace: preserve others' edits; own only what your task assigns.
- State-changing Git operations (`git add`, `git commit`, `git push`, `git reset`, `git clean`, branch creation) require current explicit user authorization; read-only inspection is permitted. Never assume standing approval.

## Code Review Rules

- Trace affected runtime flows from entry point through authorization, validation, business operation, persistence, response, and UI. Trace schema generation and migrations separately.
- Check auth, domain, and transaction invariants against `docs/DESIGN.md`.
- Check compatibility: API, CLI, migration state, supported drivers.
- Check naming, ownership, necessity, and duplication before adding code.
- Cite evidence: tests run, files/lines observed, and gaps left uncovered.
- Use [the full review passes](docs/REVIEWING.md) and [documentation index](docs/README.md).
- For admin UI changes, read [the visual system](docs/design/admin-ui-system.md).
