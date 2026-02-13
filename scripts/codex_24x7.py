#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import socket
import subprocess
import sys
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any
from urllib.parse import urlparse


RULES_TEMPLATE = """# Codex 24x7 Rules

1. Always read `ops/codex-24x7/GOALS.md` and `ops/codex-24x7/PLAN.md` first.
2. Pick a focused set of high-impact tasks each run.
3. Prioritize reliability, tests, example parity, and TODO/FIXME debt.
4. For UI/admin improvements, reuse patterns/components from `helper-projects/meta-shell` where helpful.
5. Run targeted tests for every changed area before finishing a run.
6. For framework coverage, continuously work through:
   - admin features
   - ORM + schema + migrations
   - API + server behavior
   - examples/ecommerce parity for new/existing features
7. Update this file set each run:
   - `ops/codex-24x7/STATUS.md`
   - `ops/codex-24x7/HISTORY.md`
8. In STATUS, always include:
   - Completed this run
   - Remaining work
   - Next run plan
9. If blocked, record exact blocker and move to another unblocked task.
"""

GOALS_TEMPLATE = """# Codex 24x7 Goals

## Primary Goal
Make Forge production-ready by closing framework gaps and proving features end-to-end through the ecommerce example.

## Mandatory Coverage Areas
- Admin: CRUD flows, filters, actions, auth, permissions, UI stability.
- ORM: query building, relations, aggregation, update/select/prefetch behaviors.
- Schema: model definitions, traits, hooks, relation integrity.
- Migrations: generation, apply, rollback, drift/safety checks.
- API/Server: serializers, viewsets, auth, permissions, pagination, throttling, error handling, middleware.
- Example parity: add framework features to `examples/ecommerce` and test them there.

## Definition of Done (rolling)
1. Every new/updated framework feature has test coverage in framework code.
2. Feature is represented in ecommerce example when applicable.
3. Ecommerce tests pass for those features.
4. `ops/codex-24x7/STATUS.md` and `ops/codex-24x7/HISTORY.md` are updated.
"""

PLAN_TEMPLATE = """# Codex 24x7 Plan

## Phase 1: Inventory + Baseline
- Enumerate TODO/FIXME and map to components.
- Build current coverage map for admin, ORM, schema, migrations, API/server, ecommerce.
- Identify highest-risk production blockers.

## Phase 2: Reliability Core
- Fix high-severity framework bugs first.
- Add missing tests around bug fixes.
- Keep migrations/schema/API compatibility stable.

## Phase 3: Ecommerce Parity
- Add missing framework feature usage to ecommerce example.
- Add/expand ecommerce integration + UI tests for those features.

## Phase 4: Production Readiness
- Harden security paths, error behavior, and operability checks.
- Reduce flaky tests and improve deterministic CI behavior.
- Keep status/history current with remaining blocker list.
"""

STATUS_TEMPLATE = """# Codex 24x7 Status

## Completed this run
- (pending)

## Remaining work
- (pending)

## Next run plan
- (pending)
"""

HISTORY_TEMPLATE = "# Codex 24x7 History\n"


@dataclass
class Config:
    repo_root: Path
    interval_minutes: int
    mode: str
    one_shot: bool
    state_root: Path
    max_todo_lines: int
    max_plan_files: int
    max_file_preview_lines: int
    max_run_minutes: int
    disable_auto_proxy: bool
    proxy_url: str


