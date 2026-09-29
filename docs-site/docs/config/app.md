---
sidebar_position: 2
description: Application settings.
image: /social-card.png
---

# App Settings

- Name
- Env
- Version
- Debug (`app.debug`, default `false`): allows the profiling routes when
  `server.enable_profiling` is also on, and selects the development logger
  for commands run through the `forge` CLI. It does not change a server's
  logger: the `main.go` from `forge new` builds it from `logging.level` and
  `logging.format` (see [logging](/docs/config/logging/)). Rejected at startup
  when `app.env` is `production`.
