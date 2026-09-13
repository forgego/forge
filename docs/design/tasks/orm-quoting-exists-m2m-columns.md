Your workspace is /home/hamid/Other/projects/foreit-wt/wave0-rest. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# ORM: quote identifiers in write SQL, cheap Exists, many-to-many through columns

Module root: /home/hamid/Other/projects/foreit-wt/wave0-rest/forge (run go commands from there).
Only edit files in forge/orm (and new _test.go files there). Other agents may edit other packages concurrently: ignore build errors outside forge/orm.

## Principles (mandatory)
- Tests FIRST, run them red, then fix. Smallest correct change; gofmt.
- Functions < 40 lines, early returns, errors wrapped with %w, no new package-level mutable state. Table-driven tests.
- Naming after behaviour, never after tickets.

## 1. Reserved table/column names break writes (verified)
forge/orm/manager_helpers.go builds `INSERT INTO %s (%s) VALUES (%s) RETURNING %s` (~197), `UPDATE %s SET %s WHERE %s = $%d` (~391), `DELETE FROM %s WHERE %s = $1` (~403), bulk insert (~478) with raw names, so a model named `order` or a column `user` produces invalid SQL. forge/orm already has `EscapeIdentifier` (used in orm/prefetch.go:183): use it for every table and column name in those helpers. Update the expected strings in forge/orm/manager_helpers_test.go accordingly.
Test: a table `order` with columns `user`, `group` on in-memory SQLite (see existing orm tests that open SQLite, e.g. write_column_names_test.go) — Create, Update and Delete succeed.

## 2. Exists runs a full COUNT (verified)
`BaseQuerySet.Exists` (forge/orm/queryset.go ~1480) calls `Count`. Replace it with a query that selects `1` with `LIMIT 1` using the same WHERE/JOIN building the queryset already uses for `Count` (read how Count builds its SQL and reuse that path; do not duplicate the WHERE builder). Test: Exists true/false on a filtered SQLite table, and the generated SQL (if the builder exposes it) contains `LIMIT 1` and not `COUNT`.

## 3. Many-to-many prefetch guesses through columns by trimming "s" (verified)
forge/orm/prefetch.go:179-180 uses `strings.TrimSuffix(table, "s") + "_id"`, so `categories` becomes `categorie_id`. Django's convention for an auto-created through table is `<source model name>_id` and `<target model name>_id`, using the lowercase model name, not the table name. Use the source and target schemas' model names converted to snake_case (find the model-name field on the schema struct and the snake-case helper already used in forge/orm or `github.com/iancoleman/strcase`, already a dependency). Keep `RelationInfo.Through` as the table. Test: a `Product` ↔ `Category` M2M with through table `product_categories(product_id, category_id)` and table names `products`/`categories` prefetches correctly on SQLite.

## Acceptance (from the module root)
    gofmt -l orm                                   # empty
    go vet ./orm/...
    staticcheck ./orm/...                          # no output
    go test -race ./orm/... ./admin/... -count=1

Report (max 25 lines): failing output before each fix, files changed, each command result.
