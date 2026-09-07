# Forge Known Issues

> Environment and process issues that are **not** code bugs to fix in
> a PR, but that contributors will hit. Salvaged from ops run notes.
> Code bugs live in the tracker; fixed bugs are deleted from here.

## Runner / environment

- [ ] Admin UI unit tests are environment-blocked in some runners:
      vitest/Vite fails with `spawn EPERM`. Re-attempt once spawn
      restrictions are resolved.
- [ ] The 24x7 orchestrator degraded after ~09:45 (exit-1/no-message
      loop, dozens of empty message files, identical snapshots).
      Diagnose that failure mode before any 24x7 rerun.

## Ecommerce example

- [ ] `Mark Delivered` must set `delivered_at` (currently only flips status).
- [ ] Required-field saves must return validation errors, not 500s.
