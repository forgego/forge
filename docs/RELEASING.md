# Releasing Forge

How a Forge release is cut, what evidence it needs, and how a bad release is
withdrawn. The user-facing promises a release must keep (support tiers,
compatibility, breaking-change policy) are in the
[support contract](../docs-site/docs/status.md); this file is the maintainer
procedure that enforces them.

## Versions and tags

- The repository is **one Go module**, `github.com/forgego/forge`, at the
  repository root. A release is a tag `vMAJOR.MINOR.PATCH` on a `master`
  commit, with no path prefix. `tests/` and `examples/ecommerce/` are separate
  modules that use `replace ../..`; they are never tagged or published.
- Use annotated tags (`git tag -a`). The existing tags are lightweight.
- A pushed tag is permanent. The module proxy and `sum.golang.org` record its
  content the first time anyone fetches it, so never move, delete or reuse a
  tag; publish a new version instead.
- v1.0.0 and v1.0.1 were tagged by mistake and are retracted in `go.mod`.
  They cannot be reused, so the first real v1 release must be numbered above
  v1.0.1 (v1.1.0 avoids confusion with the retracted tags). Until then, Go
  derives pseudo-versions for untagged commits from the highest tag, so a
  commit on `master` resolves as `v1.0.2-0.<date>-<hash>`. That is expected;
  `@latest` still resolves to the newest v0.x release because retracted
  versions and pseudo-versions are excluded from it.
- Choosing the number: see the stability policy in the support contract. On
  v0.x, a breaking change needs a new minor version; fixes go in a patch.

## Release checklist

Do these in a release pull request, so the tag lands on a reviewed commit.

1. **Version constant.** Set `fallbackVersion` in `cli/core/version.go` to the
   new tag. Binaries installed with `go install ...@vX.Y.Z` report the version
   from their build information; the constant is only used by builds that
   carry none (a local `go build` in a checkout) and must not lag the release.
2. **Changelog.** In [the changelog](../docs-site/docs/changelog.md), rename
   `Unreleased` to `vX.Y.Z (YYYY-MM-DD)` and start a new empty `Unreleased`
   section. Every entry is under **Breaking**, **Added**, **Fixed** or
   **Known issues**, and cites the pull request. Build it from
   `git log --oneline vPREVIOUS..HEAD`, not from memory.
3. **Upgrade guide.** A release with anything under **Breaking** includes an
   upgrade guide in the changelog entry (template below).
4. **Support contract.** If a capability changed tier, or a Go or PostgreSQL
   version was added or dropped, update the
   [support contract](../docs-site/docs/status.md) in the same pull request.
