---
sidebar_position: 2
description: Default middleware stack and utilities.
image: /social-card.png
---

# Middleware

Default middleware includes request ID, logging, recoverer, and timeouts.

## Default stack

- RequestID
- RealIP: replaces `r.RemoteAddr` with the client IP from `X-Forwarded-For`
  or `X-Real-IP`, but only when the direct peer is listed in
  `server.trusted_proxies`; headers from any other peer are ignored
- Recoverer
- Logger
- Timeout

## Additional middleware

- CORS
- Compress
- SecureHeaders
- StripSlashes, RedirectSlashes

## Next steps

- [Security](/docs/server/security/)
- [Health & Metrics](/docs/server/health/)
