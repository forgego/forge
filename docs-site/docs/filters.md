---
sidebar_position: 22
description: Advanced filtering with filtersets and AST parsing.
image: /forge-social-card.svg
---

# Filters

Use filtersets to build advanced filtering over ORM QuerySets.

## What you can do

- Define filtersets for models
- Parse query params into a filter AST
- Apply security rules and query optimization
- Use built-in filters and widgets

## FilterSet example

```go
fs, _ := filter.NewFilterSet[Post]()
qs := fs.WithQueryset(PostObjects).Where("title").Contains("forge")
```

## Next steps

- [ORM](/docs/orm/)
- [REST API](/docs/rest-api/)
