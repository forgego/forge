---
sidebar_position: 7
description: "Page-number pagination for viewset list responses."
image: /social-card.png
---

# API Pagination

Viewset `list` responses are always paginated by page number. It is the only
pagination scheme: there is no limit/offset or cursor pagination, and it cannot
be switched off or replaced per viewset.

## Query parameters

| Parameter | Default | Notes |
| --- | --- | --- |
| `page` | `1` | 1-indexed. A missing, non-numeric or smaller than 1 value means page 1. |
| `page_size` | `api.Settings.PageSize` (20) | A missing, non-numeric or smaller than 1 value uses the default. Values above `api.Settings.MaxPageSize` (100) are capped. |

```http
GET /api/v1/products/?page=3&page_size=25
```

A page past the end answers 200 with no results, not an error. The
queryset is counted and then sliced with `Offset` and `Limit`, so an
`orm.Manager[T]` paginates in SQL (see [data access](/docs/api/data-access/)).

## Response format

```json
{
  "count": 1420,
  "next": "http://localhost:8000/api/v1/products/?page=4&page_size=25",
  "previous": "http://localhost:8000/api/v1/products/?page=2&page_size=25",
  "results": [ ... ]
}
```

- `count` is the total number of matching rows after filtering.
- `next` and `previous` are absolute URLs that keep the request's other query
  parameters (filters, `ordering`, `search`). They are omitted on the last and
  first page.

## Configuring page sizes

Page sizes are global API settings:

```go
settings := api.DefaultSettings()
settings.PageSize = 50     // default page_size
settings.MaxPageSize = 200 // upper bound for page_size
api.SetSettings(settings)
```

## Paginating custom handlers

Custom actions and hand-written handlers can use the same format:

```go
page, pageSize, offset := api.ParsePaginationParams(r, 20)
rows, total := loadRows(offset, pageSize) // your query
_ = api.SendPaginatedResponse(w, r, rows, total, page, pageSize)
```

`api.BuildPaginatedResponse` returns the envelope without writing it, and
`api.NewPagination(page, pageSize, total)` computes page counts and
next/previous page numbers.

## Next Steps

- **[Throttling](/docs/api/throttling/)**: Prevent denial-of-service abuse.
- **[Filters](/docs/filters/)**: Narrow list results with query parameters.
