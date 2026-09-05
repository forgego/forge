---
name: forge-api
description: Build REST APIs with Forge's ViewSet framework. Use when creating serializers, viewsets, registering API routes, or configuring auth, permissions, pagination, and filtering.
---

# Forge API

## Overview
Use Forge's REST API framework to build CRUD endpoints with ViewSets and serializers.

## When to Use
- The task mentions serializers, viewsets, API routers, or REST endpoints.
- You need CRUD APIs for Forge models.
- The user asks about auth, permissions, filtering, ordering, or pagination for APIs.

## Quick Start
1. Create a serializer defining exposed fields.
2. Create a ViewSet bound to a QuerySet and model.
3. Register the ViewSet on an API router and mount it on the server router.

## Common Tasks
- Add filtering, ordering, and pagination defaults.
- Customize read-only fields and validation in serializers.
- Organize routes with versioned API base paths.

## Gotchas
- Keep serializer fields aligned with model fields after changes.
- Router base path affects all endpoint URLs.
- Re-run `forge generate` when model changes affect QuerySets or types.

## References
- [REST API concepts](references/rest-api.md)
- [First API walkthrough](references/first-api.md)
