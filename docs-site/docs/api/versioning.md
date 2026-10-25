---
sidebar_position: 8
description: Versioning strategies for APIs.
image: /social-card.png
---

# Versioning

API versioning supports URL path, query parameter, and header strategies.

## Strategies

- URLPathVersioning
- QueryParameterVersioning
- HeaderVersioning

## Evolution policy

APIs evolve according to strict backward compatibility rules:

- **Non-breaking (additive) changes**: Adding new endpoints, optional request parameters, or response fields is non-breaking. Clients must ignore unexpected or unrecognized fields.
- **Breaking changes**: Removing or renaming an existing field or endpoint, altering a field's data type, or changing an HTTP response status code constitutes a breaking change and requires a new API version.
- **Deprecations**: Deprecated endpoints and fields are announced through release notes and documentation ahead of removal in a subsequent major version. The `api/versioning` package does not currently emit `Deprecation` or `Sunset` HTTP headers; deprecations are communicated through published documentation and changelogs.
- For complete lifecycle commitments and support tiers, see [Versioning and Support](../contributing/versioning-and-support.md). <!-- doclinks:ignore -->

## Next steps

- [Renderers & Parsers](/docs/api/renderers-parsers/)
- [OpenAPI](/docs/api/openapi/)
