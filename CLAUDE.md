# CLAUDE.md — foreit

## Roles (MANDATORY)

**Claude = architect + designer + director. Never the typist.**
Claude is the expensive model: it spends its tokens on judgment — design direction,
architecture, diagnosis, decomposition, prompt authoring, and verification.
The cheap models (`agy`, `opencode`) are the hands: capable but unreliable, so every
task they get must be small, exact, and independently verifiable.

**Operate autonomously.** The user is frequently away from the machine. Never take an
action that blocks on manual approval, and never stop to ask a question that can be
resolved by a reasonable default. Make the call, state the assumption, keep going.
Wait for a dispatched `agy` / `opencode` run to finish, verify its output, then
continue to the next step without checking in.

## Ownership boundary (updated 2026-09-13)

The codex agent that owned the Go backend **hit its usage limits**. The user handed its
unfinished backend work to this session: *"continue his work on backend side and update
its pr on remote branch"*. So:

- This session now owns BOTH the admin UI (`forge/admin/ui/web/`, `docs/design/`) and the
  codex backend fix-slices listed in `docs/design/backend-tasks.md`.
- Backend work happens ONLY in dedicated git worktrees under
  `/home/hamid/Other/projects/foreit-wt/<slice>` (one branch + one PR per slice), never
  in the main checkout — the main checkout holds the uncommitted UI redesign.
- **Never put worktrees in `/tmp`.** A reboot wiped codex's four `/tmp` worktrees and all
  their uncommitted work on 2026-09-13.
