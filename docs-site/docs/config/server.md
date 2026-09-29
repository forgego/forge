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
- Stores (`server.stores`, `FORGE_SERVER_STORES`, default `memory`; `forge new`
  writes `database`): where sessions, API throttling counters and the admin's
  tokens, login lockout, saved views and change history live. `database`
  keeps them in the application database so they survive restarts and are
  shared by every instance; it needs `server.WithDatabase(database)` in
  `server.NewServer`, `adminSite.UseStores(ctx, settings.Server.Stores)`, and
  the framework store tables, which `forge migrate up` creates. `memory`
  keeps them in each process. Any other value is an error. See
  [Deployment](/docs/deployment/#shared-state)
