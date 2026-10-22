---
sidebar_position: 9
description: Content negotiation, parsers, and renderers.
image: /social-card.png
---

# Renderers & Parsers

Content negotiation selects a renderer and parser based on request headers.

## Parsers

- JSON
- Form
- Multipart
- XML

## Renderers

- JSON
- HTML
- XML
- CSV
- YAML

CSV struct headers use JSON field names: `json:"name"` renames a column, `json:"-"` omits it, and `omitempty` does not remove a column. Formula-like text cells receive a leading single quote; numeric Go values retain their numeric representation.

## Next steps

- [Errors](/docs/api/errors/)
- [OpenAPI](/docs/api/openapi/)
