# Forge Roadmap

> Vision: **be the Django of Go**: a type-safe, full-stack Go framework
> with batteries included (ORM, migrations, admin, API, auth).
> This file tracks scheduled work. Known debts live in [TECH-DEBT.md](TECH-DEBT.md);
> how changes are judged lives in [REVIEWING.md](REVIEWING.md).

## Release readiness

A checked box here records intent, not proof. A release is ready only when
every one of the following holds, each backed by recorded evidence (command,
commit, and pass/fail/skip counts):

1. **Framework suite.** `go test -race ./...` in `forge/` passes, and the
   database tests run against PostgreSQL. A skip caused by missing database
   infrastructure counts as a gap, not a pass.
2. **Independent consumer.** A freshly generated project, outside this
   repository's examples, compiles and serves create, read, update, delete,
   validation failures and permission denials without edits to generated files.
3. **Schema evolution.** Migration apply, status, integrity verification, and
   recovery from interrupted or altered history behave as documented on every
   supported database.
4. **Protection.** Unauthorized mutations leave data unchanged. Fields that the
   output contract does not declare never appear in responses. Errors never
   expose internal details.
5. **Example application.** `examples/ecommerce` builds and its tests pass
   (`go test ./...`). It demonstrates the framework; it is not itself the
   readiness proof.
6. **Upgrade and limits.** Upgrading from the previous release follows the
   published guide with data preserved. Every capability that is not supported
   is labeled as such and fails explicitly when used.

## Framework

### Now

- [ ] Responses are rendered through the configured serializer's declared
      fields, not raw model tags.
- [ ] Persistence and hook errors map to stable categories (400 / 404 / 500)
      without leaking internal messages.
- [ ] Viewset data-access configuration is validated at registration, not
      discovered by reflection at request time.
- [ ] External consumer journey test against PostgreSQL (CRUD, permission
      denial, validation, migration recovery).
- [ ] TODO/FIXME burn-down across admin, ORM/schema/migrations,
      API/server reliability edges.
- [x] Password-hash guard (`strings.HasPrefix(hash,"$2a$")`)
      to avoid double-hashing in `forge/identity/utils`.

### Next

- [x] REST auto-generation: ViewSets + Serializers + routes
      generated from models via `forge generate --api` (`docs/PRD.md` FR-API/OpenAPI).
- [x] CLI commands aligned with PRD Appendix A (`forge shell`, `forge test`,
      `forge check`, `forge createsuperuser`, and `forge add app/model/api`).
- [x] Migration recovery/checksum verification (`forge migrate recover`
      fail-loud + force/recover/verify implemented).
- [ ] Caching: query/instance cache, Redis + in-memory backends.
- [ ] Query power: window functions, full-text search, raw SQL escape hatch.
- [ ] Background tasks (queues/workers).
- [ ] Observability: Prometheus metrics, OpenTelemetry health/tracing.

### Later

- [ ] GraphQL + mobile API surface.
- [ ] Kubernetes deployment, service mesh, RabbitMQ/Kafka,
      microservices split.
- [ ] Multi-tenant, subscriptions, social auth.
- [ ] Data migrations, squashing UI, migration testing framework,
      perf analysis, multi-DB, migration templates.
- [ ] Read-replicas, CDN, distributed tracing/metrics/alerting
      (today: local pooling/indexes/pagination only).

## Example application (`examples/ecommerce`)

- [x] Unit/integration/E2E suites for non-category modules
      (models/services, API/DB integration, purchase/return/support flows verified in `main_test.go`).
- [x] Authenticated integration coverage + object-specific 403 paths
      where SQLite-backed (verified across Payment, Order, Warehouse, Coupon in `main_test.go`).
- [x] `make seed` runs a seed script again:
      recreated in `examples/ecommerce/scripts/seed.go`.
- [x] Production hardening: auto-generate ephemeral secrets and warn for
      placeholder keys in `ensureSecrets()`, cleared static secrets in `config.yaml`.
- [x] `Mark Delivered` sets `delivered_at`; required-field
      and validation saves return 400 `validation_error`, not 500s.
- [ ] Auth middleware, caching layer, structured logging,
      dashboard widgets, email templates, file uploads.
- [ ] Stripe/PayPal + shipping-carrier APIs.
- [ ] WebSocket notifications, customer storefront,
      analytics dashboards, Elasticsearch.
- [ ] Marketplace, affiliate, recommendation engine.

## Contributor process

- [x] Resolved "pre-built binaries are coming soon": updated install
      docs to point to `go install` and GitHub tagged releases.
- [ ] Acceptance is judged by someone other than the implementer
      (see [REVIEWING.md](REVIEWING.md#acceptance-review)).
- [ ] CI fails on unresolved relative links and referenced repository paths
      in public documentation and contributor skills.
- [ ] Analyzer corpus fixtures/goldens/benchmarks; RAG/evaluator ranking
      fixtures; harness-compatibility evidence (Claude/Codex/OpenCode/Zed/dmux).
