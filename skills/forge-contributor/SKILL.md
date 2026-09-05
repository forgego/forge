---
name: forge-contributor
description: Contributor workflow for working on the Forge framework itself (core Go, tests, admin UI, docs site, examples, CI/security validations, and release checklist). Use when changing Forge internals or contributing to framework development (not building apps with Forge).
---

# Forge Contributor

## Scope
- Use this skill for framework changes: core Go code, tests, docs site, admin UI, examples/ecommerce, or CI workflows.
- Do not use this skill for app development with Forge; use the Forge CLI or API skills instead.

## When to Use
- The task involves changing Forge internals, CI, or contributor workflows.
- You need to run or adjust framework-level validations.
- The user asks about release or security checks.

## Workflow
1. Identify which areas you touched: `forge/`, `tests/`, `docs-site/`, `forge/admin/ui/`, `examples/`, or `.github/workflows/`.
2. Run the required validations for those areas.
3. Update contributor documentation if the workflow or test strategy changed.
4. Make sure CI and security checks will pass for the changed areas.

## Required Validations (Summary)
- Go core: unit tests, lint, build, and security checks.
- Tests suite: integration tests and CLI e2e tests when core changes affect migrations, ORM, or CLI.
- Admin UI: build, lint, and unit tests when `forge/admin/ui/**` changes.
- Docs site: build the docs site when `docs-site/**` changes.
- Examples: build and test ecommerce sample when `examples/ecommerce/**` changes or core changes affect it.

## Documentation Updates
- Update `docs-site/docs/contributing/development.md` if the contributor workflow changes.
- Update `tests/README.md` and `tests/TESTING.md` if test strategy or coverage changes.

## References
- [CI checklist](references/ci-checklist.md)
- [Contributor resources](references/contributor-resources.md)
