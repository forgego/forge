# Forge Tech Debt Register

> Known framework debts.
> Each entry names the payoff for fixing it. This file is the backlog;
> `docs/ROADMAP.md` tracks what is scheduled.

## Codegen / ORM

- [x] `SelectRelated` execution with case-insensitive relation resolution,
      `_id` suffix normalization, and pointer/value struct mapping implemented and tested.
- [x] AST parser edge cases: statement-level variable assignment resolution in
      `Fields()`, `Relations()`, and `Meta()`, variadic relation builders, `oneof` choice validation,
      and non-exponential float bounds formatting.
- [x] `validateChoices` validates choice parameters; `unique` tag has a
      registered validator; decimal validators use fixed-point float formatting.

## Admin

- [ ] Admin template rendering, uploads, pickers missing.
- [x] History/audit manager: implemented thread-safe `MemoryHistoryManager`,
      hooked into `Admin[T]` lifecycle (`CreateObject`, `UpdateObject`, `DeleteObject`),
      and tested `/history` REST endpoint.

## Migrations

- [x] Migration recovery and checksum verification: implemented `forge migrate recover`
      CLI command with `--clean` and `--verify` flags, leveraging `execute.Recovery`.

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
- [x] Auto-generated `foreit` skill/instincts updated to accurately
      reflect Go/Forge architecture, conventions, and test commands.

## Test contracts (must keep green)

- Test helpers contract: timestamped DB names, `t.Cleanup()`,
  60s context timeouts, `helpers.Assert*` assertions.
