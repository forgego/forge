---
sidebar_position: 3
description: Server settings.
image: /social-card.png
---

# Server Settings

- Host, Port
- ReadTimeout, WriteTimeout
- StaticFilesPath
- HealthCheckPath
- MetricsEnabled, MetricsPath
- GracefulTimeout
- MaxRequestSize
- EnableProfiling
- InfoEndpoint (`server.info_endpoint`, default `false`): registers `/info`,
  which reports app name, version, environment, debug flag and uptime
- TrustedProxies (`server.trusted_proxies`, default empty): proxy IPs or CIDRs
  whose `X-Forwarded-For` / `X-Real-IP` headers are honored when resolving the
  client IP for rate limiting and the admin login lockout. From any other peer
  the TCP peer address is used and forwarding headers are ignored
- Framework stores: sessions, API throttling counters, admin API tokens, admin
  login lockout, saved views, and change history always live in the application
  database. Pass `server.WithDatabase(database)` to `server.NewServer` and run
  `forge migrate up`. There is no configuration setting or memory option. See
  [Deployment](/docs/deployment/#shared-state).
