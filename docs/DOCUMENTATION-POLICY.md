# Documentation Policy

Scope: what stays public vs. what stays local. This file owns only the policy.

## 1. Docs hierarchy

### Public (durable, curated)

- Root: `README.md`, `CONTRIBUTING.md`, `SECURITY.md`, `AGENTS.md`.
- Docs site source (`docs-site/`), if present, is public.
- `docs/`: `DESIGN.md`, `PRD.md`, `ROADMAP.md`, `TECH-DEBT.md`, `BUGS.md`.
- `docs/design/admin-ui-system.md`: durable design system source.
- Reusable skills (`skills/`, `.agents/skills/` where curated): public.
- Curated durable specs can remain public under `specs/` and need not move to `docs/`.

### Local (approved cleanup, stays local)

- `.codex/`, `.claude/`, `CLAUDE.md`.
- `docs/design/tasks/` (working tasks, drafts).
- `docs/design/admin-ui-tasks.md`, `docs/design/backend-tasks.md`, `docs/design/backend-handoff.md` (scratch, session-local notes).
- Local tooling: `.specify/` and installed speckit skills.
- Do NOT blanket-ignore specs: keep them local until triaged, then
  promote, archive, or delete by explicit decision.
- `specs/001-forge-v1-baseline` stays draft pending curation, do not delete.

## 2. Evidence vs targets vs history

- Evidence: current code, schema, tests, running behavior.
- Targets: `PRD.md`, `ROADMAP.md`, `DESIGN.md` — intent, not proof.
- History: git log and superseded notes — context only, not authority.
- When they disagree, investigate; code may be wrong. Do not automatically change requirements to match bugs; do not edit history to match.

## 3. gitignore is not retroactive

- Adding a path to `.gitignore` does NOT untrack tracked files.
- Tracked files remain tracked until explicit owner-approved untracking
  (`git rm --cached` or equivalent), per path.
- History is not scrubbed: already-pushed content stays in git history
  and in other clones/forks.
- Approved cleanup: coordinator untracks while preserving local files:
  `.codex/`, `.claude/`, `CLAUDE.md`, `docs/design/tasks/`,
  `docs/design/admin-ui-tasks.md`, `docs/design/backend-tasks.md`,
  `docs/design/backend-handoff.md`.
- `.specify/` and installed speckit skills are local and ignored.
- Specs are NOT blanket-ignored: `specs/001-forge-v1-baseline`
  remains an untracked draft, not included in this PR.
- Before any untracking that deletes files from other checkouts
  on pull: back up local copies first and announce.

## 4. Review source

- [Review rules](REVIEWING.md) define evidence and architectural assessment.
- Root [agent guidance](../AGENTS.md) makes that standard discoverable to reviewers.

## 5. Follow-up pipeline

1. Inventory: list local docs, skills, specs, handoffs.
2. Curate: promote (public), keep-local, archive, or delete per item.
3. Links/build: fix cross-links and docs-site build after moves.
4. Review: evaluate the change against [the review standard](REVIEWING.md).

## 6. Status

- Approved cleanup per above; enforcement is per-item by the coordinator.
- This cleanup does not certify all documentation or framework behavior.

## 7. Maintenance backlog and acceptance gates

| Follow-up | Acceptance evidence |
| --- | --- |
| Reconcile capability claims across README, user docs, PRD, and roadmap | Each advertised capability links to current implementation and tests; partial and planned behavior is labeled. |
| Curate the draft v1 specification | Resolve acceptance gaps and separate target requirements from current-state evidence before publication. |
| Extract durable decisions from historical prompts | Public design notes preserve rationale and compatibility implications; session instructions remain local. |
| Audit critical consumer flows | Apply the review guide to generation, migrations, API/admin writes, identity, and custom operations; record paths inspected and missing tests. |
| Maintain documentation links and ownership | Changed local links resolve in the proposed Git tree and the docs-site build passes. |

These are follow-up tasks, not work reported complete by this PR. Keep existing
public architecture and debt documents available while reconciling their claims.
