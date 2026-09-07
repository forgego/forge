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

- [ ] Ecommerce unit/integration/E2E suites for non-category modules
      (models/services, API/DB integration, purchase/return/support flows).
- [ ] Authenticated integration coverage + object-specific 403 paths
      where SQLite-backed.
- [ ] TODO/FIXME burn-down across admin, ORM/schema/migrations,
      API/server reliability edges.
- [ ] Password-hash `BeforeUpdate` guard (`strings.HasPrefix(hash,"$2a$")`)
      to avoid double-hashing (`skills` hooks backlog).
- [ ] Analyzer corpus fixtures/goldens/benchmarks; RAG/evaluator ranking
      fixtures; harness-compatibility evidence (Claude/Codex/OpenCode/Zed/dmux).
- [ ] `make seed` references `scripts/seed.go`, which does not exist —
      recreate the seed script or drop the target.
- [ ] Ecommerce production hardening: rotate `session_secret`/`csrf_secret`,
      CORS/SSL, rate limits, backups, monitoring (`config.yaml:35-36`).
- [ ] Ecommerce: auth middleware, caching layer, structured logging,
      dashboard widgets, email templates, file uploads.
- [ ] Ecommerce: `Mark Delivered` must set `delivered_at`; required-field
      saves must return errors, not 500s.
- [ ] Decide: "pre-built binaries are coming soon" (old install docs) —
      implement releases or drop the promise.

## Next (framework features)

- [ ] REST auto-generation: ViewSets + Serializers + routes + OpenAPI UI
      generated from models (`docs/PRD.md` FR-API/OpenAPI).
- [ ] Missing CLI: `startapp`, `shell`, `test`, `collectstatic`,
      `createsuperuser`, `dbshell`, `check` (`docs/PRD.md` App.CLI).
- [ ] Caching: query/instance cache, Redis + in-memory backends.
- [ ] Query power: window functions, full-text search, raw SQL escape hatch.
- [ ] Background tasks (queues/workers).
- [ ] Observability: Prometheus metrics, OpenTelemetry health/tracing.
- [ ] Ecommerce: Stripe/PayPal + shipping-carrier APIs.
- [ ] Ecommerce: WebSocket notifications, customer storefront,
      analytics dashboards, Elasticsearch.
- [ ] Migration recovery/drift-detection/checksum/partial-rollback
      (fail-loud + force/recover intended).

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
