---
sidebar_position: 6
description: Request throttling and rate limits.
image: /forge-social-card.svg
---

# Throttling

Throttle classes limit request rates for anonymous and authenticated users.

## Example

```go
api.SetDefaultThrottles(
    throttling.NewAnonRateThrottle("100/hour", cache),
    throttling.NewUserRateThrottle("1000/day", cache),
)
```

## Classes

- AnonRateThrottle
- UserRateThrottle
- Throttle cache backends

## Next steps

- [Pagination](/docs/api/pagination/)
- [Versioning](/docs/api/versioning/)
