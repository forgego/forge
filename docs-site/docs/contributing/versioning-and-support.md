---
sidebar_position: 33
title: Versioning and Support
description: Release policy, semantic versioning guarantees, supported versions, and deprecation process.
---

# Versioning and Support

This document defines the versioning policy, compatibility guarantees, support tiers, and deprecation lifecycle for Forge. For development environment setup and contributor workflows, see [Development](development.md) and [CONTRIBUTING.md](../../../CONTRIBUTING.md).

## Semantic Versioning

Forge follows [Semantic Versioning 2.0.0](https://semver.org/) (`vMAJOR.MINOR.PATCH`).

### Public API

The public API consists of:

- Exported Go packages and identifiers in `github.com/forgego/forge` (excluding `internal/`).
- CLI commands, subcommands, and flags provided by `cmd/forge` and `cli`.
- Application configuration keys and `FORGE_*` environment variables.
- The structure and signature of generated code produced by `forge generate`.
- The admin extension API contract and `admin.ExtensionAPIVersion` (currently `1`).
- The REST API RFC 7807 error envelope format.

### Private APIs and Internals

The following are internal and may change at any time without a deprecation cycle:

- All packages and code under `internal/`.
- Admin UI internals, unexported frontend components, and private frontend state.
- Test suites, fixtures, and files under `tests/testdata/`.

### Versioning Rules and History

- **Release history:** Historical release tags exist up to `v1.0.1`, including `v0.1.1` and `v1.0.0`. Tags `v1.0.0` and `v1.0.1` were tagged prematurely and are retracted in `go.mod` (`v1.0.0` contained no Go packages; `v1.0.1` retracted `v1.0.0`). The Go module proxy and `@latest` ignore retracted tags. Because published tags cannot be deleted or reused in the Go ecosystem, the first official stable 1.x release will be tagged above `v1.0.1`.
- **Pre-1.0 (`v0.x`):** Minor releases (`v0.1.0`, `v0.2.0`, etc.) may introduce breaking changes to the public API, configuration, or generated code. All breaking changes are documented in [Changelog](../changelog.md) entries with upgrade instructions. Patch releases (`v0.x.Y`) contain backwards-compatible bug and security fixes.
- **Post-1.0 (`v1.x` and later):**
  - **MAJOR (`vX.0.0`):** Breaking changes to the public API.
  - **MINOR (`v1.X.0`):** Backwards-compatible features, additions, and deprecations.
  - **PATCH (`v1.0.X`):** Backwards-compatible bug fixes and security patches.

For component support tiers and verification evidence, see the [Support contract](../status.md).

## Supported Versions

In accordance with [SECURITY.md](../../../SECURITY.md), version support is maintained as follows:

| Version | Supported | Maintenance Policy |
| :--- | :--- | :--- |
| `0.1.x (latest)` | :white_check_mark: | Receives security fixes and bug fixes. |
| `< 0.1` | :x: | Pre-0.1 development versions. Unsupported. |
| `1.0.0, 1.0.1 (retracted)` | :x: | Retracted in go.mod. Unsupported. |

### Patch and Maintenance Policy

- Maintenance bug fixes are applied only to the latest minor release of the active major version.
- Security fixes are backported to the latest minor release of the current major version.
- Minor releases older than the current latest minor release, as well as pre-1.0 releases, do not receive maintenance or security patches.

## Deprecation Windows

When an exported API, CLI flag, configuration key, or feature is marked for deprecation:

1. **Deprecation notice:**
   - In Go code, mark the item with a doc comment beginning with `// Deprecated: ` describing the reason and the replacement path.
   - For CLI flags or configurations, log a warning when the deprecated option is used.
   - List the deprecation in the [Changelog](../changelog.md) for that release.
2. **Deprecation window:**
   - A deprecated item remains supported for a minimum of two minor releases (`MINOR + 2`).
   - Deprecated items are removed no earlier than the next major release. For example, an item deprecated in `v1.2.0` remains supported through `v1.3.0` and `v1.4.0`, and may only be removed in `v2.0.0`.

## Upgrade Guides

When a release introduces breaking changes or migration steps:

- Upgrade guides are published directly within the corresponding [Changelog](../changelog.md) release entry.
- The internal release process and upgrade template are documented in [docs/RELEASING.md](../../../docs/RELEASING.md).
- Each upgrade guide must include:
  - An overview of breaking changes and affected consumers.
  - Code regeneration steps using `forge generate`.
  - Required schema migration steps using `forge migrate up`.
  - Configuration key or environment variable renames and additions.
  - Code refactoring requirements with before-and-after examples.
  - Rollback instructions.

For detailed schema migration and recovery procedures, refer to [Migrations & Disaster Recovery](../migrations.md).

## Reporting Vulnerabilities

Do not report security vulnerabilities through public GitHub issues.

To report a vulnerability, review and follow the instructions in [SECURITY.md](../../../SECURITY.md).
