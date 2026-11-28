---
title: Error Settings
description: Problem details, request ID, error sanitization, and idempotency settings.
image: /social-card.png
---

# Error Settings

`ErrorSettings` (`config/logging_settings.go`, `config/settings.go`) configures RFC 7807 Problem Details, request ID propagation, and error sanitization.

## Problem Details (`errors.problem_details`)

| Key | Default | Description |
|---|---|---|
| `errors.problem_details.type_base_url` | `"https://api.example.com/problems"` | Base URI for problem `type` fields |
| `errors.problem_details.include_stack_trace` | `false` | Include stack trace in internal/debug payloads |
| `errors.problem_details.include_internal_details` | `false` | Include internal error details |

## Request ID (`errors.request_id`)

| Key | Default | Description |
|---|---|---|
| `errors.request_id.header_name` | `"X-Request-ID"` | Header read/written by `RequestIDMiddleware` |
| `errors.request_id.generate_if_missing` | `true` | Generate a UUID when the header is absent |
| `errors.request_id.include_in_response` | `true` | Echo the request ID on HTTP responses |

## Sanitization (`errors.sanitization`)

| Key | Default | Description |
|---|---|---|
| `errors.sanitization.hide_database_errors` | `true` | Redact raw SQL/driver errors from responses |
| `errors.sanitization.hide_stack_traces` | `true` | Loaded for compatibility; sanitized error messages always strip recognized stack traces |
| `errors.sanitization.redact_pii` | `true` | Redact common PII patterns in error messages |

## Idempotency (`errors.idempotency`) — Experimental

Idempotency key configuration is **experimental**: `api/errors` only provides an in-memory store (no automatic HTTP replay middleware is wired), and `NewDatabaseStore` returns `NotImplemented`.

| Key | Default | Description |
|---|---|---|
| `errors.idempotency.enabled` | `true` | Enable idempotency configuration |
| `errors.idempotency.header_name` | `"Idempotency-Key"` | Idempotency header name |
| `errors.idempotency.cache_ttl` | `3600` | TTL in seconds |
| `errors.idempotency.store_type` | `"memory"` | Only `"memory"` is implemented |
| `errors.idempotency.max_nesting_depth` | `10` | Max JSON nesting depth |
