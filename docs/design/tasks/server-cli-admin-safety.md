Your workspace is /home/hamid/Other/projects/foreit-wt/wave0-rest. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# Superuser password echo, admin history race, fake HTML/SQL "sanitizers"

Module root: /home/hamid/Other/projects/foreit-wt/wave0-rest/forge (run go commands from there).
Only edit: forge/cli/commands/admin/createsuperuser.go, forge/admin/history_manager.go, forge/server/security.go, their _test.go files, and forge/go.mod / go.sum if `golang.org/x/term` must be added. Another agent is editing forge/api at the same time: if a build error appears in forge/api, ignore it.

## Principles (mandatory)
- Reproduce first where testable: write the test FIRST, run it, keep the failing output, then fix.
- Smallest correct change; gofmt. Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state.
- Naming after behaviour, never after tickets.

## 1. Superuser password is typed with echo (verified)
forge/cli/commands/admin/createsuperuser.go:126 reads the password with `reader.ReadString('\n')`, so it is shown on screen.
Fix: when stdin is a terminal (`term.IsTerminal(int(os.Stdin.Fd()))`), read with `term.ReadPassword` and print a newline afterwards; otherwise (piped input, tests) keep reading a line from the reader. Ask for confirmation the same way if the command already does. Check go.mod first: `golang.org/x/term` may already be an indirect dependency (use the version already in go.sum if present; no network may be available, so try `GOFLAGS=-mod=mod go build` and report if it cannot be resolved).

## 2. Admin history lazy-init data race (verified)
forge/admin/history_manager.go:16-21 `getMem()` sets `m.mem` without synchronisation on first use.
Fix: initialise `mem` in the constructor(s) of `HistoryManager` and keep `getMem` as a plain getter, or guard it with `sync.Once` if a zero-value `HistoryManager{}` is used anywhere (grep for it). Test: a test that calls the recording and reading methods from 20 goroutines on a fresh manager; it must pass under `-race` and fail under `-race` before the fix.

## 3. Regex HTML sanitizer and SQL keyword blacklist (verified, no callers)
forge/server/security.go: `XSS.SanitizeHTML` (~260), `XSS.SanitizeHTMLStrict` (~310), `XSS.SanitizeInput` (~338) strip tags with regexes that miss unquoted event handlers (`<img onerror=alert(1) src=x>`), and `SQLInjection.ValidateInput` (~129) / `EnsureParameterized` (~169) are keyword blacklists that are not a defence. Nothing in the repo calls them (confirm with grep across /home/hamid/Other/projects/foreit-wt/wave0-rest, including examples and cli templates).
Fix: delete those five methods and anything that only they use (regex vars, helpers). Keep `EscapeHTML`, `SanitizeIdentifier`, `ContentSecurityPolicy`, `LogQuery` and the constructors if anything still uses them; if a whole type ends up unused by any code, delete it too. Delete tests that only cover the removed methods. If grep finds a caller, STOP on this item and report it.

## Acceptance (from the module root)
    gofmt -l cli admin server                        # empty
    go vet ./cli/... ./admin/... ./server/...
    staticcheck ./cli/... ./admin/... ./server/...   # no output
    go test -race ./cli/... ./admin/... ./server/... -count=1
    go build ./... && (cd ../examples/ecommerce && go build ./...)

Report (max 25 lines): failing output before fixes, files changed, removed identifiers, each command result.
