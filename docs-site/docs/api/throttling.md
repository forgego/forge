---
sidebar_position: 6
description: Rate limiting API requests with fixed-window throttles, shared database counters and HTTP 429 responses.
image: /social-card.png
---

# API Throttling & Rate Limiting

Throttles cap how many requests a client may make in a time window. They run
after authentication and permissions, before the view, and a request over the
limit is answered with `429 Too Many Requests`.

---

## Built-in throttles

Package `github.com/forgego/forge/api/throttling` provides two throttles:

| Constructor | Counts requests per | Bucket key |
| --- | --- | --- |
| `throttling.NewAnonRateThrottle(rate)` | client IP, for every request | `throttle_anon_<ip>` |
| `throttling.NewUserRateThrottle(rate)` | authenticated user ID, or client IP when there is no user | `throttle_user_<id or ip>` |

The client IP honors `X-Forwarded-For` and `X-Real-IP` only from peers listed
in `server.trusted_proxies`. Behind a reverse proxy, list it there, or every
client shares the proxy's address and its budget.

---

## Attaching throttles

Set them on one viewset, or as the default for every viewset whose
`Throttles` is `nil`:

```go
import (
    "github.com/forgego/forge/api"
    "github.com/forgego/forge/api/throttling"
)

// One viewset: both throttles must allow the request.
viewSet.Throttles = []throttling.Throttle{
    throttling.NewAnonRateThrottle("60/min"),
    throttling.NewUserRateThrottle("1000/hour"),
}

// Every viewset that has not set Throttles.
api.SetDefaultThrottles(throttling.NewUserRateThrottle("1000/hour"))
```

A non-nil empty slice, `viewSet.Throttles = []throttling.Throttle{}`, turns
throttling off for that viewset even when defaults are set.

---

## Rate syntax

A rate is `<number>/<period>`, where the period is `sec`, `min`, `hour` (or
`hr`) or `day`; longer spellings such as `second` or `minute` also work.

- `"10/sec"`: 10 requests per second
- `"60/min"`: 60 requests per minute
- `"1000/hour"`: 1,000 requests per hour
- `"10000/day"`: 10,000 requests per day

Windows are fixed: the count for a key starts at its first request and resets
when the window ends. A rate that does not parse makes every request through
that throttle fail with an error instead of being counted, so check rates in a
test.

---

## Where counts are kept

A throttle without a store of its own counts in the default store:

- With `server.stores: memory` (the library default), each process keeps its
  own counters, which reset on restart. With N instances, a client gets N
  times the budget.
- With `server.stores: database`, `server.NewServer` switches the default to
  the `forge_rate_limits` table, so every instance on the database shares one
  count. See [the deployment guide](/docs/deployment/) for setting it up.

If the database cannot be reached, or does not answer within two seconds, the
request is allowed and the error is logged: a database outage does not turn
into refused requests. The query is also stopped when the client disconnects.

The default store is process-wide, and the most recently created server's
`server.stores` setting decides it.

### A store of your own

`WithStore` gives one throttle its own store, which ignores `server.stores`:

```go
throttle := throttling.NewUserRateThrottle("100/min").WithStore(myStore)
```

A store implements `throttling.Store`:

```go
type Store interface {
    Allow(key string) (allowed bool, retryAfter time.Duration)
}
```

A store that does I/O should also implement `throttling.ContextStore`.
Throttles then call `AllowContext` with the request's context, so the check
stops when the client goes away. An error there allows the request and is
logged:

```go
type ContextStore interface {
    Store
    AllowContext(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error)
}
```

`*stores.RateLimiter` (from `stores.Stores.RateLimiter`) implements both.

---

## Rate limit response (`429`)

A throttled request gets a problem response with a `Retry-After` header
giving the seconds until the window ends:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/problem+json
Retry-After: 42

{
  "type": "https://api.example.com/problems/rate-limit-error",
  "title": "Rate Limit Exceeded",
  "status": 429,
  "detail": "Request was throttled",
  "instance": "/api/products/",
  "code": "RATE_LIMIT_EXCEEDED",
  "meta": {"retry_after_seconds": 42}
}
```

The `type` base URL is the `errors.problem_details.type_base_url` setting.

---

## Next steps

- **[Permissions](/docs/api/permissions/)**: role and scope checks.
- **[Security settings](/docs/server/security/)**: CSRF and cookie protection.
