TASK C1: AndGroup / OrGroup / OrFilter build no groups at all (Go backend, filter package).

Working directory: /home/hamid/Other/projects/foreit-wt/filter-expression-correctness/forge
(git worktree on branch fix/filter-expression-correctness. Work ONLY there.)

Files you may modify — ONLY:
  filter/filterset.go
  NEW filter/filter_groups_test.go
Do NOT touch filter/cache.go (another agent owns it). Do NOT edit go.mod / go.sum.

=== THE BUG (High, wrong query results) ===
filter/filterset.go. FilterSet stores its predicate tree in `fs.ast`. Every terminal
method on FilterBuilder (lines ~186, 197, 208, 219, 230, 241 — Exact/Gt/…/IsNull) does:

    fb.fs.ast = combineWithAnd(fb.fs.ast, node)
    return fb.fs

AndGroup / OrGroup (lines ~246-267):

    builder := NewQueryBuilder[T](fs)
    fn(builder)
    if len(builder.nodes) > 0 { orNode := NewOrNode(builder.nodes...); fs.ast = combineWithAnd(fs.ast, orNode) }

QueryBuilder:
    func (qb *QueryBuilder[T]) Where(fieldPath string) *FilterBuilder[T] { return qb.fs.Where(fieldPath) }
    func (qb *QueryBuilder[T]) OrFilter(fieldPath string) *FilterBuilder[T] { return qb.Where(fieldPath) } // "for now"

So predicates written inside the callback go straight into the PARENT fs.ast as plain
AND terms, `builder.nodes` is always empty, and:
  fs.OrGroup(func(q){ q.Where("a").Exact(1); q.Where("b").Exact(2) })
produces  a=1 AND b=2   instead of   (a=1 OR b=2).
That silently returns the wrong rows.

=== THE FIX (do exactly this) ===
1. Read filter/filterset.go fully first, including the FilterBuilder struct and FilterSet.Where.

2. Add a field to FilterSet:
       sink func(*FilterNode) // when non-nil, new predicates go here instead of ast
   and a method:
       func (fs *FilterSet[T]) addNode(n *FilterNode) {
           if fs.sink != nil { fs.sink(n); return }
           fs.ast = combineWithAnd(fs.ast, n)
       }

3. Add a field to FilterBuilder:
       combine func(*FilterNode) // optional override used by QueryBuilder.OrFilter
   and a method:
       func (fb *FilterBuilder[T]) add(n *FilterNode) {
           if fb.combine != nil { fb.combine(n); return }
           fb.fs.addNode(n)
       }

4. Replace EVERY `fb.fs.ast = combineWithAnd(fb.fs.ast, node)` with `fb.add(node)`.
   Replace in AndGroup `fs.ast = combineWithAnd(fs.ast, andNode)` with `fs.addNode(andNode)`,
   and in OrGroup `fs.ast = combineWithAnd(fs.ast, orNode)` with `fs.addNode(orNode)`.
   (This makes nested groups inside a group callback land in the group, not the root.)
   After this, `grep -n "combineWithAnd(fb.fs.ast" filter/filterset.go` must print nothing.

5. NewQueryBuilder creates a CHILD FilterSet that shares configuration but captures nodes:
       qb := &QueryBuilder[T]{nodes: make([]*FilterNode, 0)}
       child := &FilterSet[T]{
           schema: fs.schema, filters: fs.filters, security: fs.security,
           optimizer: fs.optimizer, queryset: fs.queryset,
       }
       child.sink = func(n *FilterNode) { qb.nodes = append(qb.nodes, n) }
       qb.fs = child
       return qb
   (Copy every configuration field FilterSet has; do NOT copy ast.)

6. QueryBuilder.Where stays `return qb.fs.Where(fieldPath)` — it now targets the child.

7. QueryBuilder.OrFilter: the predicate must be OR-ed with the immediately preceding
   predicate in the group:
       fb := qb.fs.Where(fieldPath)
       fb.combine = func(n *FilterNode) {
           if len(qb.nodes) == 0 { qb.nodes = append(qb.nodes, n); return }
           last := qb.nodes[len(qb.nodes)-1]
           qb.nodes[len(qb.nodes)-1] = NewOrNode(last, n)
       }
       return fb
   Remove the "For now, just use Where" comment.

Check how NewAndNode/NewOrNode and the node type/children are represented (read the node
definitions in the filter package) so your tests can inspect the tree.

=== TESTS (filter/filter_groups_test.go) — write FIRST, see them fail ===
Use whatever test model/FilterSet constructor filterset_test.go already uses (read it).
Assert on the resulting AST structure (node type AND/OR and children), and ALSO on the
generated SQL/where-clause if the package exposes a function that renders it (look for one;
if it exists, assert the rendered string contains " OR ").
  1. OrGroup with two Where predicates -> root is an OR node with 2 leaf children.
  2. AndGroup with two predicates -> AND node with 2 children.
  3. Root predicate + OrGroup:  fs.Where("x").Exact(0); fs.OrGroup(a=1, b=2)
       -> AND(x=0, OR(a=1, b=2)).
  4. OrGroup with a single predicate -> OR node with 1 child (or the leaf — document
     whichever the constructor produces, but predicate must NOT leak into root as a bare AND
     sibling outside the group).
  6. OrFilter inside AndGroup: q.Where("a").Exact(1); q.OrFilter("b").Exact(2); q.Where("c").Exact(3)
       -> AND( OR(a=1,b=2), c=3 ).
  7. Empty group callback -> fs.ast unchanged (nil stays nil).
  8. Existing filterset_test.go tests still pass unchanged.

=== VERIFY (from working directory) ===
  gofmt -l filter                         (prints nothing)
  go vet ./filter/...
  go test -race ./filter/...
  go test ./...
  A prefetch fixture foreign-key failure in ./orm is KNOWN PRE-EXISTING on master:
  report it if you see it, do not fix it, do not investigate it.

=== HARD CONSTRAINTS ===
- NO git commands. The coordinator commits.
- Do not change exported signatures.
- Minimal diff; no unrelated refactors.
- Final report: files changed, tests added, exact tail of `go test -race ./filter/...`.