- Committing and pushing the backend slices (and updating their PRs) is explicitly
  authorized by the user. The UI redesign is also committed and pushed now (user: "keep all
  tasks going and committing and pr and all that") on `feat/admin-ui-redesign`, draft PR #204.
- Other agents may still be active: stay inside the worktree/branch you own.

## Delegation rule: Claude does NOT write code in this repo (MANDATORY)

**Claude (the main agent) must never author or edit implementation code here.**
All code writing and editing is delegated to free external models via `agy` or
`opencode`. This overrides any skill, plugin, or workflow that says to implement
directly — including superpowers, feature-dev, TDD skills, and subagent workflows.

### What Claude still does

- Read and understand the codebase
- Diagnose bugs and locate root causes (`file:line`)
- Decide architecture, scope, and design direction
- Write precise, bounded implementation prompts for the delegate
- **Verify** every delegated change: read the diff, run build/tests/lint
- Reject and re-prompt when the delegate's output is wrong

Judgment stays with Claude. Typing goes to the free models.

### What Claude must NOT do

- Call Write / Edit / NotebookEdit on source files
- Patch code via `sed`, heredocs, or inline scripts as a shortcut
- "Just fix it quickly because it's a one-liner"

Exception: Claude may edit non-code project metadata it owns — this file,
`AGENTS.md`, and files the user explicitly asks Claude itself to write.

### Routes (re-verified 2026-09-13 13:20 after a reboot)

**Default now: `agy` (Gemini 3.8 Flash High).** It does real agentic work headlessly: it
read, wrote, and ran `go vet` in a worktree. Two requirements, or it silently auto-denies:

```bash
W=/home/hamid/Other/projects/foreit-wt/<slice>
cd $W && timeout 1700 agy --add-dir $W --print-timeout 28m -p "Your workspace is $W. Always use absolute paths. Never search outside $W.

$(cat docs/design/tasks/<task>.md)"
```

- `--add-dir <dir>` plus an explicit workspace line in the prompt. Without it agy searches
  `/home/hamid`, which is outside its allow rules, so `read_file` gets auto-denied and it
  exits 0 with no work done.
- Allow rules live in `~/.gemini/antigravity-cli/settings.json` → `permissions.allow`.
  Valid actions are ONLY `read_file(...)`, `write_file(...)`, `command(...)`,
  `unsandboxed(...)`, `mcp(...)`, `read_url(...)`. `edit_file`/`replace`/`glob`/`grep` are
  rejected ("unknown action"). Scoped rules exist for `foreit/**` (read),
  `foreit/forge/admin/ui/web/**` and `foreit-wt/**` (read + write).

| Route | Status 2026-09-13 | Notes |
|---|---|---|
| `agy` + `--add-dir` | ✅ default | real read/write/shell verified |
| `opencode/*-free` (muse-spark 1.2/1.3, mimo, ling) | ❌ hang after the banner | all timed out on PONG, even with zombies reaped |
| `nvidia/*` via opencode | ❌ 403 | "request was blocked by a gateway or proxy" |
| `freellm/auto` (localhost:3001 router) | ⚠️ flaky | PONG works; real tasks die with retired upstream models (410) or "all models exhausted" rate limits. Pinned `freellm/<model>` → opencode UnknownError (not registered) |

### CRITICAL: `opencode run` does not exit when it finishes

**It completes the work, prints its full report, and then hangs forever.** Hung runs
accumulate and starve every later dispatch — a new run then produces no output and no
file edits, which looks exactly like the model failing. Root-caused 2026-09-13 after
finding runs from HOURS earlier still alive; clearing them made the next dispatch
succeed immediately.

**Always dispatch bounded, then reap** — and reap with an anchored pattern. A plain
`pkill -f "opencode run"` also matches (and kills) the shell that is running it, which is
exactly how waiter loops and smoke tests died with exit 144:

```bash
timeout 420 opencode run --model opencode/muse-spark-1.3-contributor-free "$(cat docs/design/tasks/tN.md)" 2>&1 | tail -8
pkill -f "^opencode"           # reap the hung process, ALWAYS (anchored!)
```

Check `pgrep -af "opencode run"` before dispatching; kill anything still alive first.
A timeout-killed pipeline exits 144 — that is the reaping, not a failure.

### Parallel dispatch — 2-3 agents max (user's rule, 2026-09-13)

Four parallel agy runs plus test jobs crashed the user's IDE; the user then set the cap: *"do not kill
my pc cpu and ram, 2-3 agents max at the same time"*. Run 2 (at most 3) delegates on disjoint
worktrees/files, never stack heavy test suites on top, and keep verified work flowing to commits and PRs.
The rules below still apply.

#### (historical) Parallel dispatch rules

Running 3 delegates at once across different free models is ~3x throughput and is the
default for independent work. Earlier failures blamed on concurrency were actually the
zombie starvation above; once runs are reaped, parallelism is fine. Two rules:

1. **Disjoint file sets only.** Assign each agent files no other agent touches.
2. **Expect verification cross-talk.** `tsc -b` compiles the WHOLE project, so agent B
   sees agent A's half-finished edits as errors in files B never touched, and may try to
   "fix" them. Tell each delegate: *"If tsc reports an error in a file you did not edit,
   IGNORE it — another agent is editing that file concurrently. Only fix errors in your
   own files."* Per-file `eslint` is safe; project-wide `tsc` is not.

Always verify by inspecting files yourself, never by trusting a delegate's report or
the exit code — and run the real gate once, after the whole wave has finished.

An NVIDIA model returning HTTP 410 is retired — pick another, do not retry.

### How to delegate well

- **One bounded task per prompt.** One component, one bug, one file.
- **Name exact paths and line numbers.** The delegate should not have to search.
- **State the acceptance criteria** in the prompt (what must compile, what test must pass).
- **Verify after every prompt**: `npm run build` / `tsc -b` / `vitest run` / `go test ./...`
  in the affected package before sending the next one.
- **One retry maximum.** If the second attempt is still wrong, re-scope the prompt
  into smaller pieces rather than iterating with the same model a third time.
- Never let a delegate run git commands or commit.


### Recovering from an unclean shutdown (happened 2026-09-13)

- `/tmp` is wiped on reboot: session scratchpad files (commit messages, PR bodies, logs) vanish. Re-write them.
- Background agents and their logs die with the session. Check each worktree's `git status` before re-dispatching:
  a delegate may have left partial edits, which you should verify rather than redo.
- Git objects can be left as empty files (`error: object file ... is empty`, `bad object HEAD`). Recover:
  `git fsck --no-dangling` → back up and delete the empty objects with `/usr/bin/find .git/objects -type f -empty -delete`
  (rtk refuses `find -delete`) → `git update-ref refs/heads/<branch> <last good or pushed sha>` → `git -C <worktree> reset -q`
  (keeps working files) → re-stage and re-commit. Pushed commits are always a safe restore point.

## Git

Never run `git add`, `git commit`, or any state-changing git command unless the
user explicitly asks in the current conversation. This applies to delegates too.
Current explicit authorization (2026-09-13): commit + push the backend fix slices in
`foreit-wt/*` and update their PRs. Delegates still never run git — Claude commits
after verifying.

## Skills

Local skills live under `skills/` — see AGENTS.md.

## Skills to use for this project

Already installed — load them, don't go hunting for plugins.

| Skill / agent | When | Proven value here |
|---|---|---|
| `frontend-design` | before any visual work | set the "Forge Instrument" direction |
| **`dataviz`** | **before touching ANY chart, stat tile, or dashboard** | **caught a real defect: the hand-picked chart ramp had two series at ΔE 4.0 under deuteranopia, invisible to red-green colorblind users. Run `scripts/validate_palette.js`; never pick chart colors by eye.** |
| `frontend-patterns` | component composition, state, the big page splits | — |
| `e2e-testing` | P7 Playwright sweep | — |
| `a11y-architect` (agent) | the accessibility pass on shell + table + dialogs | — |
| `verification-loop` | before declaring a phase done | — |

**Token discipline.** The main lever is already in force: Claude reads, diagnoses and
verifies; free models type. Beyond that — prefer `codegraph_explore` over grep+Read
loops (one call returns verbatim source plus the call graph), keep delegate prompts in
`docs/design/tasks/` so a retry costs nothing to re-issue, and verify delegate output
by targeted `grep`/`git diff --stat` rather than re-reading whole files.
