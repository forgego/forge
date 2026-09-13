TASK G1-<PART>: fix pre-existing staticcheck findings (Go backend, repo hygiene).

Working directory: /home/hamid/Other/projects/foreit-wt/staticcheck/forge
(git worktree on branch fix/staticcheck-findings, based on master.)

=== WHY ===
CI's "Static Analysis" job runs `go install honnef.co/go/tools/cmd/staticcheck@latest` then
`staticcheck ./...` from forge/. A newer staticcheck release now reports ~75 findings that already exist on
master, so EVERY open PR fails Static Analysis. The full list is in
/home/hamid/Other/projects/foreit/docs/design/tasks/staticcheck-findings.txt (read-only reference).

=== YOUR SCOPE ===
PART is given at the top of the prompt as a list of package directories. Fix ONLY findings whose file path
starts with one of those directories. Do NOT touch any other file.

First install staticcheck locally (allowed): `go install honnef.co/go/tools/cmd/staticcheck@latest`
then run `$(go env GOPATH)/bin/staticcheck ./<dir>/...` for each dir in scope to get the live list.

=== HOW TO FIX EACH RULE (behaviour must not change) ===
- U1000 (unused): delete the unused identifier. If it is an unused test helper/mock that looks intentionally
  kept for future tests, still delete it (it is dead code). Never delete exported API.
- ST1005 (capitalized/punctuated error strings): lowercase the first letter / drop trailing punctuation.
  BEFORE changing, grep the repo (from forge/ and ../examples, ../tests) for tests or code comparing that exact
  message string; update those comparisons in the same package only if they are in your scope, otherwise
  leave that finding alone and report it.
- S1040 (type assertion to the same type), S1039 (unnecessary fmt.Sprint), S1000/S1002/S1008/S1011/S1016/
  S1025: apply the mechanical simplification staticcheck suggests.
- SA1029 (built-in type string as context key): introduce an unexported `type ctxKey string` (or struct{} key)
  in that package and use it for BOTH WithValue and Value lookups of that key. Grep the whole package for
  every use of the same key string first. If the key is read from ANOTHER package with a raw string, do not
  change it — report it instead.
- SA1019 (deprecated API): switch to the replacement named in the message only if it is a drop-in; otherwise
  add `//lint:ignore SA1019 <reason>` on that line and report it.
- SA4000/SA4006/SA4010/SA1026: fix the actual bug the rule describes (identical expressions, value never used,
  append result unused, unmarshal into non-pointer). Explain each in the report.

=== VERIFY (from the working directory) ===
  $(go env GOPATH)/bin/staticcheck ./<each dir in scope>/...     (prints nothing for your scope)
  gofmt -l <each dir in scope>                                   (your edited files must not appear)
  go vet ./...
  go test ./<each dir in scope>/...
  go build ./... && (cd ../examples/ecommerce && go build ./...)

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- No behaviour changes except the genuine bugs flagged by SA4xxx/SA1026, each explained.
- Final report: per-finding what you did (file:line → action), anything left and why, staticcheck output.
