---
sidebar_position: 8
description: Patterns for handling errors in forge handlers and APIs.
keywords:
  - forge error handling
  - api errors
  - http errors
image: /forge-social-card.svg
---

# Error Handling

Handle errors consistently so clients receive clear responses and logs remain
actionable.

## Handler-level errors

Return HTTP status codes that match the failure:

```go
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

## API responses

For API endpoints, prefer structured error responses and keep messages
developer-friendly while avoiding sensitive details.

## Related docs

- [REST API guide](/docs/guides/rest-api/)
- [Best practices](/docs/guides/best-practices/)
