# Forge Roadmap

> Vision: **be the Django of Go** — a type-safe, full-stack Go framework
> with batteries included (ORM, migrations, admin, API, auth).
> Source: `docs/archive/ROADMAP.md`, `IMPLEMENTATION_SUMMARY.md` (ecommerce),
> `ARCHITECTURE.md`, `STATUS.md` (ops runs), `docs_old` triage.

## Definition of done (production-ready)

Per the ops run goals, a release is production-ready when **all three** hold:

1. Framework tests pass (`go test ./...` in `forge/`).
2. Ecommerce example reaches feature parity (all apps registered, working).
3. Ecommerce example tests pass (`go test ./...` in `examples/ecommerce`).

## Now (committed gaps)

- [x] Ecommerce unit/integration/E2E suites for non-category modules
      (models/services, API/DB integration, purchase/return/support flows verified in `main_test.go`).
- [x] Authenticated integration coverage + object-specific 403 paths
      where SQLite-backed (verified across Payment, Order, Warehouse, Coupon in `main_test.go`).
- [ ] TODO/FIXME burn-down across admin, ORM/schema/migrations,
      API/server reliability edges.
- [x] Password-hash guard (`strings.HasPrefix(hash,"$2a$")`)
      to avoid double-hashing in `forge/identity/utils`.
- [ ] Analyzer corpus fixtures/goldens/benchmarks; RAG/evaluator ranking
      fixtures; harness-compatibility evidence (Claude/Codex/OpenCode/Zed/dmux).
- [x] `make seed` references `scripts/seed.go` —
      recreated seed script in `examples/ecommerce/scripts/seed.go`.
- [x] Ecommerce production hardening: auto-generate ephemeral secrets and warn for
      placeholder keys in `ensureSecrets()`, cleared static secrets in `config.yaml`.
- [ ] Ecommerce: auth middleware, caching layer, structured logging,
      dashboard widgets, email templates, file uploads.
- [x] Ecommerce: `Mark Delivered` sets `delivered_at`; required-field
      and validation saves return 400 `validation_error`, not 500s.
- [x] Resolved "pre-built binaries are coming soon": updated install
      docs to point to `go install` and GitHub tagged releases.

## Next (framework features)

- [x] REST auto-generation: ViewSets + Serializers + routes
      generated from models via `forge generate --api` (`docs/PRD.md` FR-API/OpenAPI).
- [x] CLI commands aligned with PRD Appendix A (`forge shell`, `forge test`,
      `forge check`, `forge createsuperuser`, and `forge add app/model/api`).
- [ ] Caching: query/instance cache, Redis + in-memory backends.
- [ ] Query power: window functions, full-text search, raw SQL escape hatch.
- [ ] Background tasks (queues/workers).
- [ ] Observability: Prometheus metrics, OpenTelemetry health/tracing.
- [ ] Ecommerce: Stripe/PayPal + shipping-carrier APIs.
- [ ] Ecommerce: WebSocket notifications, customer storefront,
      analytics dashboards, Elasticsearch.
- [x] Migration recovery/checksum verification (`forge migrate recover`
      fail-loud + force/recover/verify implemented).

## Later (vision)

- [ ] GraphQL + mobile API surface.
- [ ] Kubernetes deployment, service mesh, RabbitMQ/Kafka,
      microservices split.
- [ ] Multi-tenant, marketplace, subscriptions, affiliate,
      recommendation engine, social auth.
- [ ] Data migrations, squashing UI, migration testing framework,
      perf analysis, multi-DB, migration templates.
- [ ] Read-replicas, CDN, distributed tracing/metrics/alerting
      (today: local pooling/indexes/pagination only).
- [ ] Admin UI vitest re-attempt once runner restrictions are resolved
      (only if the 24x7 loop is ever revived; otherwise drop).
