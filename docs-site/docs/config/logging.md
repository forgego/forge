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

## Configuration keys

```yaml
logging:
  level: info      # default level; also the level of outputs that set none
  format: json     # console selects the colored development encoder
  outputs:         # optional; without it, logs go to the console
    - type: console
      enabled: true
    - type: file
      enabled: true
      level: warn
      path: logs/app.log   # rotated at 100 MB, 30 days, 10 backups
```

`config.LoadSettings` reads these into `settings.Logging`, and
`log.NewLoggerFromSettings(settings.Logging)` builds the logger. An invalid
level, format or output type is an error when the logger is built. Without a
`logging` section the level is `info` and the format `json`.

An `outputs` value that is not a list of outputs, such as `outputs: file` or
an entry whose `enabled` is not a boolean, cannot be read. `LoadSettings`
then leaves `settings.Logging.Outputs` empty, so logs go to the console, and
the server logs a warning naming `logging.outputs` when it starts.
`config.Config.SettingsWarnings` returns these warnings for code that does
not start a `server.Server`.

## Generated projects

The `main.go` written by `forge new` builds its logger with
`log.NewLoggerFromSettings(settings.Logging)`, and its `config/config.yaml`
starts with `level: debug` and `format: console` for local development. Use
`format: json` in production.
