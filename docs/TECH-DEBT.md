# Forge Tech Debt Register

> Debts salvaged from `docs/archive/ROADMAP.md`, ops run notes
> (`STATUS.md`, `todo-current.json`), and the skills backlog.
> Each entry names the payoff for fixing it. This file is the backlog;
> `docs/ROADMAP.md` tracks what is scheduled.

## Codegen / ORM

- [ ] `SelectRelated`/`PrefetchRelated` + aggregates/annotations surfaces
      exist but execution is structure-ready only.
      Payoff: kills N+1 query class framework-wide.
- [ ] AST parser partial: field/relation/meta/hook/validation-tag
      extraction is incomplete for edge-case model files.
      Payoff: reliable `forge generate` on any model layout.
- [ ] `validateChoices` always returns true; `unique` tag has no
      registered validator; decimal validators miscount via `%g`.
      Payoff: validation users can trust.

## Admin

- [ ] Admin template rendering, uploads, history/audit, pickers missing.
- [ ] History/audit manager is a no-op while `/history` + `useModelHistory`
      are exposed (always empty).

## Migrations

- [ ] Migration recovery, drift detection, checksum verification,
      partial rollback missing (fail-loud + force/recover intended).

## Tooling / repo hygiene

- [ ] TODO scanner reports ~500 hits but only 1 is first-party
      (`forge/log/encoder.go:258`, trivial comment); the rest is
      lockfiles, `docs-site/build`, `helper-projects`, `.kilocode` noise.
      Scope the scanner to `forge/`, `examples/`, `tests/`.
- [ ] TODO-history churn is self-referential (counts ops run logs as
      new/resolved TODOs). Exclude `ops/` from the scan.
- [ ] Runner env gotchas: default Go build cache may be access-denied
      (rerun with a temp `GOCACHE`); `git dubious ownership` needs the
      safe-directory policy. Document in CI/runner notes.
- [ ] Auto-generated `foreit` skill/instincts claim a TypeScript/docs-site
      stack from a single dependabot commit while the repo is Go/forge.
      Regenerate or delete the ECC `foreit-*` bundle.

## Test contracts (must keep green)

- Test helpers contract: timestamped DB names, `t.Cleanup()`,
  60s context timeouts, `helpers.Assert*` assertions.
