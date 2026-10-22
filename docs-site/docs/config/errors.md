---
sidebar_position: 7
description: Error and problem detail settings.
image: /social-card.png
---

# Error Settings

## Problem details

- Type base URL
- Include stack trace
- Include internal details

## Request ID

- Header name
- Generate if missing
- Include in response

## Sanitization

- Hide database errors
- Sanitized error messages always strip recognized stack traces. The `errors.sanitization.hide_stack_traces` setting is loaded but does not currently change that behavior; the claim that it can show stack traces is removed.
- Redact PII

## Idempotency

- Enabled
- Header name
- Cache TTL
- Store type
- Max nesting depth
