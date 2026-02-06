---
sidebar_position: 8
description: Handle errors consistently in forge apps.
keywords:
  - forge errors
  - error handling
image: /forge-social-card.svg
---

# Error Handling Guide

Return clear HTTP errors and log server failures.

## Example

```go
if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

## Next steps

- [Validation guide](/docs/guides/validation)
- [REST API guide](/docs/guides/rest-api)
