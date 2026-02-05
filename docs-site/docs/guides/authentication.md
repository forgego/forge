---
sidebar_position: 5
description: Configure authentication and authorization in forge applications.
keywords:
  - forge authentication
  - auth backends
  - permissions
  - identity system
image: /forge-social-card.svg
---

# Authentication

forge provides authentication through the Identity System and pluggable
authentication backends. Use these building blocks to protect admin routes and
API endpoints.

## Core building blocks

- **Identity System** - User models, sessions, permissions, and auth services.
  See the [Identity System feature](/docs/features/identity-system/).
- **REST API authentication** - Apply authentication to API viewsets. See the
  [REST API guide](/docs/guides/rest-api/).
- **Security defaults** - CSRF and related protections. See the
  [Security guide](/docs/guides/security/).

## Next steps

1. Review the [Identity System feature](/docs/features/identity-system/) to
   understand available backends and services.
2. Follow the [REST API guide](/docs/guides/rest-api/) to wire authentication
   into API routes.
