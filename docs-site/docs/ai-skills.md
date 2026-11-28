---
title: AI Agent Skills & llms.txt
description: Using Forge with AI coding agents (Claude Code, Codex, Cursor, OpenCode, Gemini CLI) via Agent Skills and llms.txt.
image: /social-card.png
---

# AI Agent Skills & `llms.txt`

Forge ships machine-readable documentation and modular **Agent Skills** (`agentskills.io` open standard) so AI coding assistants generate accurate, idiomatic Forge v0.2.x code instead of guessing APIs.

## `llms.txt` & `llms-full.txt`

This documentation site publishes clean Markdown indexes at build time:

- **`/forge/llms.txt`**: Concise, categorized directory of every documentation page with one-line summaries.
- **`/forge/llms-full.txt`**: Single-file Markdown bundle of the core Forge documentation for direct context loading.

Point your AI editor, MCP docs tool, or RAG pipeline at `https://forgego.github.io/forge/llms.txt`.

## Agent Skills (`skills/`)

The Forge repository includes 9 task-focused skills under `skills/` (compatible with Claude Code, OpenAI Codex, Cursor, OpenCode, and Gemini CLI):

| Skill | Covers |
|---|---|
| `forge-cli` | `forge new`, `add`, `auth`, `generate`, `makemigrations`, `migrate`, `runserver` |
| `forge-models` | Schema DSL (`schema.*Field`), `Meta`, relations, and `ModelHooks` |
| `forge-orm` | `Manager[T]`, `QuerySet[T]`, `SelectRelated`/`PrefetchRelated`, transactions, `filter/` |
| `forge-migrations-schema` | Migration generation, apply, status, lint, rollback, and dirty-state recovery |
| `forge-admin` | `admin.Register`, `admin.Config[T]`, actions, permissions, React SPA, `UseDatabaseStores` |
| `forge-api` | `ViewSet`, `Serializer`, `api.Router`, auth, permissions, throttling, RFC 7807 errors |
| `forge-identity` | `forge auth`, `identity.SetupIdentitySystem`, password hashing, sessions, tokens |
| `forge-server` | `server.NewServer`, middleware, Viper config, `FORGE_*` env vars, `server.stores: database` |
| `forge-contributor` | Contributing to the Forge repository itself (Go tests, Playwright, docs, CI) |

Each skill includes a concise `SKILL.md` (with explicit negative triggers so agents pick the right skill), pure-Markdown `references/`, `agents/openai.yaml`, and `eval_queries.json`.

## Root `AGENTS.md`

For contributors working inside the Forge repository, the root `AGENTS.md` defines the canonical package map, verification commands, engineering guardrails, and Git collaboration rules.
