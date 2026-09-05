---
name: forge-cli
description: Use Forge CLI commands for project scaffolding, code generation, migrations, and running the dev server. Trigger this skill for `forge new`, `forge generate`, `forge makemigrations`, `forge migrate`, and `forge runserver` workflows.
---

# Forge CLI

## Overview
Forge CLI is the primary developer workflow: scaffolding projects, generating code, applying migrations, and running the server.

## When to Use
- The task mentions `forge` CLI commands or project scaffolding.
- You need code generation, migrations, or to run the dev server.
- The user asks about `forge new`, `forge generate`, `forge makemigrations`, `forge migrate`, or `forge runserver`.

## Quick Start
1. Install the CLI and verify `forge --help`.
2. Create a project with `forge new`.
3. Run `forge generate`, `forge makemigrations`, `forge migrate`, and `forge runserver`.

## Common Tasks
- Regenerate code after model changes with `forge generate`.
- Inspect migration status with `forge migrate status`.
- Start the dev server with `forge runserver`.

## Gotchas
- CLI commands assume your project layout matches Forge expectations.
- Model changes typically require codegen and migrations in sequence.

## References
- [CLI installation](references/installation.md)
- [Quickstart flow](references/quickstart.md)
- [Code generation](references/code-generation.md)
