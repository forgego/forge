# Codex 24x7 Rules

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
