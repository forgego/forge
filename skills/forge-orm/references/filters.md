# Filtering & Security (`filter/`)

Forge's `filter/` package provides a shared, serializable filter AST used by both the Admin UI and REST APIs.

## Security Model (`filter/security.go`, `filter/optimizer.go`)

- **Default deny**: Fields and lookups must be explicitly whitelisted before parsing user-supplied query parameters.
- **Cost & JOIN limits**: `filter/optimizer.go` and `filter/metrics.go` score query complexity and guard against unbounded relation traversal.

## Basic FilterSet Usage

Configure allowed filter fields on `ViewSetConfig.Filterable` or `admin.Config[T].ListFilter` / `Filters`, or parse query params with `filter.NewParser` against a `filter.FilterSet` whitelist.
