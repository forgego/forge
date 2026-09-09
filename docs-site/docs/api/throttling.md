---
sidebar_position: 6
description: Request rate limiting, sliding window throttling, and HTTP 429 Too Many Requests handling.
image: /forge-social-card.svg
---

# API Throttling & Rate Limiting

Forge includes a robust rate-limiting subsystem to prevent brute-force attacks, protect database connections, and defend against denial-of-service abuse.

---

## Setting Up Throttling

Throttles can be attached to individual ViewSets or configured globally:

```go
package catalog

import (
    "github.com/forgego/forge/api"
)

func ConfigureThrottling(viewset *ProductViewSet) {
    // 60 requests per minute for anonymous clients
    anonThrottle := api.NewAnonRateThrottle("60/min")
    
    // 1000 requests per hour for authenticated users
    userThrottle := api.NewUserRateThrottle("1000/hour")
    
    viewset.Throttling = api.NewCompositeThrottle(anonThrottle, userThrottle)
}
```

---

## Rate Syntax

Rate limits are declared using the intuitive format `<number>/<period>`:
- `"10/sec"` — 10 requests per second
- `"60/min"` — 60 requests per minute
- `"1000/hour"` — 1,000 requests per hour
- `"10000/day"` — 10,000 requests per day

---

## Rate Limit Response (`HTTP 429`)

When a client breaches their permitted rate budget, Forge immediately halts execution and responds with HTTP 429:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 42

{
  "status": 429,
  "error": "Too Many Requests",
  "detail": "Request was throttled. Expected available in 42 seconds."
}
```

---

## Next Steps

- **[Permissions](/docs/api/permissions/)**: Role and scope verification.
- **[Security Settings](/docs/server/security/)**: Platform CSRF and cookie protection.

