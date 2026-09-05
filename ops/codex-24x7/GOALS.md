# Codex 24x7 Goals

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
