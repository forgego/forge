---
sidebar_position: 6
description: Logging configuration.
image: /social-card.png
---

# Logging Settings

## Levels

- trace, debug, info, warn, error, fatal

## Formats

- json, text, console

## Outputs

- console, file
- remote: not implemented; configuring it makes logger construction fail with `NotImplemented`

## Production sampling

Configure sampling to limit log volume in production.

## Generated projects

The `main.go` written by `forge new` builds its logger with
`log.NewLogger(settings.App.Debug)`, so it does not read the `logging.*` keys.
Build the logger with `log.NewLoggerFromConfig` to use them.
