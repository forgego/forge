---
sidebar_position: 7
description: "Pagination schemes: PageNumber, LimitOffset, and high-performance CursorPagination."
image: /forge-social-card.svg
---

# API Pagination

Forge provides three distinct pagination strategies for list responses, configurable globally or on a per-ViewSet basis.

---

## 1. `LimitOffsetPagination` (Default)

Allows clients to request an arbitrary slice of records using `limit` and `offset` query parameters:

```http
GET /api/v1/products/?limit=20&offset=40
```

### JSON Response Format

```json
{
  "count": 1420,
  "next": "/api/v1/products/?limit=20&offset=60",
  "previous": "/api/v1/products/?limit=20&offset=20",
  "results": [ ... ]
}
```

---

## 2. `PageNumberPagination`

Traditional 1-indexed page number pagination:

```http
GET /api/v1/products/?page=3&page_size=25
```

```go
viewset.Pagination = api.NewPageNumberPagination(25, 100) // default 25, max 100
```

---

## 3. `CursorPagination` (High-Performance)

For massive tables (millions of rows) or real-time streaming feeds where `OFFSET` causes heavy database scanning, `CursorPagination` uses opaque base64 cursor tokens pointing to the last evaluated index value:

```http
GET /api/v1/products/?cursor=cD0yMDI2LTA5LTA4KzEwJTNBMDA=
```

```go
viewset.Pagination = api.NewCursorPagination("created_at", 50)
```

Benefits:
- Consistent query execution speed ($O(1)$ index lookup vs $O(N)$ full table scan).
- Immune to page-drift anomalies when new records are inserted between page requests.

---

## Next Steps

- **[Throttling](/docs/api/throttling/)**: Prevent denial-of-service abuse.
- **[OpenAPI 3.0](/docs/api/openapi/)**: Schema metadata for paginated responses.

