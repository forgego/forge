# Schema capability probes

Copy `models.go` into a new app, then run strict model/API generation and
`forge makemigrations initial --auto --models ./app`. Inspect SQL rather than
assuming strict generation certifies every migration capability.

On 2026-10-20 this reproduced Decimal/JSON/UUID as DOUBLE PRECISION/BYTEA/TEXT,
missing DBIndex/UniqueTogether/implicit M2M SQL, and present literal Meta index
and UNIQUE constraint SQL. The omissions are known defects, not golden outputs
that future changes must retain.

Native auth collision: in a separate generated project run `forge auth`,
`forge generate --api --strict --models ./app --output ./app`, then build.
The auth scaffold and generated API both declare UserSerializer. Generate auth
models without `--api` to use the existing handwritten auth endpoints.
