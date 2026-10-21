---
sidebar_position: 2
description: Default middleware stack and utilities.
image: /social-card.png
---

# Middleware

`NewServer` installs request ID handling, and logging and error handling when a logger is supplied. It also installs request size limiting, sessions, and CSRF when configured.

## Default stack

- RequestID
- ErrorHandler and Logger (when a logger is supplied)
- RequestSizeLimit (when `server.max_request_size` is positive)
- Session middleware (when `security.session_secret` is set)
- CSRF middleware (when `security.csrf_secret_key` is set)

The client IP resolver trusts forwarding headers only from peers in `server.trusted_proxies`. `Recoverer` and `Timeout` are available for explicit installation.

## Additional middleware

- CORS
- Compress
- SecureHeaders
- Recoverer, Timeout
- StripSlashes, RedirectSlashes

## Next steps

- [Security](/docs/server/security/)
- [Health & Metrics](/docs/server/health/)