class Runner:
    def __init__(self, cfg: Config):
        self.cfg = cfg
        self.runs_dir = cfg.state_root / "runs"
        self.state_dir = cfg.state_root / "state"
        self.runner_log = cfg.state_root / "runner.log"
        self.rules_path = cfg.state_root / "RULES.md"
        self.goals_path = cfg.state_root / "GOALS.md"
        self.plan_path = cfg.state_root / "PLAN.md"
        self.status_path = cfg.state_root / "STATUS.md"
        self.history_path = cfg.state_root / "HISTORY.md"
        self.todo_current_path = self.state_dir / "todo-current.json"
        self.todo_history_path = self.state_dir / "todo-history.jsonl"
        self.last_prompt_path = self.state_dir / "last-prompt.md"
        self.active_run_path = self.state_dir / "active-run.json"
        self.lock_path = cfg.state_root / "daemon.lock"
        self.stop_path = cfg.state_root / "STOP"
        self._lock_fd: int | None = None
        self._missing_command_warnings: set[str] = set()

    def log(self, message: str) -> None:
        line = f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] {message}"
        print(line, flush=True)
        for i in range(3):
            try:
                with self.runner_log.open("a", encoding="utf-8") as f:
                    f.write(line + "\n")
                return
            except OSError:
                if i < 2:
                    time.sleep(0.12)
        print(f"[warn] Failed to write runner log after retries: {self.runner_log}", flush=True)

    def ensure_dirs(self) -> None:
        self.cfg.state_root.mkdir(parents=True, exist_ok=True)
        self.runs_dir.mkdir(parents=True, exist_ok=True)
        self.state_dir.mkdir(parents=True, exist_ok=True)

    @staticmethod
    def ensure_file_if_missing(path: Path, content: str) -> None:
        if not path.exists():
            path.write_text(content, encoding="utf-8")

    def ensure_control_files(self) -> None:
        self.ensure_file_if_missing(self.rules_path, RULES_TEMPLATE)
        self.ensure_file_if_missing(self.goals_path, GOALS_TEMPLATE)
        self.ensure_file_if_missing(self.plan_path, PLAN_TEMPLATE)
        self.ensure_file_if_missing(self.status_path, STATUS_TEMPLATE)
        self.ensure_file_if_missing(self.history_path, HISTORY_TEMPLATE)

    def configure_proxy_if_needed(self) -> None:
        if self.cfg.disable_auto_proxy:
            self.log("AutoProxy disabled; using current proxy environment values.")
            return
        try:
            parsed = urlparse(self.cfg.proxy_url)
            host = parsed.hostname
            port = parsed.port
            if not host or not port:
                self.log(f"Proxy URL has no valid host/port: {self.cfg.proxy_url}")
                return
            with socket.create_connection((host, port), timeout=1.2):
                pass
            os.environ["HTTP_PROXY"] = self.cfg.proxy_url
            os.environ["HTTPS_PROXY"] = self.cfg.proxy_url
            os.environ["ALL_PROXY"] = self.cfg.proxy_url
            self.log(f"Proxy enabled for daemon and child runs: {self.cfg.proxy_url}")
        except Exception as exc:
            self.log(f"Proxy not reachable or invalid '{self.cfg.proxy_url}': {exc}")

    def acquire_lock(self) -> bool:
        try:
            self._lock_fd = os.open(self.lock_path, os.O_CREAT | os.O_EXCL | os.O_WRONLY)
            os.write(self._lock_fd, str(os.getpid()).encode("utf-8"))
            return True
        except FileExistsError:
            print(f"[info] codex-24x7 daemon already running (lock exists at {self.lock_path}).", flush=True)
            return False

    def release_lock(self) -> None:
        if self._lock_fd is not None:
            os.close(self._lock_fd)
        try:
            self.lock_path.unlink(missing_ok=True)
        except Exception:
            pass

    def run_command(self, args: list[str], cwd: Path | None = None) -> list[str]:
        try:
            proc = subprocess.run(args, cwd=str(cwd or self.cfg.repo_root), capture_output=True, text=True)
        except FileNotFoundError:
            cmd = args[0] if args else "<unknown>"
            if cmd not in self._missing_command_warnings:
                self._missing_command_warnings.add(cmd)
                self.log(f"Command not found: {cmd}; continuing with fallback behavior where available.")
            return []
        out = proc.stdout.splitlines() if proc.stdout else []
        if proc.returncode not in (0, 1):
            return []
        return out

    def iter_repo_files(self) -> list[Path]:
        excludes = {".git", "node_modules", "dist", "coverage"}
        files: list[Path] = []
        for root, dirs, filenames in os.walk(self.cfg.repo_root, topdown=True):
            dirs[:] = [d for d in dirs if d not in excludes]
            root_path = Path(root)
            for name in filenames:
                files.append(root_path / name)
        return files

    @staticmethod
    def is_binary_file(path: Path) -> bool:
        try:
            with path.open("rb") as f:
                sample = f.read(2048)
            return b"\0" in sample
        except OSError:
            return True

    def get_todo_entries(self) -> list[dict[str, Any]]:
        args = [
            "rg", "-n", "--hidden",
            "--glob", "!.git/**",
            "--glob", "!**/node_modules/**",
            "--glob", "!**/dist/**",
            "--glob", "!**/coverage/**",
            "(TODO|FIXME|XXX|HACK|BUG)", ".",
        ]
        entries: list[dict[str, Any]] = []
        if shutil.which("rg"):
            lines = self.run_command(args)
            pat = re.compile(r"^(.*?):(\d+):(.*)$")
            for line in lines[: self.cfg.max_todo_lines]:
                m = pat.match(line)
                if not m:
                    continue
                path, lineno, text = m.group(1), int(m.group(2)), m.group(3).strip()
                entries.append({"key": f"{path}:{lineno}", "path": path, "line": lineno, "text": text})
            return entries

        self.log("ripgrep (rg) not found; using built-in TODO scanner.")
        todo_pat = re.compile(r"(TODO|FIXME|XXX|HACK|BUG)")
        for path in self.iter_repo_files():
            if len(entries) >= self.cfg.max_todo_lines:
                break
            if self.is_binary_file(path):
                continue
            try:
                rel = path.relative_to(self.cfg.repo_root).as_posix()
                for lineno, raw in enumerate(path.read_text(encoding="utf-8", errors="ignore").splitlines(), start=1):
                    if not todo_pat.search(raw):
                        continue
                    text = raw.strip()
                    entries.append({"key": f"{rel}:{lineno}", "path": rel, "line": lineno, "text": text})
                    if len(entries) >= self.cfg.max_todo_lines:
                        break
            except OSError:
                continue
        return entries

    def get_plan_files(self) -> list[str]:
        if shutil.which("rg"):
            files = self.run_command(["rg", "--files", "--hidden", "--glob", "!.git/**"])
        else:
            files = [p.relative_to(self.cfg.repo_root).as_posix() for p in self.iter_repo_files()]
        pat = re.compile(r"(plan|roadmap|status|implementation|checklist|backlog|todo)", re.IGNORECASE)
        return [f for f in files if pat.search(f)][: self.cfg.max_plan_files]

    def load_previous_todo_map(self) -> dict[str, dict[str, Any]]:
        if not self.todo_current_path.exists():
            return {}
        try:
            obj = json.loads(self.todo_current_path.read_text(encoding="utf-8"))
            entries = obj.get("entries", [])
            return {e["key"]: e for e in entries if isinstance(e, dict) and "key" in e}
        except Exception:
            self.log("Warning: failed to parse previous todo snapshot; starting fresh.")
            return {}

    def build_run_context(self, todos: list[dict[str, Any]], previous: dict[str, dict[str, Any]], plan_files: list[str]) -> dict[str, Any]:
        current = {t["key"]: t for t in todos}
        new_items = [t for t in todos if t["key"] not in previous]
        resolved_items = [previous[k] for k in previous if k not in current]
        snapshot = {"generatedAt": datetime.now(timezone.utc).isoformat(), "total": len(todos), "entries": todos}
        self.todo_current_path.write_text(json.dumps(snapshot, ensure_ascii=False, indent=2), encoding="utf-8")
        hist = {
            "generatedAt": datetime.now(timezone.utc).isoformat(),
            "openCount": len(todos),
            "newCount": len(new_items),
            "resolvedCount": len(resolved_items),
            "newKeys": [x["key"] for x in new_items[:30]],
            "resolvedKeys": [x["key"] for x in resolved_items[:30]],
        }
        with self.todo_history_path.open("a", encoding="utf-8") as f:
            f.write(json.dumps(hist, ensure_ascii=False) + "\n")

        def to_lines(items: list[dict[str, Any]], none_msg: str, limit: int) -> str:
            if not items:
                return none_msg
            return "\n".join(f"{x['path']}:{x['line']}: {x['text']}" for x in items[:limit])

        previews: list[str] = []
        if not plan_files:
            previews.append("No plan/status/checklist files detected.")
        for rel in plan_files:
            p = self.cfg.repo_root / rel
            previews.append(f"### {rel}")
            if p.exists():
                lines = p.read_text(encoding="utf-8", errors="ignore").splitlines()[: self.cfg.max_file_preview_lines]
                previews.append("\n".join(lines) if lines else "(empty or unreadable)")
            else:
                previews.append("(missing)")
            previews.append("")

        return {
            "openCount": len(todos),
            "newCount": len(new_items),
            "resolvedCount": len(resolved_items),
            "openTodoText": to_lines(todos, "No open TODO/FIXME lines found.", 250),
            "newTodoText": to_lines(new_items, "No newly discovered TODO/FIXME items.", 80),
            "resolvedTodoText": to_lines(resolved_items, "No TODO/FIXME items resolved since last run.", 80),
            "planPreviewText": "\n".join(previews),
        }

    def build_prompt(self, ctx: dict[str, Any]) -> str:
        now = datetime.now().strftime("%Y-%m-%d %H:%M:%S %z")
        return f"""You are Codex in continuous 24x7 framework-maintenance mode for this repository.

Timestamp: {now}
Repository: {self.cfg.repo_root}

Follow these instructions:
1) Read and follow `ops/codex-24x7/RULES.md`, `ops/codex-24x7/GOALS.md`, and `ops/codex-24x7/PLAN.md`.
2) Execute one focused high-impact batch of work now.
3) Prioritize production readiness: tests, reliability fixes, TODO/FIXME debt, and example parity.
4) Run relevant tests for files you changed; include framework tests and ecommerce tests when applicable.
5) Update both files before finishing:
   - `ops/codex-24x7/STATUS.md` (Completed this run / Remaining work / Next run plan)
   - `ops/codex-24x7/HISTORY.md` (append concise run log with what changed and what remains)
6) If blocked on one task, move to the next highest-value unblocked task.
7) Prefer using `helper-projects/meta-shell` patterns/assets when reworking UI/admin ergonomics.

Coverage mandate for ongoing runs:
- Admin features
- ORM + schema + migrations
- API + server behavior
- Ecommerce feature parity and tests

State from orchestrator:
- Open TODO count: {ctx["openCount"]}
- New TODO count: {ctx["newCount"]}
- Resolved TODO count (since last run): {ctx["resolvedCount"]}

Open TODO/FIXME items (truncated):
{ctx["openTodoText"]}

New TODO/FIXME items (truncated):
{ctx["newTodoText"]}

Recently resolved TODO/FIXME items (truncated):
{ctx["resolvedTodoText"]}

Plan/status/checklist file previews (truncated):
{ctx["planPreviewText"]}
"""

    def load_active_run(self) -> dict[str, Any] | None:
        if not self.active_run_path.exists():
            return None
        try:
            return json.loads(self.active_run_path.read_text(encoding="utf-8"))
        except Exception:
            self.log("Warning: active-run state is invalid; clearing it.")
            self.clear_active_run()
            return None

    def set_active_run(self, pid: int, stamp: str, run_log: Path, last_message: Path) -> None:
        payload = {
            "pid": pid,
            "runStamp": stamp,
            "mode": self.cfg.mode,
            "startedAt": datetime.now(timezone.utc).isoformat(),
            "maxRunMinutes": self.cfg.max_run_minutes,
            "runLogPath": str(run_log),
            "lastMessagePath": str(last_message),
        }
        self.active_run_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

    def clear_active_run(self) -> None:
        self.active_run_path.unlink(missing_ok=True)

    @staticmethod
    def pid_exists(pid: int) -> bool:
        if pid <= 0:
            return False
        try:
            os.kill(pid, 0)
            return True
        except OSError:
            return False

    def get_active_decision(self) -> dict[str, Any]:
        active = self.load_active_run()
        if not active:
            return {"shouldSpawn": True, "reason": "no-active-run"}
        pid = int(active.get("pid", -1))
        if not self.pid_exists(pid):
            self.log(f"Active run PID {pid} not found. Clearing stale active-run state.")
            self.clear_active_run()
            return {"shouldSpawn": True, "reason": "stale-active-run"}
        started = datetime.fromisoformat(active.get("startedAt"))
        elapsed = datetime.now(timezone.utc) - started.astimezone(timezone.utc)
        if elapsed.total_seconds() >= self.cfg.max_run_minutes * 60:
            self.log(f"Active run PID {pid} exceeded timeout ({elapsed.total_seconds()/60:.2f} min >= {self.cfg.max_run_minutes} min). Killing process.")
            subprocess.run(["taskkill", "/PID", str(pid), "/T", "/F"], capture_output=True, text=True)
            self.log(f"Killed stuck run PID {pid}.")
            with self.history_path.open("a", encoding="utf-8") as f:
                f.write(
                    f"\n## {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"
                    f"- Runner action: timed out previous run PID {pid} after {elapsed.total_seconds()/60:.2f} minutes\n"
                    f"- Active run file was reset\n"
                )
            self.clear_active_run()
            return {"shouldSpawn": True, "reason": "timed-out-and-reset"}
        return {"shouldSpawn": False, "reason": f"active-run-pid-{pid}", "elapsedMin": elapsed.total_seconds() / 60}

    @staticmethod
    def resolve_codex_launcher() -> list[str]:
        for candidate in ("codex.exe", "codex.cmd", "codex.bat", "codex"):
            p = shutil.which(candidate)
            if p:
                ext = Path(p).suffix.lower()
                if ext == ".ps1":
                    ps = shutil.which("powershell") or "powershell"
                    return [ps, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", p]
                if ext in (".cmd", ".bat"):
                    return ["cmd.exe", "/c", p]
                return [p]
        ps = shutil.which("powershell") or "powershell"
        probe = subprocess.run(
            [ps, "-NoProfile", "-Command", "(Get-Command codex -ErrorAction Stop).Source"],
            capture_output=True,
            text=True,
        )
        source = probe.stdout.strip() if probe.returncode == 0 else ""
        if not source:
            raise RuntimeError("Unable to resolve codex command.")
        ext = Path(source).suffix.lower()
        if ext == ".ps1":
            return [ps, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", source]
        if ext in (".cmd", ".bat"):
            return ["cmd.exe", "/c", source]
        return [source]

    def invoke_codex_run(self, prompt: str, stamp: str, run_log: Path, last_message: Path) -> int:
        prompt_path = self.runs_dir / f"prompt-{stamp}.txt"
        stdout_path = self.runs_dir / f"stdout-{stamp}.log"
        stderr_path = self.runs_dir / f"stderr-{stamp}.log"
        prompt_path.write_text(prompt, encoding="utf-8")

        codex_args = ["exec", "-C", str(self.cfg.repo_root), "--skip-git-repo-check", "--output-last-message", str(last_message)]
        if self.cfg.mode == "safe":
            codex_args += ["-s", "workspace-write", "--full-auto"]
        elif self.cfg.mode == "approve":
            codex_args += ["-s", "workspace-write"]
        elif self.cfg.mode == "danger":
            codex_args += ["--dangerously-bypass-approvals-and-sandbox"]
        codex_args += ["-"]
        launcher = self.resolve_codex_launcher()
        cmd = launcher + codex_args

        with run_log.open("a", encoding="utf-8") as f:
            f.write(f"### codex command: {' '.join(cmd)}\n")
        self.log(f"Invoking codex ({self.cfg.mode} mode).")

        with prompt_path.open("rb") as stdin_f, stdout_path.open("wb") as stdout_f, stderr_path.open("wb") as stderr_f:
            proc = subprocess.Popen(cmd, cwd=str(self.cfg.repo_root), stdin=stdin_f, stdout=stdout_f, stderr=stderr_f)
            self.set_active_run(proc.pid, stamp, run_log, last_message)
            try:
                proc.wait(timeout=self.cfg.max_run_minutes * 60)
            except subprocess.TimeoutExpired:
                proc.kill()
                self.log(f"Killed run PID {proc.pid} after timeout of {self.cfg.max_run_minutes} minutes.")
                with run_log.open("a", encoding="utf-8") as f:
                    f.write(f"Run timed out after {self.cfg.max_run_minutes} minutes.\n")
                return 124
            finally:
                self.clear_active_run()

        with run_log.open("a", encoding="utf-8") as f:
            if stdout_path.exists():
                f.write("### stdout\n")
                f.write(stdout_path.read_text(encoding="utf-8", errors="ignore"))
            if stderr_path.exists():
                f.write("### stderr\n")
                f.write(stderr_path.read_text(encoding="utf-8", errors="ignore"))
        return int(proc.returncode or 0)

    def daemon_loop(self) -> int:
        self.ensure_dirs()
        self.ensure_control_files()
        self.configure_proxy_if_needed()
        if not self.acquire_lock():
            return 0
        self.log(
            f"Daemon started. Interval={self.cfg.interval_minutes} minute(s), "
            f"Mode={self.cfg.mode}, MaxRunMinutes={self.cfg.max_run_minutes}, Repo={self.cfg.repo_root}"
        )
        try:
            while True:
                if self.stop_path.exists():
                    self.log(f"STOP file detected at {self.stop_path}. Exiting daemon loop.")
                    break
                stamp = datetime.now().strftime("%Y%m%d-%H%M%S")
                run_log = self.runs_dir / f"run-{stamp}.log"
                last_message = self.runs_dir / f"last-message-{stamp}.txt"
                self.log("Refreshing TODO/plan snapshot for this tick.")
                todos = self.get_todo_entries()
                previous = self.load_previous_todo_map()
                plan_files = self.get_plan_files()
                ctx = self.build_run_context(todos, previous, plan_files)
                self.log(f"TODO snapshot updated: open={ctx['openCount']}, new={ctx['newCount']}, resolved={ctx['resolvedCount']}")
                decision = self.get_active_decision()
                if not decision.get("shouldSpawn"):
                    self.log(
                        "Managed run already active "
                        f"({decision.get('reason')}, elapsed={decision.get('elapsedMin', 0):.2f}m). Skipping new agent spawn."
                    )
                else:
                    self.log(f"No managed active run ({decision.get('reason')}); starting a new run.")
                    prompt = self.build_prompt(ctx)
                    self.last_prompt_path.write_text(prompt, encoding="utf-8")
                    exit_code = self.invoke_codex_run(prompt, stamp, run_log, last_message)
                    self.log(f"Run finished with exit code {exit_code}")
                    summary = "(no message file)"
                    if last_message.exists():
                        raw = last_message.read_text(encoding="utf-8", errors="ignore").strip()
                        if raw:
                            summary = re.sub(r"\s+", " ", raw)[:500]
                            if len(raw) > 500:
                                summary += "..."
                    with self.history_path.open("a", encoding="utf-8") as f:
                        f.write(
                            f"\n## {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n"
                            f"- Exit code: {exit_code}\n"
                            f"- TODO snapshot: open={ctx['openCount']}, new={ctx['newCount']}, resolved={ctx['resolvedCount']}\n"
                            f"- Last message summary: {summary}\n"
                            f"- Run log: runs/run-{stamp}.log\n"
                        )
                if self.cfg.one_shot:
                    self.log("OneShot mode enabled; exiting after one loop.")
                    break
                self.log(f"Sleeping for {self.cfg.interval_minutes} minute(s).")
                time.sleep(self.cfg.interval_minutes * 60)
        finally:
            self.release_lock()
            self.log("Daemon stopped.")
        return 0


def parse_args() -> Config:
    parser = argparse.ArgumentParser(description="Codex 24x7 runner daemon")
    repo_default = Path(__file__).resolve().parents[1]
    parser.add_argument("--repo-root", default=str(repo_default))
    parser.add_argument("--interval-minutes", type=int, default=10)
    parser.add_argument("--mode", choices=["safe", "approve", "danger"], default="safe")
    parser.add_argument("--one-shot", action="store_true")
    parser.add_argument("--state-root", default="")
    parser.add_argument("--max-todo-lines", type=int, default=500)
    parser.add_argument("--max-plan-files", type=int, default=25)
    parser.add_argument("--max-file-preview-lines", type=int, default=80)
    parser.add_argument("--max-run-minutes", type=int, default=45)
    parser.add_argument("--disable-auto-proxy", action="store_true")
    parser.add_argument("--proxy-url", default="http://127.0.0.1:10808")
    ns = parser.parse_args()
    repo_root = Path(ns.repo_root).resolve()
    state_root = Path(ns.state_root).resolve() if ns.state_root else repo_root / "ops" / "codex-24x7"
    return Config(
        repo_root=repo_root,
        interval_minutes=max(1, ns.interval_minutes),
        mode=ns.mode,
        one_shot=ns.one_shot,
        state_root=state_root,
        max_todo_lines=ns.max_todo_lines,
        max_plan_files=ns.max_plan_files,
        max_file_preview_lines=ns.max_file_preview_lines,
        max_run_minutes=max(1, ns.max_run_minutes),
        disable_auto_proxy=ns.disable_auto_proxy,
        proxy_url=ns.proxy_url,
    )


def main() -> int:
    cfg = parse_args()
    runner = Runner(cfg)
    return runner.daemon_loop()


if __name__ == "__main__":
    sys.exit(main())