5. **Retractions.** If the release withdraws an earlier version, add the
   `retract` directive to `go.mod` (see [Retracting](#retracting-a-release)).
6. **Evidence and review.** Attach the release evidence and get the
   independent acceptance review described below. The upgrade rehearsal is
   required for every minor and major release and for the first v1.
7. **Tag.** After the pull request is merged:

   ```bash
   git fetch origin
   git tag -a vX.Y.Z -m "Forge vX.Y.Z" <merge-commit>
   git push origin vX.Y.Z
   ```

8. **Verify the published module.** The tag push starts
   `.github/workflows/install-smoke.yml`, which installs the CLI from the
   public proxy, checks `forge version`, creates a project, builds and vets it,
   runs it against PostgreSQL and probes `/health`. The release is not done
   until that run is green. You can also fetch it by hand:

   ```bash
   GOPROXY=https://proxy.golang.org go list -m github.com/forgego/forge@vX.Y.Z
   ```

9. **GitHub release.** Publish a release for the tag with the changelog entry
   and links to the evidence.

## Release evidence

The `release-gate` job in `.github/workflows/test.yml` is the release
evidence. It starts a `postgres:15` service, sets `FORGE_REQUIRE_DB=1` so a
missing database fails tests instead of skipping them, and runs every package
of the root module and of the `tests` module serially:

```bash
go test -count=1 -p 1 -json ./... | testreport --out forge-report.txt \
  --require-no-skip '^github.com/.*/(identity|internal/testutils)'
(cd tests && go test -count=1 -p 1 -json ./... | testreport --out tests-report.txt \
  --require-no-skip '^github.com/forgego/forge/tests/(integration|pkg_migrations|e2e)' \
  --allow-skip 'tests/pkg_migrations$ ^TestMigrationApplySQLite$')
```

`internal/tools/testreport` prints pass, fail, skip and package-failure counts
per package and in total, lists failed and skipped tests, and exits non-zero
if anything failed or a package matching `--require-no-skip` skipped a test
that `--allow-skip` does not exempt. Both reports are uploaded as the
`release-test-reports` artifact. The job's `no_db` dispatch input points it at
an unreachable database, to prove the gate fails without one.

The job only runs when Go files or `test.yml` change. For a release commit
that changes neither, start it with **Run workflow** on the Tests workflow.

Record in the release pull request:

| Evidence | Where it comes from |
| --- | --- |
| Commit SHA and workflow run URLs | The Tests run for the release commit |
| Totals and every skip from `forge-report.txt` and `tests-report.txt` | `release-test-reports` artifact |
| Unit tests on Go 1.26.8 and 1.27.1, lint, build, frontend, CLI E2E, govulncheck | Tests workflow jobs |
| Ecommerce example build and tests | Ecommerce Sample workflow (example evidence, not framework evidence) |
| CodeQL | CodeQL workflow |
| Upgrade rehearsal log | Below |
| Install smoke run for the tag | Install Smoke workflow, after tagging |

A skip is a gap, not a pass. List each skip with its reason; anything not
explained blocks the release.

To reproduce the gate locally against your own PostgreSQL, build the reporter
and point the tests at a scratch database:

```bash
go build -o /tmp/testreport ./internal/tools/testreport
FORGE_REQUIRE_DB=1 \
FORGE_TEST_DATABASE_URL='postgres://USER:PASS@127.0.0.1:5432/SCRATCH_DB?sslmode=disable' \
go test -count=1 -p 1 -json ./... | /tmp/testreport --out forge-report.txt \
  --require-no-skip '^github.com/.*/(identity|internal/testutils)'
```

The tests truncate tables, so never point them at a database you care about.
Run as a non-root user: `TestChecksumBaseline_CleanRollbackErrorReconciles`
makes a file unreadable with `chmod 000`, which root can still read.

## Independent review

Follow the [review standard](REVIEWING.md), including its acceptance review:
the person or model that implemented a capability does not accept it. For a
release, the reviewer starts from the changelog and the diff since the
previous tag, exercises each changed capability through its public interface,
records at least one attempted counterexample per capability with the outcome,
and checks that the support contract still matches the evidence. The review
record goes in the release pull request.

## Upgrade rehearsal

Rehearse the upgrade a user will do, on a project created by the previous
release, with data in the database. Run it before tagging, against the
release candidate commit.

**1. Consumer on the previous release.**

```bash
go install github.com/forgego/forge/cmd/forge@vPREVIOUS
forge new rehearsal --template simple --database postgres --docker=false
cd rehearsal
forge add app blog --example
forge generate --models ./app/blog --output ./app/blog
go mod tidy && go build ./...
forge makemigrations initial --auto --models ./app/blog
forge migrate up                 # against a fresh database
```

Insert a few rows into each model table and note the counts.

**2. Snapshot.**

```bash
pg_dump -Fc -f before.dump "$FORGE_DATABASE_NAME"
forge migrate status
cp -r app/blog /tmp/generated-before
```

**3. Upgrade to the candidate.**

```bash
go install github.com/forgego/forge/cmd/forge@<candidate commit or tag>
go get github.com/forgego/forge@<candidate commit or tag> && go mod tidy
forge generate --models ./app/blog --output ./app/blog
diff -r /tmp/generated-before app/blog     # every difference must be in the changelog
go build ./... && go vet ./...
forge migrate recover --verify              # applied files still match their checksums
forge makemigrations upgrade --auto --models ./app/blog
forge migrate up
forge migrate status
```

For a candidate that is not pushed yet, use
`go mod edit -replace github.com/forgego/forge=/path/to/checkout` instead of
`go get`, and a `forge` binary built from that checkout.

**4. Verify.** Start the application with the production settings from the
[deployment guide](../docs-site/docs/deployment.md), run its smoke-check
script, compare row counts with step 1, and create, update and delete a row
through the admin or API. Then restore `before.dump` into a new database and
start the candidate against it, to prove a user who restores a backup can
still run the new release.

**5. Record** the versions, commands, output and results in the release pull
request.

Findings from rehearsing v0.1.1 to the current `master` (2026-09-28), which
anyone rehearsing should expect until they are fixed:

- Projects created by v0.1.0 and v0.1.1 do not compile (`pattern static: no
  matching files found`); delete the `//go:embed` line, the `staticFiles`
  variable and the `embed` import from `cmd/server/main.go`.
- `forge version` in v0.1.1 prints `v0.1.0`, and `forge new` pins `v0.1.0`.
- Regenerating the example model produced no diff, the checksums verified,
  and the rows were preserved.
- `forge makemigrations --auto` with no model changes wrote an empty
  migration pair in v0.1.1, which `forge migrate up` then rejected with
  `SQL is empty`. Fixed after v0.1.1; delete any empty pair.
- Projects created before graceful shutdown was added keep calling
  `srv.Start()`; they exit on SIGTERM without draining requests until
  `main.go` is changed to `StartWithGracefulShutdown`.

## Upgrade guide template

Copy this into the changelog entry of any release with breaking changes.

```markdown
### Upgrading from vA.B.C

**Who is affected:** <which projects, APIs, commands or configurations>.

1. Update the CLI and the library:
   `go install github.com/forgego/forge/cmd/forge@vX.Y.Z` and
   `go get github.com/forgego/forge@vX.Y.Z && go mod tidy`.
2. Regenerate: `forge generate` (plus `--api` if you use it). Expected changes
   in generated code: <list or "none">.
3. Code changes: <each renamed or removed API, with before and after>.
4. Configuration changes: <keys or environment variables added, renamed or
   removed; new required values>.
5. Migrations: <new migrations Forge needs, bookkeeping changes, or "none">.
   Back up first, then `forge migrate up`.
6. Verify: `forge migrate status`, then your smoke check.

**Rolling back:** <how to return to vA.B.C, including whether the migration
in step 5 has a down migration>.
```

## Retracting a release

When a published version is broken or unsafe:

1. Do not delete or move its tag.
2. Fix forward in a new patch release. In that release's `go.mod`, add the
   bad version to the `retract` block with a one-line reason:

   ```go
   retract (
       vX.Y.Z // Projects created by forge new do not compile.
   )
   ```

   A retraction takes effect only when a later version that contains it is
   published, because Go reads retractions from the latest version's `go.mod`.
3. Add a **Known issues** note to the retracted version's changelog entry and
   say which release fixes it.
4. After tagging the fix, check that
   `go list -m -json -retracted github.com/forgego/forge@vX.Y.Z` shows the
   reason under `Retracted`, and that
   `go list -m github.com/forgego/forge@latest` resolves to the fix.
5. For a vulnerability, follow [SECURITY.md](../SECURITY.md) before
   publishing any details.
