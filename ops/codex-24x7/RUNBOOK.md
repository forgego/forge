# Codex 24x7 Runbook

## Start Manually (single command)
```powershell
python .\scripts\codex_24x7.py --mode safe --interval-minutes 10 --proxy-url "http://127.0.0.1:10808"
```

## Stop
```powershell
New-Item -ItemType File -Path .\ops\codex-24x7\STOP -Force
```

## Restart Cleanly
```powershell
Remove-Item .\ops\codex-24x7\STOP -Force -ErrorAction SilentlyContinue
Remove-Item .\ops\codex-24x7\state\active-run.json -Force -ErrorAction SilentlyContinue
python .\scripts\codex_24x7.py --mode safe --interval-minutes 10 --proxy-url "http://127.0.0.1:10808"
```

## Key Files
- `ops/codex-24x7/runner.log`
- `ops/codex-24x7/STATUS.md`
- `ops/codex-24x7/HISTORY.md`
- `ops/codex-24x7/state/todo-current.json`
- `ops/codex-24x7/state/todo-history.jsonl`
- `ops/codex-24x7/state/active-run.json`
- `ops/codex-24x7/runs/`
