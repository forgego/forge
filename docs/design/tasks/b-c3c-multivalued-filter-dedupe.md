Your workspace is /home/hamid/Other/projects/foreit-wt/orm-dedupe. Always use absolute paths. Never read or edit outside it. Do NOT run any git command.

# C3c: filters across to-many relations must not duplicate parent rows

## Problem

`forge/orm/queryset.go` resolves relation paths in WHERE / ORDER BY through
`createJoinResolver` (line ~670), which appends one `LEFT JOIN` per hop to `pathJoins`.
When a hop is multi-valued (the reverse-FK branch at line ~720, where the target table holds
the FK back to the current table, or a `RelationManyToMany` relation), the join multiplies
parent rows:

    customers.Filter(orm.F("orders__total").Gt(100)).All(ctx)
    -- customer with 3 orders over 100 is returned 3 times

`Count` (line ~1339) already uses `COUNT(DISTINCT pk)`, so pagination totals disagree with the
returned rows.

## Decision (how Django admin 4.1+ and Laravel `whereHas` solve it)

Do not use `SELECT DISTINCT` (it breaks ORDER BY on joined columns in PostgreSQL and is costly).
Instead, when any WHERE filter crosses a multi-valued hop, move the filter into a primary-key
subquery:

    SELECT "customers".* FROM "customers" <select_related joins> <order-by path joins>
    WHERE "customers"."id" IN (
        SELECT "customers"."id" FROM "customers" <where path joins> WHERE <original where>
    )
    ORDER BY ... LIMIT ... OFFSET ...

To-one paths (forward FK / OneToOne) keep the current single-query SQL, byte-identical.
Ordering by a to-many path is out of scope (it may still repeat rows, same as Django); do not
change it.

## Implementation (queryset.go only, plus the test file)

1. Make the resolver report cardinality and use its own dedupe set:
   - Change `createJoinResolver(pathJoins *[]string)` to
     `createJoinResolver(pathJoins *[]string, seen map[string]bool, multiValued *bool)`.
   - Use `seen` instead of mutating `qs.joinMap` for the path-join hops (still honour the
     existing `i == 0 && qs.joinMap[rel.Name]` select_related reuse check by reading
     `qs.joinMap`, never writing it from the resolver).
   - Set `*multiValued = true` when a hop takes the reverse-FK branch, or when
     `rel.Type == RelationManyToMany`.

2. `buildSQL` (line ~514):
   - Build WHERE with resolver W (`whereJoins`, its own `seen`, `whereMulti`).
   - Build ORDER BY on the same builder after swapping in resolver O
     (`orderJoins`, its own `seen`, `orderMulti` ignored) via `builder.SetJoinResolver`.
   - If `!whereMulti`: produce exactly the SQL produced today (WHERE and ORDER joins
     merged, without duplicating a join that both used). Existing tests assert the SQL, so keep
     the output identical for these cases.
   - If `whereMulti`: outer query = select clause (qualified with the table, as when
     `hasPathJoins` is true) + FROM + `qs.joins` + `orderJoins` +
     `WHERE <table>.<pk> IN (SELECT <table>.<pk> FROM <table> <whereJoins> <original where clause>)`
     + ORDER BY + LIMIT/OFFSET. The pk column is `qs.schema.PrimaryKey`, falling back to `"id"`
     (same as Count). Placeholder args stay in the order the builder produced them (WHERE
     args first), then `rebindSQL` as today.

3. `Count` (line ~1339): when `whereMulti`, emit
   `SELECT COUNT(*) FROM <table> WHERE <table>.<pk> IN (SELECT <table>.<pk> FROM <table> <whereJoins> <where>)`.
   Otherwise keep the current behaviour (COUNT(DISTINCT pk) when there are to-one path joins,
   COUNT(*) otherwise).

4. Keep functions under ~60 lines: extract a helper such as
   `func (qs *BaseQuerySet[T]) pkSubquery(joins []string, where string) string`.

## Tests (`forge/orm/relation_path_join_test.go`, SQLite, follow the existing setup there)

Add table-driven or separate tests:
- A customer with 3 orders over 100 and one with a single order over 100:
  `Filter(F("orders__total").Gt(100)).All` returns 2 customers, each once; `Count` returns 2.
- Same filter plus `OrderBy("name")`, `Limit(1)`, `Offset(1)`: returns exactly the second
  customer by name (pagination is over distinct parents).
- To-many filter combined with a to-one ordering path (`OrderBy("company__name")`) runs without
  SQL error and returns distinct rows in company-name order.
- The existing to-one tests still pass unchanged (SQL byte-identical where they assert SQL).

## Acceptance (run from /home/hamid/Other/projects/foreit-wt/orm-dedupe/forge)

    gofmt -l orm            # prints nothing
    go vet ./orm
    go test ./orm -run 'Relation|Path|Join|Count' -count=1
    go test ./orm -count=1
    go test ./admin/... -count=1

Report the commands and results in under 30 lines. Do not touch any other file.
