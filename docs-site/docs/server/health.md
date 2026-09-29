---
sidebar_position: 4
description: Health checks, readiness, metrics, and profiling.
image: /social-card.png
---

# Health & Metrics

The server can expose health, readiness, liveness, metrics, and profiling endpoints.

## Endpoints

- `server.health_check_path` (default `/health`): runs every registered check; 503 if any fails
- `<health path>/ready`: the same checks, reported as readiness
- `<health path>/live`: always 200 while the process serves requests
- `/metrics` when `server.metrics_enabled` is true: reports uptime only
- `/info` when `server.info_endpoint` is true (default false): app name, version, environment, debug flag and uptime
- `/debug` profiling, only when both `server.enable_profiling` and `app.debug` are true

No check is registered by default, so `/health` and `/ready` return 200 even
when the database is unreachable. Register one with
`server.RegisterHealthCheckFunc`; the [deployment guide](/docs/deployment/)
shows a database check.

## Next steps

- [Server Overview](/docs/server/overview/)
- [Config Server](/docs/config/server/)
