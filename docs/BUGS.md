# Forge Known Issues

> Environment and process issues that are **not** code bugs to fix in
> a PR, but that contributors will hit. Salvaged from ops run notes.
> Code bugs live in the tracker; fixed bugs are deleted from here.

## Runner / environment

- [ ] Admin UI unit tests are environment-blocked in some runners:
      vitest/Vite fails with `spawn EPERM`. Re-attempt once spawn
      restrictions are resolved.

## Ecommerce example

- All known ecommerce example blocker bugs resolved in `feat/security-hardening`.
